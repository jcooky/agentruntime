package openaiapi

import (
	"encoding/json"
	"github.com/firebase/genkit/go/ai"
	goopenai "github.com/openai/openai-go"
)

func translateResponse(resp *goopenai.ChatCompletion, jsonMode bool) *ai.ModelResponse {
	r := &ai.ModelResponse{}
	translateCandidate(resp.Choices[0], jsonMode, r)

	r.Usage = &ai.GenerationUsage{
		InputTokens:  int(resp.Usage.PromptTokens),
		OutputTokens: int(resp.Usage.CompletionTokens),
		TotalTokens:  int(resp.Usage.TotalTokens),
	}
	r.Custom = resp
	return r
}

func translateCandidate(choice goopenai.ChatCompletionChoice, jsonMode bool, r *ai.ModelResponse) {
	switch choice.FinishReason {
	case "stop", "tool_calls":
		r.FinishReason = ai.FinishReasonStop
	case "length":
		r.FinishReason = ai.FinishReasonLength
	case "content_filter":
		r.FinishReason = ai.FinishReasonBlocked
	case "function_call":
		r.FinishReason = ai.FinishReasonOther
	default:
		r.FinishReason = ai.FinishReasonUnknown
	}

	m := &ai.Message{
		Role: ai.RoleModel,
	}

	// handle tool calls
	var toolRequestParts []*ai.Part
	for _, toolCall := range choice.Message.ToolCalls {
		toolRequestParts = append(toolRequestParts, ai.NewToolRequestPart(&ai.ToolRequest{
			Name:  toolCall.Function.Name,
			Input: json.RawMessage(toolCall.Function.Arguments),
			Ref:   toolCall.ID,
		}))
	}
	if len(toolRequestParts) > 0 {
		m.Content = toolRequestParts
		r.Message = m
		return
	}

	if jsonMode {
		m.Content = append(m.Content, ai.NewDataPart(choice.Message.Content))
	} else {
		m.Content = append(m.Content, ai.NewTextPart(choice.Message.Content))
	}

	r.Message = m
}
