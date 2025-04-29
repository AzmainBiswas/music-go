package server

import (
	"fmt"
	"html/template"
	"log"
	"music-go/database"
	"music-go/utils"
	"net/http"
)

type httpServer struct {
	configs    utils.Config
	db         database.DataBase
	indexTmpl  *template.Template
	resultTmpl *template.Template
	songsStack Stack
	songQueue  Queue
}

func NewServer(config utils.Config) *httpServer {
	return &httpServer{
		configs:    config,
		db:         database.DataBase{},
		indexTmpl:  &template.Template{},
		resultTmpl: &template.Template{},
		songsStack: *NewStack(),
		songQueue:  *NewQueue(),
	}
}

func (s *httpServer) Serve() error {
	//TODO: change path not more general position
	s.db.OpenConnection(s.configs)
	defer s.db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", s.handleIndex)

	http.HandleFunc("/songs", s.handleSongs)
	http.HandleFunc("/artists", s.handleArtists)
	http.HandleFunc("/albums", s.handleAlbums)
	http.HandleFunc("/playlists", s.NotImplemented)
	http.HandleFunc("/search", s.NotImplemented)

	http.HandleFunc("/songs/by-artist-id/", s.handleSongsByArtistID)
	http.HandleFunc("/songs/by-album/", s.handleSongsByAlbum)

	http.HandleFunc("/song/details", s.handleSongDetails)
	http.HandleFunc("/albumArt", s.handleDisplayAlbumArt)
	http.HandleFunc("/play", s.handleSongPlay)
	http.HandleFunc("/next-song", s.handleNextSong)
	http.HandleFunc("/previous-song", s.handlePreviousSong)
	http.HandleFunc("/play-all", s.handlePlayAll)

	log.Printf("INFO: server is open on 127.0.0.1:%d", s.configs.Server.Port)
	fmt.Printf("INFO: server is open on 127.0.0.1:%d\n", s.configs.Server.Port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.configs.Server.Port), nil)
	return err
}

func (s *httpServer) NotImplemented(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`<div id="menu-result">Not Implemented yet</div>`))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *httpServer) LoadTemplates() error {
	var err error
	s.indexTmpl, err = template.ParseFiles("template/index.html")
	if err != nil {
		return err
	}

	s.resultTmpl, err = template.ParseFiles("template/menu-result.html", "template/player.html")
	if err != nil {
		return err
	}

	return nil
}
