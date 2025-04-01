package main

import (
	"context"
	"fmt"
	"github.com/habiliai/agentruntime/internal/di"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()

	ctx = di.WithContainer(ctx, di.EnvProd)
	if err := newCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
