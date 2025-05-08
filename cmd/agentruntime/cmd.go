package main

import (
	"fmt"
	"github.com/jcooky/go-din"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/grpcutils"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/tool"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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
			for _, filename := range args {
				if stat, err := os.Stat(filename); os.IsNotExist(err) {
					return errors.Wrapf(err, "agent-file or agent-files-dir does not exist")
				} else if stat.IsDir() {
					files, err := os.ReadDir(filename)
					if err != nil {
						return errors.Wrapf(err, "failed to read agent-files-dir")
					}
					for _, file := range files {
						if file.IsDir() ||
							(!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
							continue
						}
						agentFiles = append(agentFiles, fmt.Sprintf("%s/%s", filename, file.Name()))
					}
				} else {
					agentFiles = append(agentFiles, filename)
				}
			}

			c := din.NewContainer(cmd.Context(), din.EnvProd)

			// Initialize the container
			cfg := din.MustGetT[*config.RuntimeConfig](c)
			logger := din.MustGet[*slog.Logger](c, mylog.Key)
			runtimeService := din.MustGetT[runtime.Service](c)
			runtimeServer := din.MustGetT[runtime.AgentRuntimeServer](c)
			toolManager := din.MustGetT[tool.Manager](c)

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
					if err := toolManager.RegisterMCPTool(c, tool.RegisterMCPToolRequest{
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
			var agentInfo []*network.AgentInfo
			for _, ac := range agentConfigs {
				a, err := runtimeService.RegisterAgent(c, ac)
				if err != nil {
					return err
				}
				agentInfo = append(agentInfo, &network.AgentInfo{
					Name:     a.Name,
					Role:     a.Role,
					Metadata: a.Metadata,
				})

				logger.Info("Agent loaded", "name", ac.Name)
			}

			// prepare to listen the grpc server
			lc := net.ListenConfig{}
			listener, err := lc.Listen(c, "tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s:%d", cfg.Host, cfg.Port)
			}

			logger.Info("Starting server", "host", cfg.Host, "port", cfg.Port)

			server := grpc.NewServer(
				grpc.UnaryInterceptor(grpcutils.NewUnaryServerInterceptor(c)),
			)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())
			runtime.RegisterAgentRuntimeServer(server, runtimeServer)

			closeCh := make(chan os.Signal, 3)
			defer close(closeCh)
			signal.Notify(closeCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

			go func() {
				<-closeCh
				server.GracefulStop()
			}()

			// register agent server
			agentManager := din.MustGetT[network.AgentNetworkClient](c)
			if _, err = agentManager.RegisterAgent(c, &network.RegisterAgentRequest{
				Addr:   cfg.RuntimeGrpcAddr,
				Secure: false,
				Info:   agentInfo,
			}); err != nil {
				return err
			}

			agentNames := gog.Map(agentInfo, func(i *network.AgentInfo) string {
				return i.Name
			})
			defer func() {
				if _, err := agentManager.DeregisterAgent(c, &network.DeregisterAgentRequest{
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
					case <-c.Done():
						return
					case <-ticker.C:
						if _, err := agentManager.CheckLive(c, &network.CheckLiveRequest{
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
