package grpcutils

import (
	"context"
	"github.com/jcooky/go-din"
	"log/slog"

	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(c *din.Container, err error) error {
	if err == nil {
		return nil
	}
	logger := din.MustGet[*mylog.Logger](c, mylog.Key)

	logger.Error("error in gRPC handler", "err", err)
	if errors.Is(err, myerrors.ErrNotFound) {
		return status.Errorf(codes.NotFound, "%v", err)
	} else if errors.Is(err, myerrors.ErrInvalidParams) {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		return status.Errorf(codes.Internal, "%v", err)
	}
}

func NewUnaryServerInterceptor(c *din.Container) func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (resp any, err error) {
	logger := din.MustGet[*slog.Logger](c, mylog.Key)
	return func(_ context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger.Info("[gRPC] call", "path", info.FullMethod)

		resp, err = handler(c, req)

		err = handleError(c, err)
		return
	}
}
