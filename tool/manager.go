package tool

import (
	"context"
	"github.com/jcooky/go-din"
	"log/slog"

	"github.com/firebase/genkit/go/genkit"

	"github.com/firebase/genkit/go/ai"
	"github.com/habiliai/agentruntime/config"
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

func init() {
	din.RegisterT(func(c *din.Container) (Manager, error) {
		conf, err := din.GetT[*config.ToolConfig](c)
		if err != nil {
			return nil, err
		}

		s := &manager{
			logger:     din.MustGet[*slog.Logger](c, mylog.Key),
			config:     conf,
			mcpClients: make(map[string]mcpclient.MCPClient),
			genkit:     din.MustGet[*genkit.Genkit](c, mygenkit.Key),
		}

		go func() {
			<-c.Done()
			s.Close()
		}()

		s.registerGetWeatherTool()
		s.registerDoneTool()
		s.registerWebSearchTool()

		return s, nil
	})
}
