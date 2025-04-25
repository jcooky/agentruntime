package di

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

func Get[T any](ctx context.Context, c *Container, key ObjectKey) (res T, err error) {
	var ok bool
	res, ok = c.objects[key].(T)
	if ok {
		return
	}

	fn, ok := registry[key]
	if !ok {
		err = errors.Errorf("object %s not registered", key)
		return
	}

	obj, err := fn(ctx, c)
	if err != nil {
		return
	}

	c.objects[key] = obj

	res, ok = obj.(T)
	if !ok {
		err = errors.Errorf("object %s is not of type %T", key, res)
		return
	}

	return
}

func MustGet[T any](ctx context.Context, c *Container, key ObjectKey) T {
	res, err := Get[T](ctx, c, key)
	if err != nil {
		panic(fmt.Sprintf("error: %+v", err))
	}

	return res
}

func Set[T any](c *Container, key ObjectKey, obj T) {
	c.objects[key] = obj
}
