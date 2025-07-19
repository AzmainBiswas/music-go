package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"music-go/database"
	"music-go/musictag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TODOOOO: add request check to all handeler
func (s *httpServer) checkGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("ERROR: for \"%s\" Method not allowed, Only \"GET\" is allowed.", r.URL.String()), http.StatusBadRequest)
		s.logger.Printf("ERROR: for \"%s\" Method not allowed, Only \"GET\" is allowed.", r.URL.String())
		return false
	}

	return true
}

func (s *httpServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	err := s.indexTmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: Could't excute index.html %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongs(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	musics, err := s.db.GetAllMusics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query songs from database: %s", err.Error())
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
		s.logger.Printf("ERROR: could't execute \"musics\" template %s", err.Error())
		return
	}
}

func (s *httpServer) handleAlbums(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	albums, err := s.db.GetAllAlbums()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query all albums from database: %s", err.Error())
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
		s.logger.Printf("ERROR: could't not execute \"albums\" template %s", err.Error())
		return
	}
}

func (s *httpServer) handleArtists(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	artists, err := s.db.GetAllArtists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query artists from database: %s", err.Error())
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
		s.logger.Printf("ERROR: could't execute \"artist\" template: %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongsByArtistID(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	artistName := r.URL.Query().Get("name")
	urlPrefix := "/by-artist-id/"
	if !strings.HasPrefix(r.URL.Path, urlPrefix) {
		s.logger.Printf("ERROR: prefix not found %s in url", urlPrefix)
		http.NotFound(w, r)
		return
	}

	artistIDstr := strings.TrimPrefix(r.URL.Path, urlPrefix)
	if artistIDstr == "" || artistIDstr == "/" {
		s.logger.Printf("ERROR: Missing artist ID in URL %s", r.URL.String())
		http.Error(w, "Bad Request: Artist ID required", http.StatusBadRequest)
		return
	}

	artistID, err := strconv.ParseInt(artistIDstr, 10, 64)
	if err != nil {
		http.Error(w, "{id}: should be integer: not "+artistIDstr, http.StatusBadRequest)
		s.logger.Printf("ERROR: {id} should be integer value: not %s\n", artistIDstr)
		return
	}

	songs, err := s.db.GetAllMusicsByArtistID(artistID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query all musics by id(%d) name(%s): %s", artistID, artistName, err.Error())
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
		s.logger.Printf("ERROR: failed to execute \"artist-song\" template %s", err.Error())
		return
	}
}

func (s *httpServer) handleSongsByAlbum(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	urlPrefix := "/by-album/"
	if !strings.HasPrefix(r.URL.Path, urlPrefix) {
		s.logger.Printf("ERROR: prefix not found %s", urlPrefix)
		http.NotFound(w, r)
		return
	}

	albumName := strings.TrimPrefix(r.URL.Path, urlPrefix)
	if albumName == "" || albumName == "/" {
		http.Error(w, "Wrong get request: path should be /songs/by-album/{album name}", http.StatusBadRequest)
		s.logger.Printf("ERROR: Wrong get request: path should be /songs/by-album/{album name}")
		return
	}
	songs, err := s.db.GetMusicsByAlbumName(albumName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could not get songs from album(%s) : %s\n", albumName, err.Error())
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
		s.logger.Printf("ERROR: failed to execute \"album-songs\" template: %s\n", err.Error())
		return
	}
}

// TODOOO: send json to server and with js show it for multiple use or do some things
func (s *httpServer) handleSongDetails(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "url should be /song/details?id={{ .Id }} not "+r.URL.String(), http.StatusBadRequest)
		s.logger.Printf("ERROR: url should be /song/details?id={{ .Id }} not %s\n", r.URL.String())
		return
	}
	toPlay := r.URL.Query().Get("toPlay") == "true"

	songId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("can't convert %s to int", id), http.StatusBadRequest)
		s.logger.Printf("ERROR: can't convert %s to int\n", id)
		return
	}

	song, err := s.db.GetMusicBYID(songId)
	if err != nil {
		http.Error(w, "Could query to database: "+err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query song by id %d: %s\n", songId, err.Error())
	}

	paylod := struct {
		Song *database.Music
	}{
		Song: song,
	}

	err = s.resultTmpl.ExecuteTemplate(w, "music-details", paylod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't execute \"music-details\" template: %s\n", err.Error())
		return
	}

	if toPlay {
		s.songsStack.Push(songId)
	}
}

func (s *httpServer) handleSongPlay(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	songPath := r.URL.Query().Get("music-path")

	if songPath == "" {
		http.Error(w, "No song path provided", http.StatusBadRequest)
		s.logger.Printf("ERROR: No song path provided path should be /play?music-path={music path}.\n")
		return
	}

	if _, err := os.Stat(songPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("%s does not exist.", songPath), http.StatusInternalServerError)
		s.logger.Printf("ERROR: %s does not exist.\n", songPath)
		return
	}

	// Set headers for streaming
	mimeType := mime.TypeByExtension(filepath.Ext(songPath))
	if mimeType == "" {
		mimeType = "audio/mpeg" // Fallback for MP3
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Accept-Ranges", "bytes")                               // Enable range requests for seeking
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate") // Minimize browser RAM

	http.ServeFile(w, r, songPath)
	s.logger.Printf("INFO: \"%s\" Song served sucessfuly.\n", songPath)
}

func (s *httpServer) handleDisplayAlbumArt(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	songPath := r.URL.Query().Get("music-path")
	songFile, err := os.Open(songPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not open %s: %v", songPath, err), http.StatusBadRequest)
		s.logger.Printf("ERROR: Could not open %s: %v\n", songPath, err)
		return
	}
	defer songFile.Close()

	tag, err := musictag.ReadFrom(songFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not read tag for %s: %v", songPath, err), http.StatusBadRequest)
		s.logger.Printf("ERROR: Could not read tag for %s: %v\n", songPath, err)
		return
	}
	albumArt := tag.GetAlbumArt()
	//TODOOOOO: handle for empty albumart
	w.Header().Set("Content-Type", albumArt.MIMEType)
	w.Header().Set("Accept-Ranges", "bytes")                               // Enable range requests for seeking
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate") // Minimize browser RAM

	w.Write(albumArt.Data)
	s.logger.Printf("INFO: album art for \"%s\" sucessfuly served.", songPath)
}

func (s *httpServer) handleGetNextSong(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

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
		s.logger.Printf("ERROR: could't access queue. %s", err.Error())
		return
	}

	if dberr != nil {
		http.Error(w, dberr.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't get next song: %s", dberr.Error())
		return
	}

	if song == nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		s.logger.Printf("ERROR: Song not found")
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
		s.logger.Printf("ERROR: could't marshel paylod data to json: %s\n", err.Error())
		return
	}

	// Set the response header and write the JSON payload
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
	s.logger.Printf("INFO: Next song details served sucessfuly: %v", string(payloadJson))
}

func (s *httpServer) handlePreviousSong(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	if len(s.songsStack.array) < 2 {
		http.Error(w, ErrEmptyStack.Error()+" This is first song.", http.StatusInternalServerError)
		s.logger.Printf("ERROR: No previouly played song found %s\n", ErrEmptyStack.Error())
		return
	}

	s.songsStack.Pop()
	songId, err := s.songsStack.Pop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could not pop song from stack, last song in stack: %s\n", err.Error())
		return
	}

	var songPath string
	err = s.db.DB.QueryRow("SELECT music_location FROM musics WHERE id = ?", songId).Scan(&songPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't found music_location for song id %d: %s\n", songId, err.Error())
		return
	}

	payload := map[string]any{
		"id":   songId,
		"path": songPath,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: In handelPreviousSong(): %s\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
	s.logger.Printf("INFO: previouly played song data served sucessfuly %v", string(payloadJson))
}

func (s *httpServer) handlePlayAll(w http.ResponseWriter, r *http.Request) {
	if !s.checkGET(w, r) {
		return
	}

	s.songQueue.Clear()

	quaryType := r.URL.Query().Get("type")
	quaryValue := r.URL.Query().Get("value")
	if quaryValue == "" {
		http.Error(w, "Error: Empty value url: /play-all?type=${type}&value=${value}", http.StatusBadRequest)
		s.logger.Printf("Error: Empty value url: /play-all?type=${type}&value=${value}\n")
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
			s.logger.Printf("Error: could't parse %s to int for artistID: %s\n", quaryValue, err.Error())
			return
		}

		songs, err = s.db.GetAllMusicsByArtistID(artistId)
	default:
		http.Error(w, "Error: Empty type url: /play-all?type=${type}&value=${value}", http.StatusBadRequest)
		s.logger.Printf("Error: Empty type url: /play-all?type=${type}&value=${value}\n")
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("Error: could't query database for songs: %s\n", err.Error())
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
		s.logger.Printf("ERROR: could not get music from queue. : %s\n", err.Error())
		return
	}

	song, err := s.db.GetMusicBYID(songId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: could't query song for song id %d: %s\n", songId, err.Error())
		return
	}

	payload := map[string]any{
		"id":   song.Id,
		"path": song.Path,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: In handelPlayAll(): %s\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadJson)
	s.logger.Printf("INFO: playall data served sucessfuly %v", string(payloadJson))
}
