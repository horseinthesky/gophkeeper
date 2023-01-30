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
		Level(zerolog.ErrorLevel).
		With().
		Timestamp().
		Logger()
}
