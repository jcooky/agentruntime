# Anthropic Plugin for Genkit

This plugin enables the use of Anthropic's Claude models (including extended thinking capabilities) within the Genkit framework.

## Features

- Support for Claude 4, Claude 3.7, and Claude 3.5 models:
  - `claude-4-opus` - Claude Opus 4 (extended thinking support)
  - `claude-4-sonnet` - Claude Sonnet 4 (extended thinking support)
  - `claude-3.7-sonnet` - Claude 3.7 Sonnet (extended thinking support, latest model)
  - `claude-3.5-haiku` - Claude 3.5 Haiku (fastest, cheaper)
  - `claude-3-opus` - Claude 3 Opus
- Streaming and non-streaming generation
- Multimodal support (text and images)
- Tool use (function calling)
- System messages
- Extended thinking capability for Claude 4 models
- Full usage tracking

## Installation

The plugin is included in the agentruntime project. To use it, ensure you have the Anthropic SDK dependency:

```bash
go get github.com/anthropics/anthropic-sdk-go
```

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/habiliai/agentruntime/internal/genkit/plugins/anthropic"
)

func main() {
    ctx := context.Background()
    g, err := genkit.Init(ctx, genkit.WithPlugins(&anthropic.Plugin{
        APIKey: "your-api-key", // or set ANTHROPIC_API_KEY environment variable
    }))
    if err != nil {
        log.Fatal(err)
    }

    // Get a model
    model := anthropic.Model(g, "claude-3.5-haiku")

    // Generate text
    resp, err := model.Generate(ctx, &ai.GenerateRequest{
        Messages: []*ai.Message{
            {
                Role: "user",
                Content: []*ai.Part{
                    {Text: "Hello, Claude!"},
                },
            },
        },
        Config: &ai.GenerationCommonConfig{
            MaxOutputTokens: 1024,
            Temperature:     ai.Ptr(0.7),
        },
    }, nil)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Candidates[0].Message.Content[0].Text)
}
```

### Simple Text Generation

```go
req := &ai.ModelRequest{
    Messages: []*ai.Message{
        {
            Role: ai.RoleUser,
            Content: []*ai.Part{
                ai.NewTextPart("What is the capital of France?"),
            },
        },
    },
    Config: &ai.GenerationCommonConfig{
        MaxOutputTokens: 100,
        Temperature:     0.7,
    },
}

resp, err := model.Generate(ctx, req, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Message.Content[0].Text)
```

### Streaming Generation

```go
err := model.Generate(ctx, req, func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
    if len(chunk.Content) > 0 {
        fmt.Print(chunk.Content[0].Text)
    }
    return nil
})
```

### Multimodal (Image) Input

```go
req := &ai.ModelRequest{
    Messages: []*ai.Message{
        {
            Role: ai.RoleUser,
            Content: []*ai.Part{
                ai.NewTextPart("What's in this image?"),
                ai.NewMediaPart("image/png", base64ImageData),
            },
        },
    },
}
```

### System Messages

```go
req := &ai.ModelRequest{
    Messages: []*ai.Message{
        {
            Role: ai.RoleSystem,
            Content: []*ai.Part{
                ai.NewTextPart("You are a helpful assistant that speaks like a pirate."),
            },
        },
        {
            Role: ai.RoleUser,
            Content: []*ai.Part{
                ai.NewTextPart("Tell me about sailing."),
            },
        },
    },
}
```

### Extended Thinking (Claude 4 models)

Extended thinking is automatically enabled for Claude 4 models when the model determines it would be helpful for complex reasoning tasks. You can influence this by setting appropriate reasoning effort in the config:

```go
req := &ai.ModelRequest{
    Messages: messages,
    Config: map[string]interface{}{
        "reasoning_effort": "high", // Placeholder for future API support
    },
}
```

## Configuration

### Environment Variables

- `ANTHROPIC_API_KEY`: Your Anthropic API key (required if not provided in plugin initialization)

### Supported Models

- `claude-4-opus`: Claude Opus 4 - Most capable model
- `claude-4-sonnet`: Claude Sonnet 4 - High-performance model
- `claude-3.7-sonnet`: Claude 3.7 Sonnet - Extended thinking support, latest model
- `claude-3.5-haiku`: Claude 3.5 Haiku - Fatest, cheaper model
- `claude-3-opus`: Claude 3 Opus

All models support:

- Multi-turn conversations
- Tool/function calling
- System messages
- Multimodal input (text and images)

## Testing

Run unit tests:

```bash
go test ./internal/genkit/plugins/anthropic/...
```

Run integration tests (requires API key):

```bash
ANTHROPIC_API_KEY=your-key go test ./internal/genkit/plugins/anthropic/... -run TestLive
```

## License

This plugin is part of the agentruntime project and follows the same license terms.
