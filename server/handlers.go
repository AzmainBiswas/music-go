package server

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"music-go/database"
	"music-go/musictag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//TODOOOO: add request check to all handeler

func (s *httpServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	// TODO: for final product move it to Serve()
	// loading all templates
	err := s.LoadTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	err = s.indexTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongs(w http.ResponseWriter, r *http.Request) {
	musics, err := s.db.GetAllMusics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	payload := struct {
		Songs []database.Music
	}{
		Songs: musics,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "musics", payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) handleAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := s.db.GetAllAlbums()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	payload := struct {
		Albums []database.Album
	}{
		Albums: albums,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "albums", payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) handleArtists(w http.ResponseWriter, r *http.Request) {
	artists, err := s.db.GetAllArtists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	payload := struct {
		Artists []database.Artist
	}{
		Artists: artists,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "artists", payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongsByArtistID(w http.ResponseWriter, r *http.Request) {
	artistName := r.URL.Query().Get("name")
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) < 4 || paths[2] != "by-artist-id" {
		http.Error(w, "Wrong get request: path should be /songs/by-artist-id/{id}", http.StatusBadRequest)
		log.Printf("ERROR: Wrong get request: path should be /songs/by-artist-id/{id}\n")
		return
	}

	artistID, err := strconv.ParseInt(paths[3], 10, 64)
	if err != nil {
		http.Error(w, "{id}: should be integer: not"+paths[3], http.StatusBadRequest)
		log.Printf("ERROR: {id}: should be integer: not %s]\n", paths[3])
		return
	}

	songs, err := s.db.GetAllMusicsByArtistID(artistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	paylod := struct {
		ArtistName string
		ArtistID   int64
		Songs      []database.Music
	}{
		ArtistName: artistName,
		ArtistID:   artistID,
		Songs:      songs,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "artist-songs", paylod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongsByAlbum(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) < 4 || paths[2] != "by-album" {
		http.Error(w, "Wrong get request: path should be /songs/by-album/{album name}", http.StatusBadRequest)
		log.Printf("ERROR: Wrong get request: path should be /songs/by-album/{album name}")
		return
	}
	albumName := paths[3]
	songs, err := s.db.GetMusicsByAlbumName(albumName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: could not get songs from album(%s) : %s\n", albumName, err.Error())
		return
	}

	paylod := struct {
		AlbumName    string
		AlbumArtPath string
		Songs        []database.Music
	}{
		AlbumName:    albumName,
		AlbumArtPath: songs[0].Path,
		Songs:        songs,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "album-songs", paylod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
}

func (s *httpServer) handleSongDetails(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "url should be /song/details?id={{ .Id }} not "+r.URL.String(), http.StatusBadRequest)
		log.Printf("ERROR: url should be /song/details?id={{ .Id }} not %s\n", r.URL.String())
		return
	}
	toPlay := r.URL.Query().Get("toPlay") == "true"

	songId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("can't convert %s to int", id), http.StatusBadRequest)
		log.Panicf("ERROR: can't convert %s to int\n", id)
		return
	}

	song, err := s.db.GetMusicBYID(songId)

	paylod := struct {
		Song *database.Music
	}{
		Song: song,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "music-details", paylod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	if toPlay {
		s.songsStack.Push(songId)
	}
}

func (s *httpServer) handleSongPlay(w http.ResponseWriter, r *http.Request) {
	songPath := r.URL.Query().Get("music-path")

	if songPath == "" {
		http.Error(w, "No song path provided", http.StatusBadRequest)
		log.Printf("ERROR: No song path provided.\n")
		return
	}

	if _, err := os.Stat(songPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("%s does not exist", songPath), http.StatusInternalServerError)
		log.Printf("ERROR: %s does not exist\n", songPath)
		return
	}

	log.Printf("Requested song path: %q\n", songPath)

	// Set headers for streaming
	mimeType := mime.TypeByExtension(filepath.Ext(songPath))
	if mimeType == "" {
		mimeType = "audio/mpeg" // Fallback for MP3
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Accept-Ranges", "bytes")                               // Enable range requests for seeking
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate") // Minimize browser RAM

	http.ServeFile(w, r, songPath)
}

func (s *httpServer) handleDisplayAlbumArt(w http.ResponseWriter, r *http.Request) {
	songPath := r.URL.Query().Get("music-path")
	songFile, err := os.Open(songPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not open %s: %v", songPath, err), http.StatusBadRequest)
		log.Panicf("ERROR: Could not open %s: %v\n", songPath, err)
		return
	}
	defer songFile.Close()

	tag, err := musictag.ReadFrom(songFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not read tag for %s: %v", songPath, err), http.StatusBadRequest)
		log.Panicf("ERROR: Could not read tag for %s: %v\n", songPath, err)
		return
	}
	albumArt := tag.GetAlbumArt()
	//TODOOOOO: handle for empty albumart
	w.Header().Set("Content-Type", albumArt.MIMEType)
	w.Write(albumArt.Data)
}

func (s *httpServer) handleNextSong(w http.ResponseWriter, r *http.Request) {
	var (
		song   *database.Music
		err    error
		dberr  error
		songId int64
	)

	songId, err = s.songQueue.Dequeue()
	if err == nil {
		song, dberr = s.db.GetMusicBYID(songId)
	} else if err == ErrEmptyQueue {
		song, dberr = s.db.GetRandomMusic()
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}

	if dberr != nil {
		http.Error(w, dberr.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: can't get next song: %s", dberr.Error())
		return
	}

	if song == nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		log.Printf("ERROR: Song not found")
		return
	}

	// Prepare the payload
	payload := map[string]any{
		"id":   song.Id,
		"path": song.Path,
	}

	// Marshal the payload to JSON
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err.Error())
		return
	}

	// Set the response header and write the JSON payload
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
}

func (s *httpServer) handlePreviousSong(w http.ResponseWriter, r *http.Request) {
	if len(s.songsStack.array) < 2 {
		http.Error(w, ErrEmptyStack.Error()+" Play more song", http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", ErrEmptyStack.Error())
		return
	}

	s.songsStack.Pop()
	songId, err := s.songsStack.Pop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPreviousSong(): %s\n", err.Error())
		return
	}

	var songPath string
	err = s.db.DB.QueryRow("SELECT music_location FROM musics WHERE id = ?", songId).Scan(&songPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPreviousSong(): %s\n", err.Error())
		return
	}

	payload := map[string]any{
		"id":   songId,
		"path": songPath,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPreviousSong(): %s\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
}

func (s *httpServer) handlePlayAll(w http.ResponseWriter, r *http.Request) {
	s.songQueue.Clear()

	quaryType := r.URL.Query().Get("type")
	quaryValue := r.URL.Query().Get("value")
	if quaryValue == "" {
		http.Error(w, "Error: Empty value url: /play-all?type=${type}&value=${value}", http.StatusBadRequest)
		log.Printf("Error: Empty value url: /play-all?type=${type}&value=${value}\n")
		return
	}

	var songs []database.Music
	var err error

	switch quaryType {
	case "album":
		albumName := quaryValue
		songs, err = s.db.GetMusicsByAlbumName(albumName)
	case "artist":
		var artistId int64
		artistId, err = strconv.ParseInt(quaryValue, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error: %s\n", err.Error())
			return
		}

		songs, err = s.db.GetAllMusicsByArtistID(artistId)
	default:
		http.Error(w, "Error: Empty type url: /play-all?type=${type}&value=${value}", http.StatusBadRequest)
		log.Printf("Error: Empty type url: /play-all?type=${type}&value=${value}\n")
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error: %s\n", err.Error())
		return
	}

	ids := make([]int64, len(songs))
	for i, s := range songs {
		ids[i] = s.Id
	}
	s.songQueue.Enqueue(ids...)

	songId, err := s.songQueue.Dequeue()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPlayAll(): %s\n", err.Error())
		return
	}

	song, err := s.db.GetMusicBYID(songId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPlayAll(): %s\n", err.Error())
		return
	}

	payload := map[string]any{
		"id":   song.Id,
		"path": song.Path,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: In handelPlayAll(): %s\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
}
