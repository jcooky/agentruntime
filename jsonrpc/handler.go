package jsonrpc

import (
	"context"
	"net/http"

	"github.com/gorilla/rpc/v2"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/jcooky/go-din"
)

type ServerOption = func(c *din.Container, server *rpc.Server)

func NewHandler(c *din.Container, opts ...ServerOption) http.Handler {
	logger := din.MustGet[*mylog.Logger](c, mylog.Key)

	rpcServer := newRPCServer(c, opts...)

	return newRecoveryHandler(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			rpcServer.ServeHTTP(w, r.WithContext(ctx))
		}),
	)
}

func NewHandlerWithHealth(c *din.Container, opts ...ServerOption) http.Handler {
	logger := din.MustGet[*mylog.Logger](c, mylog.Key)
	mainHandler := NewHandler(c, opts...)
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Warn("failed to write health response", "err", err)
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle different paths
		switch r.URL.Path {
		case "/health":
			healthHandler.ServeHTTP(w, r)
		case "/rpc":
			mainHandler.ServeHTTP(w, r)
		default:
			// For any other path, return 404
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
