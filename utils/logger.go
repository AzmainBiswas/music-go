package utils

import (
	"io"
	"log"
	"os"
)

const LogFileLocation = "./log.log"

type CLogger struct {
	Logger  *log.Logger
	LogFile *os.File
}

func NewCLogger(config Config) (*CLogger, error) {
	var cl *CLogger = &CLogger{}

	// if logging is disable in config
	if !config.Log.Enable {
		cl.Logger = log.New(io.Discard, "", 0)
		cl.LogFile = nil
		return cl, nil
	}

	var writers []io.Writer
	if config.Log.Destination == LogToConsole || config.Log.Destination == LogToBoth {
		writers = append(writers, os.Stdout)
		cl.LogFile = nil
	}

	if config.Log.Destination == LogToFile || config.Log.Destination == LogToBoth {
		var err error
		cl.LogFile, err = os.OpenFile(LogFileLocation, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		writers = append(writers, cl.LogFile)
	}

	if len(writers) == 0 {
		cl.LogFile = nil
		cl.Logger = log.New(io.Discard, "", 0)
		return cl, nil
	}

	multiWriter := io.MultiWriter(writers...)
	cl.Logger = log.New(multiWriter, "[MUSIC-GO] ", log.Ldate|log.Ltime)
	return cl, nil
}

func (cl *CLogger) Close() error {
	if cl.LogFile != nil {
		return cl.LogFile.Close()
	}
	return nil
}

func (cl *CLogger) Println(v ...any) {
	cl.Logger.Println(v...)
}

func (cl *CLogger) Printf(format string, v ...any) {
	cl.Logger.Printf(format, v...)
}
