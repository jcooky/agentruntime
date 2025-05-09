package jsonrpc

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/habiliai/agentruntime/internal/mylog"
)

type recoveryLogger struct {
	logger *slog.Logger
}

func (rl *recoveryLogger) Println(v ...any) {
	if len(v) == 0 {
		return
	}

	if err, ok := v[0].(error); ok {
		rl.logger.Error("[JSON-RPC] recovery", mylog.Err(err))
	} else {
		rl.logger.Error("[JSON-RPC] recovery", slog.Any("message", v))
	}
}

func newRecoveryHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return handlers.RecoveryHandler(handlers.RecoveryLogger(&recoveryLogger{logger}))
}
