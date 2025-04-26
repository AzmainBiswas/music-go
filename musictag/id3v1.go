package musictag

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

type ID3v1Metadata map[string]any

func (ID3v1Metadata) GetTagFormat() TagFormat  { return ID3v1 }
func (ID3v1Metadata) GetFileType() FileType    { return MP3 }
func (m ID3v1Metadata) GetTitle() string       { return m["title"].(string) }
func (m ID3v1Metadata) GetArtist() string      { return m["artist"].(string) }
func (m ID3v1Metadata) GetAlbum() string       { return m["album"].(string) }
func (m ID3v1Metadata) GetAlbumArtist() string { return "" }
func (m ID3v1Metadata) GetGenre() string       { return m["genre"].(string) }
func (m ID3v1Metadata) GetComment() string     { return m["comment"].(string) }
func (m ID3v1Metadata) GetAlbumArt() *Picture  { return nil }

func (m ID3v1Metadata) GetYear() int {
	year := m["year"].(string)
	n, err := strconv.Atoi(year)
	if err != nil {
		return 0
	}
	return n
}

//TODOOO: AlbumArt implement

var ErrNotID3V1 = errors.New("Invalid ID3v1 tag")

func ReadID3v1Tags(filePointer io.ReadSeeker) (Metadata, error) {
	_, err := filePointer.Seek(-128, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	if tag, err := readString(filePointer, 3); err != nil {
		return nil, err
	} else if tag != "TAG" {
		return nil, ErrNotID3V1
	}

	title, err := readString(filePointer, 30)
	if err != nil {
		return nil, err
	}

	artist, err := readString(filePointer, 30)
	if err != nil {
		return nil, err
	}

	album, err := readString(filePointer, 30)
	if err != nil {
		return nil, err
	}

	year, err := readString(filePointer, 4)
	if err != nil {
		return nil, err
	}
	year = trimString(year)

	//TODO: modify comment if needed
	comment, err := readString(filePointer, 30)
	if err != nil {
		return nil, err
	}

	var genre string
	genreId, err := readBytes(filePointer, 1)
	if err != nil {
		return nil, err
	}

	if int(genreId[0]) < len(id3Genres) {
		genre = id3Genres[int(genreId[0])]
	}

	var id3v1Metadata = make(map[string]any)
	id3v1Metadata["title"] = trimString(title)
	id3v1Metadata["artist"] = trimString(artist)
	id3v1Metadata["album"] = trimString(album)
	id3v1Metadata["genre"] = genre
	id3v1Metadata["year"] = trimString(year)
	id3v1Metadata["comment"] = trimString(comment)

	return ID3v1Metadata(id3v1Metadata), nil
}

func trimString(input string) string {
	return strings.TrimSpace(strings.Trim(input, "\x00"))
}
