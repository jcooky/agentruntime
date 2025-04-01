package grpcutils_testing

import (
	"context"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientStreamMock struct {
	mock.Mock
}

func (c *ClientStreamMock) Header() (metadata.MD, error) {
	args := c.Called()
	return args.Get(0).(metadata.MD), args.Error(1)
}

func (c *ClientStreamMock) Trailer() metadata.MD {
	args := c.Called()
	return args.Get(0).(metadata.MD)
}

func (c *ClientStreamMock) CloseSend() error {
	args := c.Called()
	return args.Error(0)
}

func (c *ClientStreamMock) Context() context.Context {
	args := c.Called()
	return args.Get(0).(context.Context)
}

func (c *ClientStreamMock) SendMsg(m any) error {
	args := c.Called(m)
	return args.Error(0)
}

func (c *ClientStreamMock) RecvMsg(m any) error {
	args := c.Called(m)
	return args.Error(0)
}

// NewClientStreamMock creates a new mock client stream
func NewClientStreamMock() *ClientStreamMock {
	return &ClientStreamMock{}
}

var (
	_ grpc.ClientStream = (*ClientStreamMock)(nil) // Ensure ClientStreamMock implements grpc.ClientStream
)
