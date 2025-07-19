package main

import (
	"fmt"
	"log"
	"music-go/database"
	"music-go/server"
	"music-go/utils"
	"os"
	"path/filepath"
)

func main() {
	cfg := utils.ReadConfig("config.json")
	logger, err := utils.NewCLogger(*cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close()

	db, err := database.OpenConnection(*cfg, *logger)
	if err != nil {
		log.Fatal(err)
	}
    defer db.Close()

    // musics, err := FindAllFilesRecursively(os.Args[1])
    // logger.Println(err)
    // err = db.PushMusicsTOmusicsTable(musics)
    // logger.Println(err)

	server, err := server.NewServer(*cfg, db, *logger)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Serve())
}

// find all the file recursively in a given directory
func FindAllFilesRecursively(dir_path string) ([]string, error) {
	entries, err := os.ReadDir(dir_path)

	if err != nil {
		return nil, fmt.Errorf("ERROR: unable to read directory %s: %w", dir_path, err)
	}

	var files []string

	for _, entry := range entries {
		fullPath := filepath.Join(dir_path, entry.Name())

		if entry.IsDir() {
			sub_entries, _ := FindAllFilesRecursively(fullPath)
			if sub_entries != nil {
				files = append(files, sub_entries...)
			}
		} else {
			files = append(files, fullPath)
		}
	}

	return files, nil
}
