package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/goccy/go-yaml"
	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/internal/mylog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRootCmd() *cobra.Command {
	params := &struct {
		Port int
	}{}
	cmd := &cobra.Command{
		Use:   "agentruntime <agent-file OR agent-files-dir> [...<agent-file OR agent-files-dir>]",
		Short: "Agent runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			logger := mylog.NewLogger("debug", "text")

			var agentFiles []string
			for _, arg := range args {
				if stat, err := os.Stat(arg); os.IsNotExist(err) {
					return errors.Wrapf(err, "agent-file or agent-files-dir does not exist")
				} else if stat.IsDir() {
					files, err := os.ReadDir(arg)
					if err != nil {
						return errors.Wrapf(err, "failed to read agent-files-dir")
					}
					for _, file := range files {
						if file.IsDir() ||
							(!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
							continue
						}
						agentFiles = append(agentFiles, fmt.Sprintf("%s/%s", arg, file.Name()))
					}
				} else {
					agentFiles = append(agentFiles, arg)
				}
			}

			if len(agentFiles) == 0 {
				return errors.New("no agent files found")
			}

			var agents []entity.Agent
			for _, agentFile := range agentFiles {
				var agent entity.Agent
				agentFileBytes, err := os.ReadFile(agentFile)
				if err != nil {
					return errors.Wrapf(err, "failed to read agent file: %s", agentFile)
				}
				if err := yaml.Unmarshal(agentFileBytes, &agent); err != nil {
					return errors.Wrapf(err, "failed to unmarshal agent file: %s", agentFile)
				}
				agents = append(agents, agent)
			}

			db, err := gorm.Open(sqlite.Open("agentruntime.db"), &gorm.Config{})
			if err != nil {
				return errors.Wrap(err, "failed to open database")
			}

			if err := db.AutoMigrate(&Thread{}, &Message{}); err != nil {
				return errors.Wrap(err, "failed to migrate database")
			}

			runtimes := make(map[string]*agentruntime.AgentRuntime)
			for _, agent := range agents {
				runtime, err := agentruntime.NewAgentRuntime(
					ctx,
					agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
					agentruntime.WithAnthropicAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
					agentruntime.WithXAIAPIKey(os.Getenv("XAI_API_KEY")),
					agentruntime.WithLogger(logger),
					agentruntime.WithTraceVerbose(true),
					agentruntime.WithAgent(agent),
				)
				if err != nil {
					return errors.Wrap(err, "failed to create agent runtime")
				}
				runtimes[strings.ToLower(agent.Name)] = runtime
			}

			messageCh := make(chan *Message, len(agents)*2)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer close(messageCh)
				defer wg.Done()
				loopMentionedBy(ctx, db, runtimes, logger, messageCh)
			}()

			handler, err := createServerHandler(
				ctx,
				agents,
				db,
				runtimes,
				logger,
				messageCh,
			)
			if err != nil {
				return err
			}

			logger.Info("server started", "port", params.Port)
			defer logger.Info("server stopped")

			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", params.Port),
				Handler: handler,
				BaseContext: func(l net.Listener) context.Context {
					return ctx
				},
			}

			go func() {
				<-ctx.Done()
				if err := server.Shutdown(context.WithoutCancel(ctx)); err != nil {
					logger.Error("failed to shutdown server", "error", err)
				}
			}()

			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&params.Port, "port", "p", 3001, "Port to listen on")

	return cmd
}

func Execute() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "panic: %+v\n", err)
		os.Exit(1)
	}
}
