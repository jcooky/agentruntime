package mylog

import (
	"context"
	"log/slog"
	"os"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
)

type Logger = slog.Logger

var (
	Key = di.NewKey()
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
	di.Register(Key, func(c context.Context, container *di.Container) (any, error) {
		conf, err := di.Get[*config.LogConfig](c, container, config.LogConfigKey)
		if err != nil {
			return nil, err
		}

		logger := NewLogger(conf.LogLevel, conf.LogHandler)
		return logger, nil
	})
}
