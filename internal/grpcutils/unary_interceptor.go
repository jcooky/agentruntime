package grpcutils

import (
	"context"
	myerrors "github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	logger := di.MustGet[*mylog.Logger](ctx, mylog.Key)

	logger.Error("error in gRPC handler", "err", err)
	if errors.Is(err, myerrors.ErrNotFound) {
		return status.Errorf(codes.NotFound, "%v", err)
	} else if errors.Is(err, myerrors.ErrInvalidParams) {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	} else {
		return status.Errorf(codes.Internal, "%v", err)
	}
}

func NewUnaryServerInterceptor(ctx context.Context) func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (resp any, err error) {
	logger := di.MustGet[*mylog.Logger](ctx, mylog.Key)
	return func(_ context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger.Info("[gRPC] call", "path", info.FullMethod)

		resp, err = handler(ctx, req)

		err = handleError(ctx, err)
		return
	}
}
