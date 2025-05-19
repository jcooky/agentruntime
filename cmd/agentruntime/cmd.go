package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/errors"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/habiliai/agentruntime/internal/tool"
	"github.com/habiliai/agentruntime/memory"
	"github.com/habiliai/agentruntime/network"
	"github.com/habiliai/agentruntime/runtime"
	"github.com/jcooky/go-din"
	"github.com/spf13/cobra"
)

func processMentioned(
	ctx context.Context,
	mentionedThreadIds []uint32,
	agentInfo *network.AgentInfo,
	runtimeService runtime.Service,
	logger *slog.Logger,
) {
	for _, threadId := range mentionedThreadIds {
		logger.Info("mention received", "name", agentInfo.Name, "thread_id", threadId)
		found, err := runtimeService.FindAgentsByNames([]string{agentInfo.Name})
		if err != nil {
			logger.Warn("failed to find agent by name", "name", agentInfo.Name, "err", err)
			continue
		}
		if len(found) == 0 {
			logger.Warn("agent not found", "name", agentInfo.Name)
			continue
		}
		if err := runtimeService.Run(ctx, uint(threadId), found); err != nil {
			logger.Warn("failed to run agent", "name", agentInfo.Name, "thread_id", threadId, "err", err)
			continue
		}
		logger.Info("agent run completed", "name", agentInfo.Name, "thread_id", threadId)
	}
}

func startAgentLoop(
	agentInfo *network.AgentInfo,
	networkClient network.JsonRpcClient,
	memoryService memory.Service,
	runtimeService runtime.Service,
	logger *slog.Logger,
	c *din.Container,
) {
	agentContext, err := memoryService.GetContext(c, agentInfo.Name)
	if err != nil {
		logger.Warn("failed to get agent context", "name", agentInfo.Name, "err", err)
	}
	defer func() {
		if err := memoryService.SetContext(c, agentContext); err != nil {
			logger.Warn("failed to set agent context", "name", agentInfo.Name, "err", err)
		} else {
			logger.Info("agent context saved", "name", agentInfo.Name)
		}
	}()

	ctx, cancel := signal.NotifyContext(c, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	liveTicker := time.NewTicker(30 * time.Second)
	defer liveTicker.Stop()
	mentionTicker := time.NewTicker(500 * time.Millisecond)
	defer mentionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-liveTicker.C:
			// check if agent is alive
			if err := networkClient.CheckLive(ctx, &network.CheckLiveRequest{
				Names: []string{agentInfo.Name},
			}); err != nil {
				logger.Warn("failed to check live", "err", err)
				break
			}
			logger.Info("agent is alive", "name", agentInfo.Name)

		case <-mentionTicker.C:
			// get mentions for the agent
			var mentionedThreadIds []uint32
			if reply, err := networkClient.IsMentionedOnce(ctx, &network.IsMentionedRequest{
				AgentName: agentInfo.Name,
			}); err != nil {
				logger.Warn("failed to get mentions", "err", err)
				break
			} else {
				logger.Debug("mention received", "name", agentInfo.Name, "thread_ids", reply.ThreadIds)
				mentionedThreadIds = reply.ThreadIds
			}
			if len(mentionedThreadIds) == 0 {
				break
			}

			// process the mentions
			processMentioned(ctx, mentionedThreadIds, agentInfo, runtimeService, logger)
		}
	}
}

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentruntime <agent-file OR agent-files-dir>",
		Short: "Start agent-runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

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

			c := din.NewContainer(ctx, din.EnvProd)

			// Initialize the container
			cfg := din.MustGetT[*config.RuntimeConfig](c)
			logger := din.MustGet[*slog.Logger](c, mylog.Key)
			runtimeService := din.MustGetT[runtime.Service](c)
			toolManager := din.MustGetT[tool.Manager](c)
			memoryConfig := din.MustGetT[*config.MemoryConfig](c)
			var memoryService memory.Service
			if memoryConfig.SqliteEnabled {
				memoryService = din.MustGet[memory.Service](c, memory.SqliteServiceName)
			} else {
				return errors.New("memory service is not enabled, please check the configuration")
			}

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

			// register agent server
			networkClient := network.NewJsonRpcClient(cfg.NetworkBaseUrl)
			if err := networkClient.RegisterAgent(c, &network.RegisterAgentRequest{
				Addr: cfg.RuntimeBaseUrl,
				Info: agentInfo,
			}); err != nil {
				return errors.Wrapf(err, "failed to register agent")
			}

			// start agent runtime
			var wg sync.WaitGroup
			for _, agentInfo := range agentInfo {
				wg.Add(1)
				go func(agentInfo *network.AgentInfo) {
					defer wg.Done()
					defer func() {
						if err := networkClient.DeregisterAgent(c, &network.DeregisterAgentRequest{
							Names: []string{agentInfo.Name},
						}); err != nil {
							logger.Warn("failed to deregister agent", "err", err)
						}
					}()

					startAgentLoop(agentInfo, networkClient, memoryService, runtimeService, logger, c)
				}(agentInfo)
			}
			wg.Wait()

			logger.Info("agent runtime stopped")
			return nil
		},
	}

	return cmd
}
