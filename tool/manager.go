package tool

import (
	"context"

	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	mygenkit "github.com/habiliai/agentruntime/internal/genkit"
	"github.com/habiliai/agentruntime/internal/mylog"
	mcpclient "github.com/mark3labs/mcp-go/client"
)

type (
	Manager interface {
		GetTool(toolName string) ai.Tool
		GetMCPTool(serverName, toolName string) ai.Tool
		GetMCPTools(ctx context.Context, serverName string) []ai.Tool
		RegisterMCPTool(ctx context.Context, req RegisterMCPToolRequest) error
	}
)

var (
	ManagerKey = di.NewKey()
)

func init() {
	di.Register(ManagerKey, func(ctx context.Context, container *di.Container) (any, error) {
		conf, err := di.Get[*config.ToolConfig](ctx, container, config.ToolConfigKey)
		if err != nil {
			return nil, err
		}

		s := &manager{
			logger:     di.MustGet[*mylog.Logger](ctx, container, mylog.Key),
			config:     conf,
			mcpClients: make(map[string]mcpclient.MCPClient),
			genkit:     di.MustGet[*genkit.Genkit](ctx, container, mygenkit.Key),
		}

		go func() {
			<-ctx.Done()
			s.Close()
		}()

		s.registerGetWeatherTool()
		s.registerDoneTool()
		s.registerWebSearchTool()

		return s, nil
	})
}
