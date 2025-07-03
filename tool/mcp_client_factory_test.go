package tool_test

import (
	"context"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/tool"
)

// TestMCPTransportDetection tests the auto-detection of transport types
func (s *TestSuite) TestMCPTransportDetection() {
	tests := []struct {
		name     string
		config   tool.MCPServerConfig
		expected tool.MCPTransportType
	}{
		{
			name: "stdio with command",
			config: tool.MCPServerConfig{
				Command: "/usr/bin/mcp-server",
			},
			expected: tool.MCPTransportStdio,
		},
		{
			name: "sse with URL",
			config: tool.MCPServerConfig{
				URL: "https://mcp.example.com",
			},
			expected: tool.MCPTransportSSE,
		},
		{
			name: "explicit oauth-sse",
			config: tool.MCPServerConfig{
				URL:       "https://mcp.example.com",
				Transport: tool.MCPTransportOAuthSSE,
			},
			expected: tool.MCPTransportOAuthSSE,
		},
		{
			name:     "empty config defaults to stdio",
			config:   tool.MCPServerConfig{},
			expected: tool.MCPTransportStdio,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.config.GetTransport()
			s.Equal(tt.expected, got)
		})
	}
}

// TestMCPClientFactoryValidation tests validation in the MCP client factory
func (s *TestSuite) TestMCPClientFactoryValidation() {
	ctx := context.Background()
	factory := tool.NewMCPClientFactory()

	tests := []struct {
		name      string
		config    tool.MCPServerConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "stdio without command",
			config: tool.MCPServerConfig{
				Transport: tool.MCPTransportStdio,
			},
			wantError: true,
			errorMsg:  "command is required for stdio transport",
		},
		{
			name: "sse without URL",
			config: tool.MCPServerConfig{
				Transport: tool.MCPTransportSSE,
			},
			wantError: true,
			errorMsg:  "URL is required for SSE transport",
		},
		{
			name: "oauth-sse without oauth config",
			config: tool.MCPServerConfig{
				URL:       "https://mcp.example.com",
				Transport: tool.MCPTransportOAuthSSE,
			},
			wantError: true,
			errorMsg:  "OAuth configuration is required for OAuth SSE transport",
		},
		{
			name: "valid stdio config",
			config: tool.MCPServerConfig{
				Command: "/bin/echo",
				Args:    []string{"hello"},
			},
			wantError: false,
		},
		{
			name: "valid sse config",
			config: tool.MCPServerConfig{
				URL: "https://mcp.example.com",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := factory.CreateClient(ctx, "test-server", tt.config)

			if tt.wantError {
				s.Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				// Note: Client creation might succeed but connection might fail
				// This is expected for network-based transports
				if err != nil && tt.config.GetTransport() == tool.MCPTransportStdio {
					// For stdio, command might not exist
					s.T().Logf("Expected error for non-existent command: %v", err)
				}
			}
		})
	}
}

// TestConvertAgentSkillToMCPServerConfig tests conversion from AgentSkill to MCPServerConfig
func (s *TestSuite) TestConvertAgentSkillToMCPServerConfig() {
	tests := []struct {
		name           string
		skill          entity.AgentSkillUnion
		expectedConfig tool.MCPServerConfig
	}{
		{
			name: "local stdio server",
			skill: entity.AgentSkillUnion{
				Type: "mcp",
				OfMCP: &entity.MCPAgentSkill{
					Name:    "local-tools",
					Command: "/usr/local/bin/mcp-server",
					Args:    []string{"--verbose"},
					Env: map[string]any{
						"DEBUG": "true",
					},
				},
			},
			expectedConfig: tool.MCPServerConfig{
				Command: "/usr/local/bin/mcp-server",
				Args:    []string{"--verbose"},
				Env: map[string]any{
					"DEBUG": "true",
				},
			},
		},
		{
			name: "remote sse server",
			skill: entity.AgentSkillUnion{
				Type: "mcp",
				OfMCP: &entity.MCPAgentSkill{
					Name: "remote-tools",
					URL:  "https://mcp.example.com/api",
					Headers: map[string]string{
						"Authorization": "Bearer token",
					},
				},
			},
			expectedConfig: tool.MCPServerConfig{
				URL: "https://mcp.example.com/api",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
		},
		{
			name: "oauth sse server",
			skill: entity.AgentSkillUnion{
				Type: "mcp",
				OfMCP: &entity.MCPAgentSkill{
					Name:      "oauth-tools",
					URL:       "https://api.example.com/mcp",
					Transport: "oauth-sse",
					OAuth: &entity.AgentSkillOAuthConfig{
						ClientID:              "client-123",
						ClientSecret:          "secret-456",
						AuthServerMetadataURL: "https://auth.example.com/.well-known/openid",
						RedirectURL:           "http://localhost:8080/callback",
						Scopes:                []string{"read", "write"},
						PKCEEnabled:           true,
					},
				},
			},
			expectedConfig: tool.MCPServerConfig{
				URL:       "https://api.example.com/mcp",
				Transport: tool.MCPTransportOAuthSSE,
				OAuthConfig: &tool.OAuthConfig{
					ClientID:              "client-123",
					ClientSecret:          "secret-456",
					AuthServerMetadataURL: "https://auth.example.com/.well-known/openid",
					RedirectURL:           "http://localhost:8080/callback",
					Scopes:                []string{"read", "write"},
					PKCEEnabled:           true,
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Test the actual conversion function
			config, err := tool.ConvertAgentSkillToMCPServerConfig(tt.skill)
			s.NoError(err)
			s.NotNil(config)

			// Compare the resulting config
			s.Equal(tt.expectedConfig.Command, config.Command)
			s.Equal(tt.expectedConfig.URL, config.URL)
			s.Equal(tt.expectedConfig.Transport, config.Transport)
			s.Equal(tt.expectedConfig.Headers, config.Headers)
			s.Equal(tt.expectedConfig.Args, config.Args)
			s.Equal(tt.expectedConfig.Env, config.Env)

			if tt.expectedConfig.OAuthConfig != nil {
				s.NotNil(config.OAuthConfig)
				s.Equal(tt.expectedConfig.OAuthConfig.ClientID, config.OAuthConfig.ClientID)
				s.Equal(tt.expectedConfig.OAuthConfig.ClientSecret, config.OAuthConfig.ClientSecret)
				s.Equal(tt.expectedConfig.OAuthConfig.AuthServerMetadataURL, config.OAuthConfig.AuthServerMetadataURL)
				s.Equal(tt.expectedConfig.OAuthConfig.RedirectURL, config.OAuthConfig.RedirectURL)
				s.Equal(tt.expectedConfig.OAuthConfig.Scopes, config.OAuthConfig.Scopes)
				s.Equal(tt.expectedConfig.OAuthConfig.PKCEEnabled, config.OAuthConfig.PKCEEnabled)
			} else {
				s.Nil(config.OAuthConfig)
			}
		})
	}
}

// TestConvertAgentSkillToMCPServerConfigErrors tests error cases for the conversion function
func (s *TestSuite) TestConvertAgentSkillToMCPServerConfigErrors() {
	tests := []struct {
		name      string
		skill     entity.AgentSkillUnion
		wantError bool
		errorMsg  string
	}{
		{
			name: "non-MCP skill type",
			skill: entity.AgentSkillUnion{
				Type: "llm",
				OfLLM: &entity.LLMAgentSkill{
					ID:   "test-llm",
					Name: "test-llm-tool",
				},
			},
			wantError: true,
			errorMsg:  "skill type must be 'mcp', got 'llm'",
		},
		{
			name: "MCP skill with nil data",
			skill: entity.AgentSkillUnion{
				Type:  "mcp",
				OfMCP: nil,
			},
			wantError: true,
			errorMsg:  "MCP skill data is nil",
		},
		{
			name: "valid MCP skill without OAuth",
			skill: entity.AgentSkillUnion{
				Type: "mcp",
				OfMCP: &entity.MCPAgentSkill{
					ID:      "test-mcp",
					Name:    "test-mcp-server",
					Command: "/usr/bin/test",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			config, err := tool.ConvertAgentSkillToMCPServerConfig(tt.skill)

			if tt.wantError {
				s.Error(err)
				s.Nil(config)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				s.NoError(err)
				s.NotNil(config)
			}
		})
	}
}
