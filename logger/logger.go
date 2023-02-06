package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var formatMap = map[string]io.Writer{
	"prod": os.Stdout,
	"dev":  zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
}

func New(env string) zerolog.Logger {
	return zerolog.New(formatMap[env]).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}

func NewFileLogger(env, filename string) (zerolog.Logger, error) {
	file, err := os.OpenFile(
		filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		return zerolog.Logger{}, err
	}

	return zerolog.New(file).With().Timestamp().Logger(), nil
}
