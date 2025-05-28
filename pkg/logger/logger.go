package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logDir := "./logs"
	logFile := filepath.Join(logDir, "log.txt")

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		panic("unable to create logs directory: " + err.Error())
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic("unable to open log file: " + err.Error())
	}

	multi := io.MultiWriter(file)
	Log = zerolog.New(multi).With().Timestamp().Logger()
}
