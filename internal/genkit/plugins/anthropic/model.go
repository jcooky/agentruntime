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

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(apiModelName),
		Messages:  messages,
		MaxTokens: 4096, // Default max tokens
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

	// Handle generation config
	if genRequest.Config != nil {
		jsonBytes, err := json.Marshal(genRequest.Config)
		if err != nil {
			return anthropic.MessageNewParams{}, err
		}

		// Try to parse as GenerationCommonConfig
		{
			var c ai.GenerationCommonConfig
			if err := json.Unmarshal(jsonBytes, &c); err == nil {
				if c.MaxOutputTokens > 0 {
					params.MaxTokens = int64(c.MaxOutputTokens)
				}
				if c.Temperature > 0 {
					params.Temperature = anthropic.Float(c.Temperature)
				}
				if c.TopP > 0 {
					params.TopP = anthropic.Float(c.TopP)
				}
				if c.TopK > 0 {
					params.TopK = anthropic.Int(int64(c.TopK))
				}
				if len(c.StopSequences) > 0 {
					params.StopSequences = c.StopSequences
				}
			}
		}

		// Try to parse as GenerationReasoningConfig for extended thinking
		{
			var c ExtendedThinkingConfig
			if err := json.Unmarshal(jsonBytes, &c); err == nil && c.Enabled {
				// Enable extended thinking
				// For Claude 4 models, extended thinking is enabled automatically
				// when the model determines it would be helpful
				params.Thinking = anthropic.ThinkingConfigParamOfEnabled(c.BudgetTokens)
			}
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
			switch part.Custom["type"].(string) {
			case "thinking":
				blocks = append(blocks, anthropic.NewThinkingBlock(part.Custom["signature"].(string), part.Custom["thinking"].(string)))
			default:
				return nil, fmt.Errorf("unsupported custom part type: %s", part.Custom["type"])
			}
		} else if part.IsText() {
			// Use the NewTextBlock helper function
			blocks = append(blocks, anthropic.NewTextBlock(part.Text))
		} else if part.IsMedia() {
			// Handle image content
			data := part.Text

			// Check if it's a base64 data URL
			if strings.HasPrefix(data, "data:") && strings.Contains(data, ";base64,") {
				// Extract base64 data from data URL
				parts := strings.SplitN(data, ",", 2)
				if len(parts) == 2 {
					data = parts[1]
				}

				// Create base64 image source param
				imageSource := anthropic.Base64ImageSourceParam{
					Data:      data,
					MediaType: getAnthropicMediaType(part.ContentType),
				}

				// Create image block with base64 source
				blocks = append(blocks, anthropic.NewImageBlock(imageSource))
			} else {
				// Assume it's a URL
				// Create URL image source param
				imageSource := anthropic.URLImageSourceParam{
					URL: data,
				}

				// Create image block with URL source
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
		case anthropic.ThinkingBlock:
			parts = append(parts, ai.NewCustomPart(map[string]any{
				"type":      "thinking",
				"thinking":  block.Thinking,
				"signature": block.Signature,
			}))
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
