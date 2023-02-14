package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	formatMap = map[string]io.Writer{
		"prod": os.Stdout,
		"dev":  zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	}

	levelMap = map[string]zerolog.Level{
		"prod": zerolog.WarnLevel,
		"dev":  zerolog.InfoLevel,
	}
)

func New(env string) zerolog.Logger {
	return zerolog.New(formatMap[env]).
		Level(levelMap[env]).
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

	var out io.Writer
	out = file
	if env == "dev" {
		out = zerolog.ConsoleWriter{Out: file, TimeFormat: time.RFC3339}
	}

	return zerolog.New(out).
		Level(levelMap[env]).
		With().
		Timestamp().
		Logger(), nil
}
