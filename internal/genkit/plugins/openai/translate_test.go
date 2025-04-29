package openai

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/firebase/genkit/go/ai"
	goopenai "github.com/openai/openai-go"
)

func TestTranslateCandidate(t *testing.T) {
	tests := []struct {
		name  string
		input struct {
			choice   goopenai.ChatCompletionChoice
			jsonMode bool
		}
		want *ai.ModelResponse
	}{
		{
			name: "text",
			input: struct {
				choice   goopenai.ChatCompletionChoice
				jsonMode bool
			}{
				choice: goopenai.ChatCompletionChoice{
					Index: 0,
					Message: goopenai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "Tell a joke about dogs.",
					},
					FinishReason: "length",
				},
				jsonMode: false,
			},
			want: &ai.ModelResponse{
				FinishReason: ai.FinishReasonLength,
				Message: &ai.Message{
					Role:    ai.RoleModel,
					Content: []*ai.Part{ai.NewTextPart("Tell a joke about dogs.")},
				},
				Custom: nil,
			},
		},
		{
			name: "json",
			input: struct {
				choice   goopenai.ChatCompletionChoice
				jsonMode bool
			}{
				choice: goopenai.ChatCompletionChoice{
					Index: 0,
					Message: goopenai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "{\"json\": \"test\"}",
					},
					FinishReason: "content_filter",
				},
				jsonMode: true,
			},
			want: &ai.ModelResponse{
				FinishReason: ai.FinishReasonBlocked,
				Message: &ai.Message{
					Role:    ai.RoleModel,
					Content: []*ai.Part{ai.NewDataPart("{\"json\": \"test\"}")},
				},
				Custom: nil,
			},
		},
		{
			name: "tools",
			input: struct {
				choice   goopenai.ChatCompletionChoice
				jsonMode bool
			}{
				choice: goopenai.ChatCompletionChoice{
					Index: 0,
					Message: goopenai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "Tool call",
						ToolCalls: []goopenai.ChatCompletionMessageToolCall{
							{
								ID:   "exampleTool",
								Type: "function",
								Function: goopenai.ChatCompletionMessageToolCallFunction{
									Name:      "exampleTool",
									Arguments: "{\"param\": \"value\"}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
				jsonMode: false,
			},
			want: &ai.ModelResponse{
				FinishReason: ai.FinishReasonStop,
				Message: &ai.Message{
					Role: ai.RoleModel,
					Content: []*ai.Part{ai.NewToolRequestPart(&ai.ToolRequest{
						Name:  "exampleTool",
						Input: json.RawMessage(`{"param": "value"}`),
						Ref:   "exampleTool",
					})},
				},
				Custom: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r ai.ModelResponse
			translateCandidate(tt.input.choice, tt.input.jsonMode, &r)
			assert.Equal(t, tt.want, &r)
		})
	}
}
