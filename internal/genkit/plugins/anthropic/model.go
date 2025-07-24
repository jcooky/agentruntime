package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// DefineModel creates and registers a new generative model with Genkit.
func DefineModel(g *genkit.Genkit, client *anthropic.Client, labelPrefix, provider, modelName, apiModelName string, caps ai.ModelSupports) ai.Model {
	meta := &ai.ModelInfo{
		Label:    labelPrefix + " - " + modelName,
		Supports: &caps,
	}

	return genkit.DefineModel(
		g,
		provider,
		modelName,
		meta,
		func(ctx context.Context, req *ai.ModelRequest, cb core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
			if cb == nil {
				// Non-streaming generation
				return generate(ctx, client, req, apiModelName)
			}
			// Streaming generation
			return generateStream(ctx, client, req, apiModelName, cb)
		},
	)
}

func generate(ctx context.Context, client *anthropic.Client, genRequest *ai.ModelRequest, apiModelName string) (*ai.ModelResponse, error) {
	params, err := buildMessageParams(genRequest, apiModelName)
	if err != nil {
		return nil, err
	}

	// Use standard Messages API (which supports extended thinking for Claude 4 models)
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic message generation failed: %w", err)
	}

	return translateResponse(*resp, genRequest)
}

func generateStream(ctx context.Context, client *anthropic.Client, genRequest *ai.ModelRequest, apiModelName string, cb core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
	params, err := buildMessageParams(genRequest, apiModelName)
	if err != nil {
		return nil, err
	}

	// Use standard streaming API
	stream := client.Messages.NewStreaming(ctx, params)
	defer stream.Close()

	type ToolUsePart struct {
		Index int64
		Ref   string
		Name  string
		Input string
	}
	type WebSearchToolResultPart struct {
		Index  int64
		Ref    string
		Name   string
		Result string
	}
	var toolUsePart *ToolUsePart
	var webSearchToolResultPart *WebSearchToolResultPart

	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		err := message.Accumulate(event)
		if err != nil {
			return nil, fmt.Errorf("error accumulating message: %w", err)
		}
		switch event := event.AsAny().(type) {
		case anthropic.ContentBlockStartEvent:
			switch block := event.ContentBlock.AsAny().(type) {
			case anthropic.ToolUseBlock:
				if toolUsePart != nil {
					return nil, fmt.Errorf("received tool use block but no tool use part found")
				}
				toolUsePart = &ToolUsePart{
					Index: event.Index,
					Ref:   block.ID,
					Name:  block.Name,
				}
			case anthropic.ServerToolUseBlock:
				if toolUsePart != nil {
					return nil, fmt.Errorf("received server tool use block but no tool use part found")
				}
				toolUsePart = &ToolUsePart{
					Index: event.Index,
					Ref:   block.ID,
					Name:  string(block.Name),
				}
			case anthropic.WebSearchToolResultBlock:
				if webSearchToolResultPart != nil {
					return nil, fmt.Errorf("received web search tool result block but no web search tool result part found")
				}
				webSearchToolResultPart = &WebSearchToolResultPart{
					Index:  event.Index,
					Ref:    block.ToolUseID,
					Name:   "web_search",
					Result: block.Content.RawJSON(),
				}
			}
		case anthropic.ContentBlockDeltaEvent:
			chunk := &ai.ModelResponseChunk{
				Index:      int(event.Index),
				Role:       ai.RoleModel,
				Aggregated: false,
			}

			switch delta := event.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				chunk.Content = []*ai.Part{ai.NewTextPart(delta.Text)}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			case anthropic.InputJSONDelta:
				if toolUsePart == nil {
					return nil, fmt.Errorf("received input JSON delta but no tool use part found")
				}
				toolUsePart.Input += delta.PartialJSON
			case anthropic.ThinkingDelta:
				chunk.Content = []*ai.Part{ai.NewReasoningPart(delta.Thinking, []byte{})}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			case anthropic.SignatureDelta:
				chunk.Content = []*ai.Part{ai.NewReasoningPart("", []byte(delta.Signature))}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			case anthropic.CitationsDelta:
				var citation map[string]any
				if err := json.Unmarshal([]byte(delta.Citation.RawJSON()), &citation); err != nil {
					return nil, fmt.Errorf("could not unmarshal citation delta into citation type: %w", err)
				}
				chunk.Content = []*ai.Part{ai.NewCustomPart(map[string]any{
					"type": "citation",
					"body": citation,
				})}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			}
		case anthropic.ContentBlockStopEvent:
			if toolUsePart != nil {
				defer func() {
					toolUsePart = nil
				}()
				chunk := &ai.ModelResponseChunk{
					Index:      int(event.Index),
					Role:       ai.RoleModel,
					Aggregated: false,
					Content: []*ai.Part{ai.NewToolRequestPart(&ai.ToolRequest{
						Ref:   toolUsePart.Ref,
						Name:  toolUsePart.Name,
						Input: json.RawMessage(toolUsePart.Input),
					})},
				}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			}
			if webSearchToolResultPart != nil {
				defer func() {
					webSearchToolResultPart = nil
				}()
				chunk := &ai.ModelResponseChunk{
					Index:      int(event.Index),
					Role:       ai.RoleModel,
					Aggregated: false,
					Content: []*ai.Part{ai.NewToolResponsePart(&ai.ToolResponse{
						Ref:    webSearchToolResultPart.Ref,
						Name:   webSearchToolResultPart.Name,
						Output: json.RawMessage(webSearchToolResultPart.Result),
					})},
				}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("anthropic streaming error: %w", err)
	}

	return translateResponse(message, genRequest)
}

func buildMessageParams(genRequest *ai.ModelRequest, apiModelName string) (anthropic.MessageNewParams, error) {
	messages, systems, err := convertMessages(genRequest.Messages)
	if err != nil {
		return anthropic.MessageNewParams{}, err
	}

	defaultParams, ok := defaultModelParams[apiModelName]
	if !ok {
		return anthropic.MessageNewParams{}, fmt.Errorf("model %s not found", apiModelName)
	}

	params := anthropic.MessageNewParams{
		Model:    anthropic.Model(apiModelName),
		Messages: messages,
	}

	// Convert systems prompt to TextBlockParam array
	for _, system := range systems {
		if strings.TrimSpace(system) == "" {
			continue
		}
		params.System = append(params.System, anthropic.TextBlockParam{
			Text: system,
		})
	}

	if genRequest.Config == nil {
		genRequest.Config = map[string]any{}
	}

	// Handle generation config
	jsonBytes, err := json.Marshal(genRequest.Config)
	if err != nil {
		return anthropic.MessageNewParams{}, err
	}

	// Extract and apply extended thinking config
	type configWithExtendedThinking struct {
		ai.GenerationCommonConfig
		ExtendedThinkingConfig
	}

	// Start with defaults
	config := configWithExtendedThinking{
		GenerationCommonConfig: defaultParams.GenerationCommonConfig,
		ExtendedThinkingConfig: defaultParams.ExtendedThinkingConfig,
	}

	// Unmarshal user config - this will only override provided fields
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return anthropic.MessageNewParams{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply basic config
	if config.MaxOutputTokens > 0 {
		params.MaxTokens = int64(config.MaxOutputTokens)
	} else {
		return anthropic.MessageNewParams{}, fmt.Errorf("maxOutputTokens is required")
	}
	if config.Temperature > 0 {
		params.Temperature = anthropic.Float(config.Temperature)
	}
	if config.TopP > 0 {
		params.TopP = anthropic.Float(config.TopP)
	}
	if config.TopK > 0 {
		params.TopK = anthropic.Int(int64(config.TopK))
	}
	if len(config.StopSequences) > 0 {
		params.StopSequences = config.StopSequences
	}

	// Apply extended thinking configuration
	if config.ExtendedThinkingEnabled {
		// Calculate budget based on ratio
		budgetRatio := config.ExtendedThinkingBudgetRatio
		if budgetRatio == 0 {
			budgetRatio = 0.5 // Default to 50% if not specified
		}

		budget := int64(float64(config.MaxOutputTokens) * budgetRatio)

		// Only enable if budget meets minimum requirement (1024 tokens)
		if budget >= 1024 {
			params.Thinking = anthropic.ThinkingConfigParamOfEnabled(budget)
		}
	}

	// Handle tools if present
	if len(genRequest.Tools) > 0 {
		tools := make([]anthropic.ToolUnionParam, len(genRequest.Tools))
		for i, tool := range genRequest.Tools {
			switch tool.Name {
			case "web_search":
				tools[i] = anthropic.ToolUnionParam{
					OfWebSearchTool20250305: &anthropic.WebSearchTool20250305Param{
						MaxUses: anthropic.Int(99),
					},
				}
			default:
				tools[i] = convertTool(tool)
			}
		}
		params.Tools = tools
	}

	return params, nil
}

func convertMessages(messages []*ai.Message) ([]anthropic.MessageParam, []string, error) {
	var systems []string
	var anthropicMessages []anthropic.MessageParam

	for _, msg := range messages {
		var role anthropic.MessageParamRole
		switch msg.Role {
		case ai.RoleUser:
			role = anthropic.MessageParamRoleUser
		case ai.RoleModel:
			role = anthropic.MessageParamRoleAssistant
		case ai.RoleTool:
			role = anthropic.MessageParamRoleUser
		case ai.RoleSystem:
			for _, part := range msg.Content {
				if part.IsText() && part.Text != "" {
					systems = append(systems, part.Text)
				}
			}
			continue
		default:
			return nil, nil, fmt.Errorf("unsupported message role: %s", msg.Role)
		}

		content, err := convertContent(msg.Content)
		if err != nil {
			return nil, nil, err
		}

		anthropicMessages = append(anthropicMessages, anthropic.MessageParam{
			Role:    role,
			Content: content,
		})
	}

	return anthropicMessages, systems, nil
}

func convertContent(parts []*ai.Part) ([]anthropic.ContentBlockParamUnion, error) {
	var blocks []anthropic.ContentBlockParamUnion

	for _, part := range parts {
		if part.IsCustom() {
			custom := part.Custom
			customType, ok := custom["type"].(string)
			if !ok {
				return nil, fmt.Errorf("custom type not found in custom part")
			}
			body, ok := custom["body"].(string)
			if !ok {
				return nil, fmt.Errorf("custom body not found in custom part")
			}
			switch customType {
			case "web_search_tool_result":
				block := anthropic.WebSearchToolResultBlockParam{}
				if err := block.UnmarshalJSON([]byte(body)); err != nil {
					return nil, fmt.Errorf("failed to unmarshal web search tool result: %w", err)
				}
				blocks = append(blocks, anthropic.ContentBlockParamUnion{
					OfWebSearchToolResult: &block,
				})
			case "redacted_thinking":
				block := anthropic.RedactedThinkingBlockParam{}
				if err := block.UnmarshalJSON([]byte(body)); err != nil {
					return nil, fmt.Errorf("failed to unmarshal redacted thinking: %w", err)
				}
				blocks = append(blocks, anthropic.ContentBlockParamUnion{
					OfRedactedThinking: &block,
				})
			default:
				return nil, fmt.Errorf("unsupported custom type: %s", customType)
			}
		} else if part.IsReasoning() {
			signature, ok := part.Metadata["signature"].([]byte)
			if !ok {
				return nil, fmt.Errorf("signature not found in reasoning part")
			}
			blocks = append(blocks, anthropic.NewThinkingBlock(string(signature), part.Text))
		} else if part.IsText() {
			// Use the NewTextBlock helper function
			blocks = append(blocks, anthropic.NewTextBlock(part.Text))
		} else if part.IsMedia() {
			// Handle image content
			data := part.Text

			isUrl := strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://")
			// Check if it's a URL
			if !isUrl && strings.HasPrefix(data, "data:") {
				if !strings.Contains(data, ";base64,") {
					return nil, fmt.Errorf("data URL is not base64 encoded")
				}
				parts := strings.SplitN(data, ",", 2)
				if len(parts) == 2 {
					data = parts[1]
				}
			}

			switch strings.ToLower(part.ContentType) {
			case "image/jpeg", "image/png", "image/webp", "image/gif", "image/jpg":
				if isUrl {
					// Create image block with URL source
					blocks = append(blocks, anthropic.NewImageBlock(anthropic.URLImageSourceParam{
						URL: data,
					}))
				} else {
					// Create image block with base64 source
					blocks = append(blocks, anthropic.NewImageBlock(anthropic.Base64ImageSourceParam{
						Data:      data,
						MediaType: getAnthropicMediaType(part.ContentType),
					}))
				}
			case "application/pdf":
				if isUrl {
					blocks = append(blocks, anthropic.NewDocumentBlock(anthropic.URLPDFSourceParam{
						URL: data,
					}))
				} else {
					blocks = append(blocks, anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{
						Data: data,
					}))
				}
			case "text/plain":
				if isUrl {
					resp, err := http.Get(data)
					if err != nil {
						return nil, fmt.Errorf("failed to get URL: %w", err)
					}
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, fmt.Errorf("failed to read URL body: %w", err)
					}
					data = string(body)
				}
				blocks = append(blocks, anthropic.NewDocumentBlock(anthropic.PlainTextSourceParam{
					Data: data,
				}))
			default:
				return nil, fmt.Errorf("unsupported media type: %s", part.ContentType)
			}

		} else if part.IsToolRequest() {
			// Convert tool request to Anthropic format
			toolReq := part.ToolRequest

			// Marshal the input to get the string representation
			var inputMsg json.RawMessage
			if toolReq.Input != nil {
				inputJSON, err := json.Marshal(toolReq.Input)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool input: %w", err)
				}
				inputMsg = inputJSON
			}
			if len(inputMsg) == 0 {
				inputMsg = json.RawMessage("{}")
			}

			toolUse := anthropic.NewToolUseBlock(
				toolReq.Ref,  // ID
				inputMsg,     // Input as JSON string
				toolReq.Name, // Name
			)
			blocks = append(blocks, toolUse)
		} else if part.IsToolResponse() {
			// Convert tool response to Anthropic format
			toolResp := part.ToolResponse

			contents, err := convertToolResultBlockContents(toolResp.Output)
			if err != nil {
				blocks = append(blocks, anthropic.NewToolResultBlock(
					toolResp.Ref,
					err.Error(),
					true,
				))
			} else {
				blocks = append(blocks, anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: toolResp.Ref,
						Content:   contents,
						IsError:   anthropic.Opt(false),
					},
				})
			}
		}
	}

	return blocks, nil
}

func convertToolResultBlockContents(output any) (contents []anthropic.ToolResultBlockParamContentUnion, err error) {
	// Handle tool response content
	switch v := output.(type) {
	case string:
		contents = append(contents, anthropic.ToolResultBlockParamContentUnion{
			OfText: &anthropic.TextBlockParam{
				Text: v,
			},
		})
	case []any:
		for _, item := range v {
			children, err := convertToolResultBlockContents(item)
			if err != nil {
				return nil, err
			}
			contents = append(contents, children...)
		}
	case map[string]any:
		if v, ok := v["error"]; ok {
			var content string
			switch err := v.(type) {
			case string:
				content = err
			case error:
				content = err.Error()
			default:
				content = fmt.Sprintf("%v", err)
			}

			if content != "" {
				contents = append(contents, anthropic.ToolResultBlockParamContentUnion{
					OfText: &anthropic.TextBlockParam{
						Text: content,
					},
				})
			}
		}
		if output, ok := v["output"]; ok {
			children, err := convertToolResultBlockContents(output)
			if err != nil {
				return nil, err
			}
			contents = append(contents, children...)
			break
		}

		if contentType, ok := v["contentType"].(string); ok {
			if url, ok := v["url"].(string); ok {
				// Handle ai.Media type
				contents = append(contents, anthropic.ToolResultBlockParamContentUnion{
					OfImage: &anthropic.ImageBlockParam{
						Source: func() (source anthropic.ImageBlockParamSourceUnion) {
							if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
								source.OfURL = &anthropic.URLImageSourceParam{
									URL: url,
								}
							} else {
								source.OfBase64 = &anthropic.Base64ImageSourceParam{
									Data:      url,
									MediaType: getAnthropicMediaType(contentType),
								}
							}
							return
						}(),
					},
				})
				break
			}
		}

		// Marshal non-string outputs to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		contents = append(contents, anthropic.ToolResultBlockParamContentUnion{
			OfText: &anthropic.TextBlockParam{
				Text: string(jsonBytes),
			},
		})

	default:
		return nil, fmt.Errorf("unsupported tool result block content type: %T", v)
	}

	return
}

func convertTool(tool *ai.ToolDefinition) anthropic.ToolUnionParam {
	// Convert the InputSchema map to ToolInputSchemaParam
	inputSchema := anthropic.ToolInputSchemaParam{
		Type:       "object",
		Properties: tool.InputSchema["properties"],
	}

	return anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: inputSchema,
		},
	}
}

func translateContent(content anthropic.ContentBlockUnion) *ai.Part {
	switch block := content.AsAny().(type) {
	case anthropic.TextBlock:
		return ai.NewTextPart(block.Text)
	case anthropic.ToolUseBlock:
		return ai.NewToolRequestPart(&ai.ToolRequest{
			Ref:   block.ID,
			Name:  block.Name,
			Input: json.RawMessage(block.Input),
		})
	case anthropic.WebSearchToolResultBlock:
		return ai.NewCustomPart(map[string]any{
			"type": "web_search_tool_result",
			"body": block.RawJSON(),
		})
	case anthropic.RedactedThinkingBlock:
		return ai.NewCustomPart(map[string]any{
			"type": "redacted_thinking",
			"body": block.RawJSON(),
		})
	case anthropic.ThinkingBlock:
		return ai.NewReasoningPart(block.Thinking, []byte(block.Signature))
	case anthropic.ServerToolUseBlock:
		return ai.NewCustomPart(map[string]any{
			"type": "server_tool_use",
			"body": block.RawJSON(),
		})
	}

	return nil
}

func translateContents(contents []anthropic.ContentBlockUnion) []*ai.Part {
	var parts []*ai.Part

	for _, content := range contents {
		parts = append(parts, translateContent(content))
	}

	return parts
}

func translateResponse(resp anthropic.Message, genRequest *ai.ModelRequest) (*ai.ModelResponse, error) {
	r := &ai.ModelResponse{}

	m := &ai.Message{
		Role: ai.RoleModel,
	}

	m.Content = translateContents(resp.Content)
	r.Message = m

	// Map stop reason
	switch resp.StopReason {
	case anthropic.StopReasonEndTurn:
		r.FinishReason = ai.FinishReasonStop
	case anthropic.StopReasonMaxTokens:
		r.FinishReason = ai.FinishReasonLength
	case anthropic.StopReasonStopSequence:
		r.FinishReason = ai.FinishReasonStop
	case anthropic.StopReasonToolUse:
		r.FinishReason = ai.FinishReasonStop
	default:
		if resp.StopReason != "" {
			r.FinishReason = ai.FinishReasonOther
		}
	}

	// Extract usage information
	if resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
		r.Usage = &ai.GenerationUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			TotalTokens:  int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		}
	}

	// Set custom data
	r.Custom = resp
	r.Request = genRequest

	return r, nil
}

func getAnthropicMediaType(mimeType string) anthropic.Base64ImageSourceMediaType {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg":
		return anthropic.Base64ImageSourceMediaTypeImageJPEG
	case "image/png":
		return anthropic.Base64ImageSourceMediaTypeImagePNG
	case "image/gif":
		return anthropic.Base64ImageSourceMediaTypeImageGIF
	case "image/webp":
		return anthropic.Base64ImageSourceMediaTypeImageWebP
	default:
		// Default to JPEG if unsupported
		return anthropic.Base64ImageSourceMediaTypeImageJPEG
	}
}
