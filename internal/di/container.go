package di

import (
	"github.com/google/uuid"
)

type (
	Env       string
	Container struct {
		objects map[ObjectKey]any
		Env     Env
	}
	ObjectKey uuid.UUID
)

const (
	EnvProd Env = "prod"
	EnvTest Env = "test"
)

func NewKey() ObjectKey {
	return ObjectKey(uuid.New())
}

func NewContainer(env Env) *Container {
	return &Container{
		Env:     env,
		objects: make(map[ObjectKey]any),
	}
}
