package mylog

import (
	"github.com/jcooky/go-din"
	"log/slog"
	"os"

	"github.com/habiliai/agentruntime/config"
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

func init() {
	din.Register(Key, func(c *din.Container) (any, error) {
		conf, err := din.GetT[*config.LogConfig](c)
		if err != nil {
			return nil, err
		}

		logger := NewLogger(conf.LogLevel, conf.LogHandler)
		return logger, nil
	})
}
