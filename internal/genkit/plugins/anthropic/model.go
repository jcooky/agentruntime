package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
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

	return translateResponse(*resp)
}

func generateStream(ctx context.Context, client *anthropic.Client, genRequest *ai.ModelRequest, apiModelName string, cb core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
	params, err := buildMessageParams(genRequest, apiModelName)
	if err != nil {
		return nil, err
	}

	// Use standard streaming API
	stream := client.Messages.NewStreaming(ctx, params)

	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		err := message.Accumulate(event)
		if err != nil {
			return nil, fmt.Errorf("error accumulating message: %w", err)
		}

		// Send chunks to callback
		switch event := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := event.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				chunk := &ai.ModelResponseChunk{
					Content: []*ai.Part{ai.NewTextPart(delta.Text)},
				}
				if err := cb(ctx, chunk); err != nil {
					return nil, err
				}
			}
		case anthropic.ContentBlockStartEvent:
			// Handle the start of a new content block (e.g., thinking block)
			// This is where we might detect the beginning of a reasoning block
			// For now, we'll just continue as the accumulated message will handle it
			// TODO: Implement proper handling of reasoning blocks
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("anthropic streaming error: %w", err)
	}

	return translateResponse(message)
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
			budgetRatio = 0.25 // Default to 25% if not specified
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
			tools[i] = convertTool(tool)
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
				block.UnmarshalJSON([]byte(body))
				blocks = append(blocks, anthropic.ContentBlockParamUnion{
					OfWebSearchToolResult: &block,
				})
			case "redacted_thinking":
				block := anthropic.RedactedThinkingBlockParam{}
				block.UnmarshalJSON([]byte(body))
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

			// Check if it's a URL
			if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
				// Create URL image source param
				imageSource := anthropic.URLImageSourceParam{
					URL: data,
				}

				// Create image block with URL source
				blocks = append(blocks, anthropic.NewImageBlock(imageSource))
			} else {
				// Assume it's base64 data
				// If it has data URL prefix, extract the base64 part
				if strings.HasPrefix(data, "data:") && strings.Contains(data, ";base64,") {
					parts := strings.SplitN(data, ",", 2)
					if len(parts) == 2 {
						data = parts[1]
					}
				}

				// Create base64 image source param
				imageSource := anthropic.Base64ImageSourceParam{
					Data:      data,
					MediaType: getAnthropicMediaType(part.ContentType),
				}

				// Create image block with base64 source
				blocks = append(blocks, anthropic.NewImageBlock(imageSource))
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

			println("inputStr", inputMsg)
			toolUse := anthropic.NewToolUseBlock(
				toolReq.Ref,  // ID
				inputMsg,     // Input as JSON string
				toolReq.Name, // Name
			)
			blocks = append(blocks, toolUse)
		} else if part.IsToolResponse() {
			// Convert tool response to Anthropic format
			toolResp := part.ToolResponse

			// Handle tool response content
			var content string
			switch v := toolResp.Output.(type) {
			case string:
				content = v
			default:
				// Marshal non-string outputs to JSON
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool output: %w", err)
				}
				content = string(jsonBytes)
			}

			toolResult := anthropic.NewToolResultBlock(
				toolResp.Ref, // Tool use ID
				content,      // Content
				false,        // Is error
			)
			blocks = append(blocks, toolResult)
		}
	}

	return blocks, nil
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

func translateResponse(resp anthropic.Message) (*ai.ModelResponse, error) {
	r := &ai.ModelResponse{}

	m := &ai.Message{
		Role: ai.RoleModel,
	}

	var parts []*ai.Part

	for _, content := range resp.Content {
		switch block := content.AsAny().(type) {
		case anthropic.TextBlock:
			parts = append(parts, ai.NewTextPart(block.Text))
		case anthropic.ToolUseBlock:
			parts = append(parts, ai.NewToolRequestPart(&ai.ToolRequest{
				Ref:   block.ID,
				Name:  block.Name,
				Input: json.RawMessage(block.Input),
			}))
		case anthropic.WebSearchToolResultBlock:
			parts = append(parts, ai.NewCustomPart(map[string]any{
				"type": "web_search_tool_result",
				"body": block.RawJSON(),
			}))
		case anthropic.RedactedThinkingBlock:
			parts = append(parts, ai.NewCustomPart(map[string]any{
				"type": "redacted_thinking",
				"body": block.RawJSON(),
			}))
		case anthropic.ThinkingBlock:
			parts = append(parts, ai.NewReasoningPart(block.Thinking, []byte(block.Signature)))
		}
	}

	m.Content = parts
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
