package di

import "context"

type (
	RegisterFn func(ctx context.Context, c *Container) (any, error)
)

var (
	registry = map[ObjectKey]RegisterFn{}
)

func Register(key ObjectKey, fn RegisterFn) {
	registry[key] = fn
}
