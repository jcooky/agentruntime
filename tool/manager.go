package tool

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/internal/di"
	"github.com/habiliai/agentruntime/internal/genkit/plugins/mcp"
	"github.com/habiliai/agentruntime/internal/mylog"
	mcpclient "github.com/mark3labs/mcp-go/client"
)

type (
	Manager interface {
		mcp.MCPClientRegistry
		LocalToolService
		GetTool(ctx context.Context, toolName string) ai.Tool
		GetMCPTool(ctx context.Context, serverName, toolName string) ai.Tool
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
		}

		go func() {
			<-ctx.Done()
			s.Close()
		}()

		RegisterGetWeatherTool()
		RegisterDoneTool()
		RegisterWebSearchTool()

		return s, nil
	})
}
