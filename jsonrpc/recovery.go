package jsonrpc

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/handlers"
)

func newRecoveryHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return handlers.RecoveryHandler(
		handlers.RecoveryLogger(slog.NewLogLogger(logger.Handler(), slog.LevelError)),
		handlers.PrintRecoveryStack(true),
	)
}
