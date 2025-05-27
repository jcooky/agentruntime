package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/jsonrpc"
	"github.com/jcooky/go-din"
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentnetwork",
		Short: "Agent Network CLI by HabiliAI",
	}

	cmd.AddCommand(
		newNetworkThreadCmd(),
		newNetworkServeCmd(),
		newConnectCmd(),
	)

	return cmd
}

func newNetworkServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve the network",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := din.NewContainer(cmd.Context(), din.EnvProd)
			defer c.Close()
			onSig := make(chan os.Signal, 3)
			defer close(onSig)
			signal.Notify(onSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

			// Initialize the container
			cfg := din.MustGetT[*config.NetworkConfig](c)
			logger := din.MustGet[*mylog.Logger](c, mylog.Key)

			logger.Debug("start agent-network", "config", cfg)

			server := http.Server{
				Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
				Handler: jsonrpc.NewHandlerWithHealth(c, jsonrpc.WithNetwork()),
			}

			go func() {
				<-onSig
				if err := server.Shutdown(c); err != nil {
					logger.Error("failed to shutdown server", "err", err)
				}
			}()

			logger.Info("Starting server", "addr", cfg.Host, "port", cfg.Port)
			return server.ListenAndServe()
		},
	}
}
