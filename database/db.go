package database

import (
	"database/sql"
	"errors"
	"fmt"
	"music-go/musictag"
	"music-go/utils"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

var artistSpLitter = regexp.MustCompile(`\s*(?:/|&|,)\s*`)

type DataBase struct {
	config   utils.Config
	DB       *sql.DB
	Location string
	logger   utils.CLogger
}

// open connection with the given database name.
// close the database after use.
func OpenConnection(config utils.Config, logger utils.CLogger) (*DataBase, error) {
	d := &DataBase{
		config:   config,
		Location: fmt.Sprintf("file:%s", path.Join(config.Database.Path, "music.db")),
		logger:   logger,
	}

	var err error
	d.DB, err = sql.Open("libsql", d.Location)
	if err != nil {
		return nil, err
	}

	err = d.DB.Ping()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DataBase) Close() error {
	if err := d.DB.Close(); err != nil {
		return err
	}

	return nil
}

func (d *DataBase) ReConnect() error {
	var err error
	d.DB, err = sql.Open("libsql", d.Location)
	if err != nil {
		return err
	}

	err = d.DB.Ping()
	if err != nil {
		return nil
	}

	return nil
}

// creat musics table to database
func (d *DataBase) CreatMusicsTable() error {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return err
		}
	}

	querys := []string{
		`CREATE TABLE IF NOT EXISTS musics (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            artist TEXT NOT NULL DEFAULT 'Unknown',
            album TEXT NOT NULL DEFAULT 'Unknown',
            album_artist TEXT NOT NULL DEFAULT 'Unknown',
            year INT NOT NULL DEFAULT 0,
            genre TEXT DEFAULT 'Unknown',
            music_location TEXT NOT NULL UNIQUE,
            UNIQUE(title, artist, album)
        );`,
		`CREATE TABLE IF NOT EXISTS artists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);`,
		`CREATE TABLE IF NOT EXISTS music_artists (
			music_id INTEGER NOT NULL,
			artist_id INTEGER NOT NULL,
			PRIMARY KEY (music_id, artist_id),
			FOREIGN KEY (music_id) REFERENCES musics(id) ON DELETE CASCADE,
			FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range querys {
		_, err = d.DB.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func defaultIfEmptyString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// extract details from music
func (d *DataBase) extractMusicTag(musicPath string) (map[string]any, error) {
	file, err := os.Open(musicPath)
	if err != nil {
		d.logger.Printf("ERROR: %s -> %s\n", err.Error(), musicPath)
		return nil, err
	}
	defer file.Close()

	tag, err := musictag.ReadFrom(file)
	if err != nil {
		d.logger.Printf("ERROR: failed to read the tag from %s: %v", musicPath, err)
		return nil, err
	}
	var musicDetails = map[string]any{
		"title":       defaultIfEmptyString(tag.GetTitle(), filepath.Base(musicPath)),
		"album":       defaultIfEmptyString(tag.GetAlbum(), "Unknown"),
		"artistRaw":   defaultIfEmptyString(tag.GetArtist(), "Unknown"),
		"albumArtist": defaultIfEmptyString(tag.GetAlbumArtist(), "Unknown"),
		"year":        tag.GetYear(),
		"genre":       defaultIfEmptyString(tag.GetGenre(), "Unknown"),
	}

	return musicDetails, nil
}

// can be *sql.db or *sql.Tx
type Queryer interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

// insert or get the artist id from database
func (d *DataBase) insertOrGetArtistID(db Queryer, artist string) (int64, error) {
	var artistID int64
	err := db.QueryRow(`SELECT id FROM artists WHERE name = ?`, artist).Scan(&artistID)

	if err == sql.ErrNoRows {
		result, err := db.Exec(`INSERT OR IGNORE INTO artists (name) VALUES (?)`, artist)
		if err != nil {
			d.logger.Println("ERROR: unable to insert artist:", err)
			return 0, nil
		}

		artistID, err = result.LastInsertId()
		if err != nil {
			d.logger.Println("ERROR: getting artist ID:", err)
			return 0, nil
		}
	} else if err != nil {
		d.logger.Printf("ERROR: quary failed for artist: %s : %v", artist, err)
		return 0, nil
	}

	return artistID, nil
}

// Read and store to database single music
func (d *DataBase) PushSingleMusicsToTable(musicPath string) error {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return err
		}
	}

	spLitter := regexp.MustCompile(`\s*(?:/|&|,)\s*`)

	tag, err := d.extractMusicTag(musicPath)
	if err != nil {
		d.logger.Printf("ERROR: failed to read the tag from %s: %v", musicPath, err)
		return err
	}

	result, err := d.DB.Exec("INSERT INTO musics(title, artist, album, album_artist, year, genre, music_location) VALUES( ?, ?, ?, ?, ?, ?, ? )", tag["title"], tag["artistRaw"], tag["album"], tag["albumArtist"], tag["year"], tag["genre"], musicPath)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			d.logger.Printf("INFO: music \"%s\" alrady exists :)\n", tag["title"].(string))
			return nil
		}
		d.logger.Printf("ERROR: unable to insert music: %v", err)
		return err
	}

	musicID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if musicID == 0 {
		return errors.New("Music may be exit or some other error have to check")
	}

	artistNames := spLitter.Split(tag["artistRaw"].(string), -1)
	for _, artist := range artistNames {
		artist = strings.TrimSpace(artist)
		if artist == "" {
			continue
		}

		artistID, err := d.insertOrGetArtistID(d.DB, artist)
		if err != nil || artistID == 0 {
			continue
		}

		_, err = d.DB.Exec(`INSERT OR IGNORE INTO music_artists (music_id, artist_id) VALUES (?, ?)`, musicID, artistID)
		if err != nil {
			d.logger.Println("ERROR: insert into music_artists:", err)
			return err
		}
	}

	return nil
}

// Read and store to database list of musics
func (d *DataBase) PushMusicsTOmusicsTable(musicPaths []string) error {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return err
		}
	}

	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	//TODOO: handel propery
	stmt, err := tx.Prepare("INSERT INTO musics(title, artist, album, album_artist, year, genre, music_location) VALUES( ?, ?, ?, ?, ?, ?, ? )")
	if err != nil {
		return err
	}

	defer func() {
		if cerr := stmt.Close(); cerr != nil {
			d.logger.Println(cerr)
		}
	}()

	spLitter := regexp.MustCompile(`\s*(?:/|&|,)\s*`)
	for _, mPath := range musicPaths {
		tag, err := d.extractMusicTag(mPath)
		if err != nil {
			d.logger.Printf("ERROR: failed to read the tag from %s: %v", mPath, err)
			continue
		}

		//TODOO: handel error
		result, err := stmt.Exec(tag["title"], tag["artistRaw"], tag["album"], tag["albumArtist"], tag["year"].(int), tag["genre"], mPath)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				d.logger.Printf("INFO: music \"%s\" alrady exists :)\n", tag["title"].(string))
				continue
			}
			d.logger.Printf("ERROR: unable to insert music: %v", err)
			continue
		}

		// Get the inserted music ID
		musicID, err := result.LastInsertId()

		if err != nil {
			d.logger.Printf("ERROR: %s", err.Error())
			continue
		}

		if musicID == 0 {
			continue
		}

		// Normalize artist(s) immediately
		artistNames := spLitter.Split(tag["artistRaw"].(string), -1)

		for _, artist := range artistNames {
			artist = strings.TrimSpace(artist)
			if artist == "" {
				continue
			}

			artistID, err := d.insertOrGetArtistID(d.DB, artist)
			if err != nil || artistID == 0 {
				continue
			}

			_, err = tx.Exec(`INSERT OR IGNORE INTO music_artists (music_id, artist_id) VALUES (?, ?)`, musicID, artistID)
			if err != nil {
				d.logger.Println("ERROR: insert into music_artists:", err)
			}
		}
	}

	//TODO: handel error properly
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// music structure
type Music struct {
	Id          int64
	Title       string
	Artists     []string
	Album       string
	AlbumArtist string
	Genre       string
	Year        int
	Path        string
}

// random music quary
// SELECT * FROM musics ORDER BY RANDOM() LIMIT 1;
func (d *DataBase) GetRandomMusic() (*Music, error) {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return nil, err
		}
	}
	var m = new(Music)
	var artistRaw string
	err = d.DB.QueryRow("SELECT * FROM musics ORDER BY RANDOM() LIMIT 1").Scan(&m.Id, &m.Title, &artistRaw, &m.Album, &m.AlbumArtist, &m.Year, &m.Genre, &m.Path)
	if err != nil {
		return nil, err
	}
	m.Artists = artistSpLitter.Split(artistRaw, -1)
	return m, nil
}

func (d *DataBase) GetMusicBYID(songId int64) (*Music, error) {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return nil, err
		}
	}
	var m = new(Music)
	var artistRaw string
	err = d.DB.QueryRow("SELECT * FROM musics WHERE id = ?", songId).Scan(&m.Id, &m.Title, &artistRaw, &m.Album, &m.AlbumArtist, &m.Year, &m.Genre, &m.Path)
	if err != nil {
		return nil, err
	}
	m.Artists = artistSpLitter.Split(artistRaw, -1)
	return m, nil
}

func (d *DataBase) GetAllMusics() ([]Music, error) {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return nil, err
		}
	}

	songs := make([]Music, 0)
	rows, err := d.DB.Query(`SELECT * FROM musics ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Music
		var artistRaw string
		err = rows.Scan(&m.Id, &m.Title, &artistRaw, &m.Album, &m.AlbumArtist, &m.Year, &m.Genre, &m.Path)
		if err != nil {
			d.logger.Printf("ERROR: could not scan row: %v\n", err)
			continue
		}
		m.Artists = artistSpLitter.Split(artistRaw, -1)
		songs = append(songs, m)
	}

	if len(songs) == 0 {
		return nil, err
	}

	return songs, nil
}

func (d *DataBase) GetAllMusicsByArtistID(artistID int64) ([]Music, error) {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return nil, err
		}
	}

	songs := make([]Music, 0)
	query := `
	SELECT m.*
	FROM musics m
	JOIN music_artists ma ON m.id = ma.music_id
	JOIN artists a ON ma.artist_id = a.id
	WHERE a.id = ?
	ORDER BY m.id ASC`

	rows, err := d.DB.Query(query, artistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Music
		var rawArtists string
		err = rows.Scan(&m.Id, &m.Title, &rawArtists, &m.Album, &m.AlbumArtist, &m.Year, &m.Genre, &m.Path)
		if err != nil {
			d.logger.Printf("ERROR: could not scan row: %v\n", err)
			continue
		}
		m.Artists = artistSpLitter.Split(rawArtists, -1)
		songs = append(songs, m)
	}

	if len(songs) == 0 {
		return nil, err
	}

	return songs, nil
}

func (d *DataBase) GetMusicsByAlbumName(albumName string) ([]Music, error) {
	err := d.DB.Ping()
	if err != nil {
		if err := d.ReConnect(); err != nil {
			return nil, err
		}
	}

	songs := make([]Music, 0)
	query := `SELECT * FROM musics WHERE album = ? ORDER BY id ASC`

	rows, err := d.DB.Query(query, albumName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Music
		var rawArtists string
		err = rows.Scan(&m.Id, &m.Title, &rawArtists, &m.Album, &m.AlbumArtist, &m.Year, &m.Genre, &m.Path)
		if err != nil {
			d.logger.Printf("ERROR: could not scan row: %v\n", err)
			continue
		}
		m.Artists = artistSpLitter.Split(rawArtists, -1)
		songs = append(songs, m)
	}

	if len(songs) == 0 {
		return nil, err
	}

	return songs, err
}

// album struct
type Album struct {
	Name       string
	Artist     string
	SongsCount int
}

// extract all the albums
func (d *DataBase) GetAllAlbums() ([]Album, error) {
	if err := d.DB.Ping(); err != nil {
		err = d.ReConnect()
		if err != nil {
			return nil, err
		}
	}

	var albums = make([]Album, 0)
	rows, err := d.DB.Query(`
		SELECT album, album_artist, COUNT(*) as songs_count
		FROM musics
		GROUP BY album
		ORDER BY songs_count DESC, album`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a Album
		err = rows.Scan(&a.Name, &a.Artist, &a.SongsCount)
		if err != nil {
			d.logger.Printf("ERROR: could not scan row: %v\n", err)
			continue
		}

		albums = append(albums, a)
	}

	if len(albums) == 0 {
		return nil, err
	}

	return albums, nil
}

type Artist struct {
	ID         int64
	Name       string
	SongsCount int
}

func (d *DataBase) GetAllArtists() ([]Artist, error) {
	if err := d.DB.Ping(); err != nil {
		err = d.ReConnect()
		if err != nil {
			return nil, err
		}
	}

	artists := make([]Artist, 0)
	query := `
SELECT
	a.id AS id,
	a.name AS artist_name,
	COUNT(m.id) AS song_count
FROM artists a
LEFT JOIN music_artists ma ON a.id = ma.artist_id
LEFT JOIN musics m ON ma.music_id = m.id
GROUP BY a.name
ORDER BY song_count DESC, artist_name`

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist Artist
		err = rows.Scan(&artist.ID, &artist.Name, &artist.SongsCount)
		if err != nil {
			d.logger.Printf("ERROR: could not scan row: %v\n", err)
			continue
		}

		artists = append(artists, artist)
	}

	if len(artists) == 0 {
		return nil, err
	}

	return artists, nil
}
