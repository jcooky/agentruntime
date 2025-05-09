package mcp

import (
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type (
	InitializeRequest           = mcp.InitializeRequest
	InitializeResult            = mcp.InitializeResult
	ListToolsRequest            = mcp.ListToolsRequest
	ListToolsResult             = mcp.ListToolsResult
	CallToolRequest             = mcp.CallToolRequest
	CallToolResult              = mcp.CallToolResult
	JSONRPCError                = mcp.JSONRPCError
	ToolOption                  = mcp.ToolOption
	Tool                        = mcp.Tool
	ToolInputSchema             = mcp.ToolInputSchema
	ToolListChangedNotification = mcp.ToolListChangedNotification
	Implementation              = mcp.Implementation

	MCPClient      = client.MCPClient
	StdioMCPClient = client.StdioMCPClient
)

var (
	LATEST_PROTOCOL_VERSION = mcp.LATEST_PROTOCOL_VERSION

	NewStdioMCPClient = client.NewStdioMCPClient
)
