package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/handlers"
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
				Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
				Handler: handlers.CORS(
					handlers.AllowedOrigins([]string{"*"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
					handlers.AllowedHeaders([]string{
						"Content-Type",
						"Authorization",
						"Accept",
						"Accept-Language",
						"Accept-Encoding",
						"X-Requested-With",
						"Origin",
						"User-Agent",
						"Referer",
						"Cache-Control",
						"Pragma",
					}),
					handlers.ExposedHeaders([]string{"Content-Length", "Content-Type"}),
					handlers.MaxAge(86400), // Cache preflight for 24 hours
					handlers.AllowCredentials(),
				)(jsonrpc.NewHandlerWithHealth(c, jsonrpc.WithNetwork())),
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
