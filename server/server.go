package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Server struct {
	db         DataBase
	port       int
	indexTmpl  *template.Template
	resultTmpl *template.Template
	songsStack Stack
	songQueue  Queue
}

func NewServer(port int) *Server {
	return &Server{
		db:         DataBase{},
		port:       port,
		indexTmpl:  &template.Template{},
		resultTmpl: &template.Template{},
		songsStack: *NewStack(),
		songQueue:  *NewQueue(),
	}
}

func (s *Server) Serve() error {
	//TODO: change path not more general position
	s.db.OpenConnection("data/music.db")
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

	log.Printf("INFO: server is open on 127.0.0.1:%d", s.port)
	fmt.Printf("INFO: server is open on 127.0.0.1:%d\n", s.port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
	return err
}

func (s *Server) NotImplemented(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`<div id="menu-result">Not Implemented yet</div>`))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("ERROR: %s", err.Error())
		return
	}
}

func (s *Server) LoadTemplates() error {
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
