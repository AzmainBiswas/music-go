package musictag

import (
	"errors"
	"io"
)

var ErrNoTagFound = errors.New("No tag found")

// read the tag from music file currently supported id3v1,2.{2,3,4}
func ReadFrom(r io.ReadSeeker) (Metadata, error) {
	b, err := readBytes(r, 10)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(-10, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	if string(b[0:3]) == "ID3" {
		return ReadID3v2Tags(r)
	}

	m, err := ReadID3v1Tags(r)
	if err != nil {
		if err == ErrNotID3V1 {
			return nil, ErrNoTagFound
		}
		return nil, err
	}

	return m, nil
}

type TagFormat string

// Supported tag formats
const (
	UnknownFormat TagFormat = "" // Unknown Format
	ID3v1         TagFormat = "ID3v1"
	ID3v2_2       TagFormat = "ID3V2.2"
	ID3v2_3       TagFormat = "ID3V2.3"
	ID3v2_4       TagFormat = "ID3V2.4"
)

type FileType string

// Supported file type
const (
	UnknownFileType FileType = ""
	MP3             FileType = "MP3"
)

// Music metadata interface
type Metadata interface {
	// GetTagFormat returns the metadata format to encode the data
	GetTagFormat() TagFormat

	// GetFileType returns the file type of the audio file
	GetFileType() FileType

	// GetTitle returns the title of the audio track
	GetTitle() string

	// GetArtist returns the track artist
	GetArtist() string

	// GetAlbum returns the album name of the track
	GetAlbum() string

	// GetAlbumArtist returns the album artist name of the track.
	GetAlbumArtist() string

	// GetYear returns the year of the track
	GetYear() int

	// GetGenre returns the genre of the track
	GetGenre() string

	// returns album art of the track
	GetAlbumArt() *Picture
}
