package server

import (
	"fmt"
	"html/template"
	"music-go/database"
	"music-go/utils"
	"net/http"
)

type httpServer struct {
	configs    utils.Config
	db         *database.DataBase
	indexTmpl  *template.Template
	resultTmpl *template.Template
	songsStack Stack
	songQueue  Queue
	logger     utils.CLogger
}

func NewServer(config utils.Config, db *database.DataBase, logger utils.CLogger) (*httpServer, error) {
	server := &httpServer{
		configs:    config,
		db:         db,
		songsStack: *NewStack(),
		songQueue:  *NewQueue(),
		logger:     logger,
	}

	if err := server.loadTemplates(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *httpServer) Serve() error {
	defer s.db.Close()

	var mux *http.ServeMux = http.DefaultServeMux

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.handleIndex)

	mux.HandleFunc("/artists", s.handleArtists)
	mux.HandleFunc("/albums", s.handleAlbums)
	mux.HandleFunc("/playlists", s.NotImplemented)
	mux.HandleFunc("/search", s.NotImplemented)

	songMux := http.NewServeMux()
	songMux.HandleFunc("/", s.handleSongs)                        // Handle /songs/
	songMux.HandleFunc("/by-artist-id/", s.handleSongsByArtistID) // Handle /songs/by-artist-id/
	songMux.HandleFunc("/by-album/", s.handleSongsByAlbum)        // Handle /songs/by-album/
	mux.Handle("/songs/", http.StripPrefix("/songs", songMux))

	mux.HandleFunc("/song/details", s.handleSongDetails)
	mux.HandleFunc("/albumArt", s.handleDisplayAlbumArt)
	mux.HandleFunc("/play", s.handleSongPlay)
	mux.HandleFunc("/get-next-song", s.handleGetNextSong)
	mux.HandleFunc("/previous-song", s.handlePreviousSong)
	mux.HandleFunc("/play-all", s.handlePlayAll)

	s.logger.Printf("INFO: server is open on 127.0.0.1:%d", s.configs.Server.Port)
	fmt.Printf("Server is open on 127.0.0.1:%d\n", s.configs.Server.Port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.configs.Server.Port), s.loggingMiddleware(s.recoveryMiddleware(mux)))
	if err != nil && err != http.ErrServerClosed {
		s.logger.Printf("ERROR: Server failed: %v", err)
		return err
	}

	return nil
}

// middle ware for log
func (s *httpServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf("INFO: Request reseved Method: \"%s\" URL: \"%s\"", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}

// recover middle wire
func (s *httpServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Printf("ERROR: panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *httpServer) NotImplemented(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`<div id="menu-result">Not Implemented yet</div>`))
	s.logger.Printf("ERROR: %s not implemented yet.", r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.logger.Printf("ERROR: Could'n write to responce: %s", err.Error())
		return
	}
}

func (s *httpServer) loadTemplates() error {
	var err error
	s.indexTmpl, err = template.ParseFiles("template/index.html")
	if err != nil {
		return err
	}

	s.resultTmpl, err = template.ParseFiles("template/menu-result.html", "template/player.html")
	if err != nil {
		return err
	}

	s.logger.Println("INFO: All the files are parsed from template/")
	return nil
}
