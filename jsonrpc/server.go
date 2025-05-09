package jsonrpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/thread"
	"github.com/jcooky/go-din"
)

type (
	StartTimeCtxKey string
)

var (
	startTimeCtxKey StartTimeCtxKey = "jsonrpc.startTime"
)

func WithNetwork() ServerOption {
	return func(c *din.Container, s *rpc.Server) {
		if err := network.RegisterJsonRpcService(c, s); err != nil {
			panic(err)
		}
		if err := thread.RegisterJsonRpcService(c, s); err != nil {
			panic(err)
		}

	}
}

func WithRuntime() ServerOption {
	return func(c *din.Container, s *rpc.Server) {
		if err := runtime.RegisterJsonRpcService(c, s); err != nil {
			panic(err)
		}
	}
}

func newRPCServer(c *din.Container, opts ...ServerOption) *rpc.Server {
	logger := din.MustGet[*mylog.Logger](c, mylog.Key)

	server := rpc.NewServer()
	for _, opt := range opts {
		opt(c, server)
	}
	server.RegisterBeforeFunc(func(i *rpc.RequestInfo) {
		startTime := time.Now()
		ctx := context.WithValue(i.Request.Context(), startTimeCtxKey, startTime)
		req := i.Request.WithContext(ctx)
		i.Request = req
	})
	server.RegisterAfterFunc(func(i *rpc.RequestInfo) {
		logger := logger.WithGroup("jsonrpc")
		if startTime, ok := i.Request.Context().Value(startTimeCtxKey).(time.Time); ok {
			duration := time.Since(startTime)
			logger = logger.With(slog.Duration("duration", duration))
		}
		if i.Error != nil {
			logger = logger.With(mylog.Err(i.Error))
		}
		logger.Info("[JSON-RPC] call",
			slog.Int("statusCode", i.StatusCode),
			slog.String("method", i.Method),
			slog.Bool("error", i.Error != nil),
		)
	})
	server.RegisterCodec(json2.NewCustomCodecWithErrorMapper(
		rpc.DefaultEncoderSelector,
		func(err error) error {
			if err == nil {
				return nil
			}

			logger.Error("[JSON-RPC] error", mylog.Err(err))
			e := &json2.Error{}
			if errors.As(err, e) {
				return err
			}

			e.Message = err.Error()

			if errors.Is(err, errors.ErrInvalidParams) {
				e.Code = json2.E_BAD_PARAMS
			} else if errors.Is(err, errors.ErrInternal) {
				e.Code = json2.E_INTERNAL
			} else if errors.Is(err, errors.ErrInvalidRequest) {
				e.Code = json2.E_INVALID_REQ
			} else if errors.Is(err, errors.ErrNoMore) || errors.Is(err, errors.ErrNotFound) {
				e.Code = json2.E_SERVER
			} else {
				e.Code = json2.E_INTERNAL
			}

			return e
		},
	), "application/json")

	return server
}
