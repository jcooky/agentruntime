package main

import (
	"fmt"
	"github.com/habiliai/agentruntime/config"
	di "github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"os"
	"strings"
	"time"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentruntime <agent-file OR agent-files-dir>",
		Short: "Start agent-runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.Errorf("agent-file or agent-files-dir is required")
			}

			var agentFiles []string
			if stat, err := os.Stat(args[0]); os.IsNotExist(err) {
				return errors.Wrapf(err, "agent-file or agent-files-dir does not exist")
			} else if stat.IsDir() {
				files, err := os.ReadDir(args[0])
				if err != nil {
					return errors.Wrapf(err, "failed to read agent-files-dir")
				}
				for _, file := range files {
					if file.IsDir() ||
						(!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
						continue
					}
					agentFiles = append(agentFiles, fmt.Sprintf("%s/%s", args[0], file.Name()))
				}
			} else {
				agentFiles = append(agentFiles, args[0])
			}

			ctx := cmd.Context()
			ctx = di.WithContainer(ctx, di.EnvProd)

			// Initialize the container
			cfg := di.MustGet[*config.RuntimeConfig](ctx, config.RuntimeConfigKey)
			logger := di.MustGet[*mylog.Logger](ctx, mylog.Key)
			runtimeService := di.MustGet[runtime.Service](ctx, runtime.ServiceKey)
			runtimeServer := di.MustGet[runtime.AgentRuntimeServer](ctx, runtime.ServerKey)
			toolManager := di.MustGet[tool.Manager](ctx, tool.ManagerKey)

			logger.Debug("start agent-runtime", "config", cfg)

			// load agent config files
			agentConfigs, err := config.LoadAgentsFromFiles(agentFiles)
			if err != nil {
				return errors.Wrapf(err, "failed to load agent config")
			}

			// register mcp servers or others
			mcpServerChecklist := map[string]struct{}{}
			for _, ac := range agentConfigs {
				for name, mcpServer := range ac.MCPServers {
					if _, ok := mcpServerChecklist[name]; ok {
						continue
					}
					if err := toolManager.RegisterMCPTool(ctx, tool.RegisterMCPToolRequest{
						ServerName: name,
						Command:    mcpServer.Command,
						Args:       mcpServer.Args,
						Env:        mcpServer.Env,
					}); err != nil {
						return err
					}
				}
			}

			// save agents from config files
			var agentNames []string
			for _, ac := range agentConfigs {
				a, err := runtimeService.RegisterAgent(ctx, ac)
				if err != nil {
					return err
				}
				agentNames = append(agentNames, a.Name)

				logger.Info("Agent loaded", "name", ac.Name)
			}

			// prepare to listen the grpc server
			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s:%d", cfg.Host, cfg.Port)
			}

			logger.Info("Starting server", "host", cfg.Host, "port", cfg.Port)

			server := grpc.NewServer(
				grpc.UnaryInterceptor(grpcutils.NewUnaryServerInterceptor(ctx)),
			)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())
			runtime.RegisterAgentRuntimeServer(server, runtimeServer)

			go func() {
				<-ctx.Done()
				server.GracefulStop()
			}()

			// register agent server
			networkCC := di.MustGet[*grpc.ClientConn](ctx, network.GrpcClientConnKey)
			if err != nil {
				return err
			}
			defer networkCC.Close()

			agentManager := network.NewAgentNetworkClient(networkCC)
			if _, err = agentManager.RegisterAgent(ctx, &network.RegisterAgentRequest{
				Addr:   cfg.RuntimeGrpcAddr,
				Secure: false,
				Names:  agentNames,
			}); err != nil {
				return err
			}
			defer func() {
				if _, err := agentManager.DeregisterAgent(ctx, &network.DeregisterAgentRequest{
					Names: agentNames,
				}); err != nil {
					logger.Warn("failed to deregister agent", "err", err)
				}
			}()
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if _, err := agentManager.CheckLive(ctx, &network.CheckLiveRequest{
							Names: agentNames,
						}); err != nil {
							logger.Warn("failed to check live", "err", err)
						} else {
							logger.Info("agent is alive", "names", agentNames)
						}
					}
				}
			}()

			// start the grpc server
			return server.Serve(listener)
		},
	}

	return cmd
}
