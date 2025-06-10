package mylog

import (
	"log/slog"
	"os"

	"github.com/jcooky/go-din"
)

type Logger = slog.Logger

var (
	Key = din.NewRandomName()
)

func ToLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewLogger(logLevel string, logHandler string) *Logger {
	slogLevel := ToLogLevel(logLevel)

	var handler slog.Handler
	switch logHandler {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slogLevel,
		})
	default:
		handler = newHandler(slogLevel, os.Stderr)
	}

	return slog.New(handler)
}
