package mylog

import (
	"context"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"log/slog"
	"os"
)

type Logger = slog.Logger

var (
	Key = di.NewKey()
)

func NewLogger(logLevel string, logHandler string) *Logger {
	var slogLevel slog.Level
	switch logLevel {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

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
	di.Register(Key, func(c context.Context, _ di.Env) (any, error) {
		conf, err := di.Get[*config.LogConfig](c, config.LogConfigKey)
		if err != nil {
			return nil, err
		}

		logger := NewLogger(conf.LogLevel, conf.LogHandler)
		slog.SetDefault(logger)
		return logger, nil
	})
}
