package musictag

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type frameNames map[string][2]string

func (f frameNames) Name(s string, fm TagFormat) string {
	l, ok := f[s]
	if !ok {
		return ""
	}

	switch fm {
	case ID3v2_2:
		return l[0]
	case ID3v2_3:
		return l[1]
	case ID3v2_4:
		if s == "year" {
			return "TDRC"
		}
		return l[1]
	}
	return ""
}

var frames = frameNames(map[string][2]string{
	"title":        {"TT2", "TIT2"},
	"artist":       {"TP1", "TPE1"},
	"album":        {"TAL", "TALB"},
	"album_artist": {"TP2", "TPE2"},
	"composer":     {"TCM", "TCOM"},
	"year":         {"TYE", "TYER"},
	"track":        {"TRK", "TRCK"},
	"genre":        {"TCO", "TCON"},
	"picture":      {"PIC", "APIC"},
	"comment":      {"COM", "COMM"},
})

// metadataID3v2 is the implementation of Metadata used for ID3v2 tags.
type ID3v2Metadata struct {
	header *ID3v2Header
	frames map[string]any
}

func (m ID3v2Metadata) getString(k string) string {
	v, ok := m.frames[k]
	if !ok {
		return ""
	}
	return v.(string)
}

func (m ID3v2Metadata) GetTagFormat() TagFormat { return m.header.Version }
func (ID3v2Metadata) GetFileType() FileType     { return MP3 }
func (m ID3v2Metadata) GetTitle() string        { return m.getString(frames.Name("title", m.GetTagFormat())) }
func (m ID3v2Metadata) GetArtist() string {
	return m.getString(frames.Name("artist", m.GetTagFormat()))
}

func (m ID3v2Metadata) GetAlbum() string {
	return m.getString(frames.Name("album", m.GetTagFormat()))
}

func (m ID3v2Metadata) GetAlbumArtist() string {
	return m.getString(frames.Name("album_artist", m.GetTagFormat()))
}

func (m ID3v2Metadata) GetGenre() string {
	return id3v2genre(m.getString(frames.Name("genre", m.GetTagFormat())))
}

func (m ID3v2Metadata) GetComment() string {
	return m.getString(frames.Name("comment", m.GetTagFormat()))
}

func (m ID3v2Metadata) GetAlbumArt() *Picture {
	v, ok := m.frames[frames.Name("picture", m.GetTagFormat())]
	if !ok {
		return nil
	}
	return v.(*Picture)
}

func (m ID3v2Metadata) GetYear() int {
	stringYear := m.getString(frames.Name("year", m.GetTagFormat()))

	if year, err := strconv.Atoi(stringYear); err == nil {
		return year
	}

	date, err := time.Parse(time.DateOnly, stringYear)
	if err != nil {
		return 0
	}

	return date.Year()
}

var id3v2genreRe = regexp.MustCompile(`(.*[^(]|.* |^)\(([0-9]+)\) *(.*)$`)

// id3v2genre parse a id3v2 genre tag and expand the numeric genres
func id3v2genre(genre string) string {
	c := true
	for c {
		orig := genre
		if match := id3v2genreRe.FindStringSubmatch(genre); len(match) > 0 {
			if genreID, err := strconv.Atoi(match[2]); err == nil {
				if genreID < len(id3Genres) {
					genre = id3Genres[genreID]
					if match[1] != "" {
						genre = strings.TrimSpace(match[1]) + " " + genre
					}
					if match[3] != "" {
						genre = genre + " " + match[3]
					}
				}
			}
		}
		c = (orig != genre)
	}
	return strings.Replace(genre, "((", "(", -1)
}
