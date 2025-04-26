package main

import (
	"fmt"
	"log"
	"music-go/server"
	"os"
	"path/filepath"
)

func main() {
	// set log flags
	log.SetFlags(log.Lshortfile)
	// logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer logFile.Close()
	// log.SetOutput(logFile)

	server := server.NewServer(6969)
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
