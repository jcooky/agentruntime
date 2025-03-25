package agentruntime

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/habiliai/agentruntime/agent"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/db"
	di "github.com/habiliai/agentruntime/internal/di"
	interceptors "github.com/habiliai/agentruntime/internal/grpc-interceptors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/habiliai/agentruntime/thread"
	"github.com/habiliai/agentruntime/tool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"net"
	"os"
)

func newServeCmd() *cobra.Command {
	flags := &struct {
		watch bool
	}{}
	cmd := &cobra.Command{
		Use:   "serve <agent-file OR agent-files-dir>",
		Short: "Start agent runtime server",
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
					if file.IsDir() {
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
			agentManager := di.MustGet[agent.Manager](ctx, agent.ManagerKey)
			cfg := di.MustGet[*config.RuntimeConfig](ctx, config.RuntimeConfigKey)
			logger := di.MustGet[*mylog.Logger](ctx, mylog.Key)
			threadManagerServer := di.MustGet[thread.ThreadManagerServer](ctx, thread.ManagerServerKey)
			dbInstance := di.MustGet[*gorm.DB](ctx, db.Key)
			agentManagerServer := di.MustGet[agent.AgentManagerServer](ctx, agent.ManagerServerKey)
			runtimeServer := di.MustGet[runtime.AgentRuntimeServer](ctx, runtime.ServerKey)
			toolManager := di.MustGet[tool.Manager](ctx, tool.ManagerKey)

			logger.Debug("start agent-runtime", "config", cfg)

			// auto migrate the database
			if err := db.AutoMigrate(dbInstance); err != nil {
				return errors.Wrapf(err, "failed to migrate database")
			}

			// load agent config files
			agentConfigs, err := config.LoadAgentsFromFiles(agentFiles)

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
			if err != nil {
				return errors.Wrapf(err, "failed to load agent config")
			}
			for _, ac := range agentConfigs {
				if _, err := agentManager.SaveAgentFromConfig(ctx, ac); err != nil {
					return err
				}

				logger.Info("Agent loaded", "name", ac.Name)
			}

			// register agent file watcher
			var watcher *fsnotify.Watcher
			if flags.watch {
				watcher, err = fsnotify.NewWatcher()
				if err != nil {
					return errors.Wrapf(err, "failed to create watcher")
				}

				for _, file := range agentFiles {
					if err := watcher.Add(file); err != nil {
						return errors.Wrapf(err, "failed to watch file %s", file)
					}
				}

				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case event, ok := <-watcher.Events:
							if !ok {
								return
							}
							if event.Op&fsnotify.Write == fsnotify.Write {
								agentConfigs, err := config.LoadAgentsFromFiles(agentFiles)
								if err != nil {
									logger.Error("Failed to load agent config", "error", err)
									continue
								}
								for _, ac := range agentConfigs {
									if _, err := agentManager.SaveAgentFromConfig(ctx, ac); err != nil {
										logger.Error("Failed to save agent config", "error", err)
										continue
									}

									logger.Info("Agent Reloaded", "name", ac.Name)
								}
							}
						case err, ok := <-watcher.Errors:
							if !ok {
								return
							}
							logger.Error("Watcher error", "error", err)
						}
					}
				}()
			}
			defer func() {
				if watcher != nil {
					watcher.Close()
				}
			}()

			// prepare to listen the grpc server
			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return errors.Wrapf(err, "failed to listen on %s:%d", cfg.Host, cfg.Port)
			}

			logger.Info("Starting server", "host", cfg.Host, "port", cfg.Port)

			server := grpc.NewServer(
				grpc.UnaryInterceptor(interceptors.NewUnaryServerInterceptor(ctx)),
			)
			grpc_health_v1.RegisterHealthServer(server, health.NewServer())
			thread.RegisterThreadManagerServer(server, threadManagerServer)
			agent.RegisterAgentManagerServer(server, agentManagerServer)
			runtime.RegisterAgentRuntimeServer(server, runtimeServer)

			go func() {
				<-ctx.Done()
				server.GracefulStop()
			}()

			// start the grpc server
			return server.Serve(listener)
		},
	}

	return cmd
}
