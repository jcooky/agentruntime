package main

import (
	"context"
	"github.com/habiliai/agentruntime/internal/di"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = di.WithContainer(ctx, di.EnvProd)
	if err := newCmd().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
