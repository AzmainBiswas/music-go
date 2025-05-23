package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const ConfigErrExitCode int = 2

type LogDestination int

const (
	LogToConsole LogDestination = iota
	LogToFile
	LogToBoth
)

func (ld LogDestination) MarshalJSON() ([]byte, error) {
	switch ld {
	case LogToConsole:
		return json.Marshal("console")
	case LogToFile:
		return json.Marshal("file")
	case LogToBoth:
		return json.Marshal("both")
	default:
		return nil, fmt.Errorf("Invalid LogDestination: %d", ld)
	}
}

func (ld *LogDestination) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "console":
		*ld = LogToConsole
	case "file":
		*ld = LogToFile
	case "both":
		*ld = LogToBoth
	default:
		return fmt.Errorf("invalid log destination: %s", s)
	}
	return nil
}

type Config struct {
	MusicDir string `json:"music_dir"`
	Database struct {
		Path string `json:"path"`
	} `json:"database"`
	Server struct {
		Port uint64 `json:"port"`
	} `json:"server"`
	Log struct {
		Enable      bool           `json:"enable"`
		Destination LogDestination `json:"destination"` // 0 -> console, 1 -> log file, 2 -> both
	}
}

func newDefaultConfig() *Config {
	var defaultConfig *Config = &Config{}
	defaultConfig.MusicDir = "~/Music"
	defaultConfig.Database.Path = "./data"
	defaultConfig.Server.Port = 6969
	defaultConfig.Log.Enable = true
	defaultConfig.Log.Destination = LogToBoth

	return defaultConfig
}

func writeDefaultConfig(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("ERROR: Could not open file %s: %s\n", path, err.Error())
		os.Exit(ConfigErrExitCode)
	}
	defer file.Close()

	var cfg = *newDefaultConfig()
	jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: Coun't not convert struct to json: %s\n", err.Error())
		os.Exit(ConfigErrExitCode)
	}

	if _, err := file.Write(jsonBytes); err != nil {
		fmt.Printf("ERROR: could't write the data to file: %s\n", err.Error())
		os.Exit(ConfigErrExitCode)
	}
}

func ReadConfig(path string) *Config {
	var cfg *Config = newDefaultConfig()
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		writeDefaultConfig(path)
		return cfg
	} else if err != nil {
		fmt.Printf("ERROR: could't open file %s: %s\n", path, err.Error())
		os.Exit(ConfigErrExitCode)
	}

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(file); err != nil {
		fmt.Printf("ERROR: could't read byte: %s\n", err.Error())
		os.Exit(ConfigErrExitCode)
	}

	if err := json.Unmarshal(buffer.Bytes(), cfg); err != nil {
		fmt.Printf("ERROR: could't parse %s: %s\n", path, err.Error())
		fmt.Println("ERROR: Using default config.")
		return newDefaultConfig()
	}

	return cfg
}
