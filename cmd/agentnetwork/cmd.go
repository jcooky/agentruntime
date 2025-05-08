package main

import (
	"fmt"
	"github.com/jcooky/go-din"
	"net"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/thread"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
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

			// Initialize the container
			cfg := din.MustGetT[*config.NetworkConfig](c)
			logger := din.MustGet[*mylog.Logger](c, mylog.Key)
			dbInstance := din.MustGet[*gorm.DB](c, db.Key)
			threadManagerServer := din.MustGetT[thread.ThreadManagerServer](c)
			agentNetworkServer := din.MustGetT[network.AgentNetworkServer](c)

			logger.Debug("start agent-runtime", "config", cfg)

			// auto migrate the database
			if err := db.AutoMigrate(dbInstance); err != nil {
				return errors.Wrapf(err, "failed to migrate database")
			}

			// prepare to listen the grpc server
			lc := net.ListenConfig{}
			listener, err := lc.Listen(c, "tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s:%d", cfg.Host, cfg.Port)
			}

			logger.Info("Starting server", "addr", cfg.Host, "port", cfg.Port)

			server := grpc.NewServer(
				grpc.UnaryInterceptor(grpcutils.NewUnaryServerInterceptor(c)),
			)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())
			thread.RegisterThreadManagerServer(server, threadManagerServer)
			network.RegisterAgentNetworkServer(server, agentNetworkServer)

			go func() {
				<-c.Done()
				server.GracefulStop()
			}()

			// start the grpc server
			return server.Serve(listener)
		},
	}
}
