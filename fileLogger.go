package flock

import (
	"io"
	"os"
	"github.com/sirupsen/logrus"
)

// Create a new instance of the logger. You can have any number of instances.

type FileLogger struct {
	File *os.File
	Log *logrus.Logger
}

func NewFileLogger(daemonPID string) *FileLogger {
	var log = logrus.New()

	//TODO: reset this for prod
	log.Level = logrus.DebugLevel

	// open a file
	file, err := os.OpenFile("out/" + daemonPID + ".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	} else {
	  	log.Info("Failed to log to file, using default stderr")
	}
	return &FileLogger{
		file,
		log,
	}
}