package musictag

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

func ReadID3v2Tags(r io.ReadSeeker) (Metadata, error) {
	h, offset, err := readID3v2Header(r)
	if err != nil {
		return nil, err
	}

	f, err := readID3v2Frames(r, offset, h)
	if err != nil {
		return nil, err
	}
	return ID3v2Metadata{header: h, frames: f}, nil
}

// Represent id3v2 header tag usualy contents in first 10bytes
type ID3v2Header struct {
	Version           TagFormat
	Unsynchronisation bool
	ExtendedHeader    bool
	Experimental      bool
	Size              uint
}

// reads id3v2 header from a io.Reader
// offset is number of bytes of header that was read
func readID3v2Header(r io.Reader) (header *ID3v2Header, offset uint, err error) {
	offset = 10
	b, err := readBytes(r, 10)
	if err != nil {
		return nil, 0, fmt.Errorf("expected to read 10 bytes (ID3v2Header): %v", err)
	}

	if string(b[0:3]) != "ID3" {
		return nil, 0, fmt.Errorf("expected to read \"ID3\"")
	}

	b = b[3:]
	var format TagFormat
	switch uint(b[0]) {
	case 2:
		format = ID3v2_2
	case 3:
		format = ID3v2_3
	case 4:
		format = ID3v2_4
	default:
		return nil, offset, fmt.Errorf("ID3 version: %v, expected: 2, 3 and 4", uint(b[0]))
	}

	//NOTE: for now we are ignoring the minor version b[1]
	header = &ID3v2Header{
		Version:           format,
		Unsynchronisation: getBit(b[2], 7),
		ExtendedHeader:    getBit(b[2], 5),
		Experimental:      getBit(b[2], 5),
		Size:              uint(get7BitChunkedInt(b[3:7])),
	}

	// If extended header exist increment the offset accordingly
	if header.ExtendedHeader {
		switch format {
		case ID3v2_3:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, offset, fmt.Errorf("expected to read 4 bytes (ID3v23 extended header len): %v", err)
			}
			// skip header, size is excluding len bytes
			extendedHeaderSize := uint(getInt(b))
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v23 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize

		case ID3v2_4:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v24 extended header len): %v", err)
			}
			// skip header, size is synchsafe int including len bytes
			extendedHeaderSize := uint(get7BitChunkedInt(b)) - 4
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v24 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize
		default:
			// nop, only 2.3 and 2.4 should have extended header
		}
	}

	return header, offset, nil
}

// id3v2FrameFlags is a type which represents the flags which can be set on an ID3v2 frame.
type id3v2FrameFlags struct {
	// Message (ID3 2.3.0 and 2.4.0)
	TagAlterPreservation  bool
	FileAlterPreservation bool
	ReadOnly              bool

	// Format (ID3 2.3.0 and 2.4.0)
	Compression   bool
	Encryption    bool
	GroupIdentity bool
	// ID3 2.4.0 only (see http://id3.org/id3v2.4.0-structure sec 4.1)
	Unsynchronisation   bool
	DataLengthIndicator bool
}

func readID3v23FrameFlags(r io.Reader) (*id3v2FrameFlags, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return nil, err
	}

	msg := b[0]
	fmt := b[1]

	return &id3v2FrameFlags{
		TagAlterPreservation:  getBit(msg, 7),
		FileAlterPreservation: getBit(msg, 6),
		ReadOnly:              getBit(msg, 5),
		Compression:           getBit(fmt, 7),
		Encryption:            getBit(fmt, 6),
		GroupIdentity:         getBit(fmt, 5),
	}, nil
}

func readID3v24FrameFlags(r io.Reader) (*id3v2FrameFlags, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return nil, err
	}

	msg := b[0]
	fmt := b[1]

	return &id3v2FrameFlags{
		TagAlterPreservation:  getBit(msg, 6),
		FileAlterPreservation: getBit(msg, 5),
		ReadOnly:              getBit(msg, 4),
		GroupIdentity:         getBit(fmt, 6),
		Compression:           getBit(fmt, 3),
		Encryption:            getBit(fmt, 2),
		Unsynchronisation:     getBit(fmt, 1),
		DataLengthIndicator:   getBit(fmt, 0),
	}, nil

}

func readID3v2_2FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 3)
	if err != nil {
		return
	}
	size, err = readUint(r, 3)
	if err != nil {
		return
	}
	headerSize = 6
	return
}

func readID3v2_3FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 4)
	if err != nil {
		return
	}
	size, err = readUint(r, 4)
	if err != nil {
		return
	}
	headerSize = 8
	return
}

func readID3v2_4FrameHeader(r io.Reader) (name string, size uint, headerSize uint, err error) {
	name, err = readString(r, 4)
	if err != nil {
		return
	}
	size, err = read7BitChunkedUint(r, 4)
	if err != nil {
		return
	}
	headerSize = 8
	return
}

// readID3v2Frames reads ID3v2 frames from the given reader using the ID3v2Header.
func readID3v2Frames(r io.Reader, offset uint, h *ID3v2Header) (map[string]any, error) {
	result := make(map[string]any)

	for offset < h.Size {
		var err error
		var name string
		var size, headerSize uint
		var flags *id3v2FrameFlags

		switch h.Version {
		case ID3v2_2:
			name, size, headerSize, err = readID3v2_2FrameHeader(r)

		case ID3v2_3:
			name, size, headerSize, err = readID3v2_3FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v23FrameFlags(r)
			headerSize += 2

		case ID3v2_4:
			name, size, headerSize, err = readID3v2_4FrameHeader(r)
			if err != nil {
				return nil, err
			}
			flags, err = readID3v24FrameFlags(r)
			headerSize += 2
		}

		if err != nil {
			return nil, err
		}

		// FIXME: Do we still need this?
		// if size=0, we certainly are in a padding zone. ignore the rest of
		// the tags
		if size == 0 {
			break
		}

		offset += headerSize + size

		// Avoid corrupted padding (see http://id3.org/Compliance%20Issues).
		if !validID3Frame(h.Version, name) && offset > h.Size {
			break
		}

		if flags != nil {
			if flags.Compression {
				switch h.Version {
				case ID3v2_3:
					// No data length indicator defined.
					if _, err := read7BitChunkedUint(r, 4); err != nil { // read 4
						return nil, err
					}
					size -= 4

				case ID3v2_4:
					// Must have a data length indicator (to give the size) if compression is enabled.
					if !flags.DataLengthIndicator {
						return nil, errors.New("compression without data length indicator")
					}

				default:
					return nil, fmt.Errorf("unsupported compression flag used in %v", h.Version)
				}
			}

			if flags.DataLengthIndicator {
				if h.Version == ID3v2_3 {
					return nil, fmt.Errorf("data length indicator set but not defined for %v", ID3v2_3)
				}

				size, err = read7BitChunkedUint(r, 4)
				if err != nil { // read 4
					return nil, err
				}
			}

			if flags.Encryption {
				_, err = readBytes(r, 1) // read 1 byte of encryption method
				if err != nil {
					return nil, err
				}
				size--
			}
		}

		b, err := readBytes(r, size)
		if err != nil {
			return nil, err
		}

		// There can be multiple tag with the same name. Append a number to the
		// name if there is more than one.
		rawName := name
		if _, ok := result[rawName]; ok {
			for i := 0; ok; i++ {
				rawName = name + "_" + strconv.Itoa(i)
				_, ok = result[rawName]
			}
		}

		switch {
		case name[0] == 'T':
			txt, err := readTFrame(b)
			if err != nil {
				return nil, err
			}
			result[rawName] = txt

		case name == "COMM" || name == "COM" || name == "USLT" || name == "ULT":
			t, err := readTextWithDescrFrame(b, true, true) // both lang and enc
			if err != nil {
				return nil, fmt.Errorf("could not read %q (%q): %v", name, rawName, err)
			}
			result[rawName] = t

		case name == "APIC":
			p, err := readAPICFrame(b)
			if err != nil {
				return nil, err
			}
			// only insert cover art
			if p.Type == pictureTypes[0x03] {
				result[rawName] = p
			}

		case name == "PIC":
			p, err := readPICFrame(b)
			if err != nil {
				return nil, err
			}

			// only insert cover art
			if p.Type == pictureTypes[0x03] {
				result[rawName] = p
			}

		default:
			continue
		}
	}
	return result, nil
}
