package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/jsonrpc"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/jcooky/go-din"
	"github.com/mokiat/gog"
	"github.com/spf13/cobra"
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

			server := http.Server{
				Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
				Handler: jsonrpc.NewHandlerWithHealth(c, jsonrpc.WithRuntime()),
				BaseContext: func(_ net.Listener) context.Context {
					return c
				},
			}

			server.SetKeepAlivesEnabled(false)

			closeCh := make(chan os.Signal, 3)
			defer close(closeCh)
			signal.Notify(closeCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

			go func() {
				<-closeCh
				if err := server.Shutdown(c); err != nil {
					logger.Error("failed to shutdown server", "err", err)
				}
			}()

			// register agent server
			networkClient := network.NewJsonRpcClient(cfg.NetworkBaseUrl)
			if err := networkClient.RegisterAgent(c, &network.RegisterAgentRequest{
				Addr: cfg.RuntimeBaseUrl,
				Info: agentInfo,
			}); err != nil {
				return errors.Wrapf(err, "failed to register agent")
			}

			agentNames := gog.Map(agentInfo, func(i *network.AgentInfo) string {
				return i.Name
			})
			server.RegisterOnShutdown(func() {
				if err := networkClient.DeregisterAgent(c, &network.DeregisterAgentRequest{
					Names: agentNames,
				}); err != nil {
					logger.Warn("failed to deregister agent", "err", err)
				}
			})
			ctx, cancel := context.WithCancel(c)
			defer cancel()
			go func() {
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := networkClient.CheckLive(ctx, &network.CheckLiveRequest{
							Names: agentNames,
						}); err != nil {
							logger.Warn("failed to check live", "err", err)
						} else {
							logger.Info("agent is alive", "names", agentNames)
						}
					}
				}
			}()

			logger.Info("Starting server", "host", cfg.Host, "port", cfg.Port)
			return server.ListenAndServe()
		},
	}

	return cmd
}
