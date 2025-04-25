package main

import (
	"context"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := newCmd().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
