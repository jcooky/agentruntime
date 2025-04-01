package main

import (
	"fmt"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	"github.com/habiliai/agentruntime/internal/di"
	interceptors "github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/thread"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"net"
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
			ctx := cmd.Context()
			ctx = di.WithContainer(ctx, di.EnvProd)

			// Initialize the container
			cfg := di.MustGet[*config.NetworkConfig](ctx, config.NetworkConfigKey)
			logger := di.MustGet[*mylog.Logger](ctx, mylog.Key)
			dbInstance := di.MustGet[*gorm.DB](ctx, db.Key)
			threadManagerServer := di.MustGet[thread.ThreadManagerServer](ctx, thread.ManagerServerKey)
			agentNetworkServer := di.MustGet[network.AgentNetworkServer](ctx, network.ManagerServerKey)

			logger.Debug("start agent-runtime", "config", cfg)

			// auto migrate the database
			if err := db.AutoMigrate(dbInstance); err != nil {
				return errors.Wrapf(err, "failed to migrate database")
			}

			// prepare to listen the grpc server
			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s:%d", cfg.Host, cfg.Port)
			}

			logger.Info("Starting server", "addr", cfg.Host, "port", cfg.Port)

			server := grpc.NewServer(
				grpc.UnaryInterceptor(interceptors.NewUnaryServerInterceptor(ctx)),
			)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())
			thread.RegisterThreadManagerServer(server, threadManagerServer)
			network.RegisterAgentNetworkServer(server, agentNetworkServer)

			go func() {
				<-ctx.Done()
				server.GracefulStop()
			}()

			// start the grpc server
			return server.Serve(listener)
		},
	}
}
