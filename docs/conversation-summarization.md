# Conversation Summarization

The conversation history summarization feature provided by AgentRuntime allows maintaining context without exceeding token limits even in long conversations.

## Background

When conversations with AI agents become lengthy, the token count can exceed the model's context window. To address this issue, AgentRuntime provides automatic conversation summarization functionality.

## How It Works

1. **Token Counting**:

   - Accurately measure **text** tokens using tiktoken
   - Calculate **image** tokens using OpenAI's official algorithm
   - Estimate calculation for **other media** files like PDFs, audio, etc.

2. **Separated Token Management**:

   - **Conversation History** (past conversations): Contains text only
   - **Current Request Files**: Managed separately in RunRequest.Files
   - **Total Tokens = Conversation History Tokens + Current Request File Tokens**

3. **Smart Truncation**: Compress conversation history considering file tokens
4. **Automatic Summarization**: Convert old conversations into meaningful summaries using LLM
5. **Context Preservation**: Maintain continuity by combining files + summary + recent conversations

## Configuration

### Using Default Settings

```go
runtime, err := agentruntime.NewAgentRuntime(
    ctx,
    agentruntime.WithAgent(agent),
    agentruntime.WithOpenAIAPIKey(apiKey),
    agentruntime.WithDefaultConversationSummary(), // Enable summarization with default settings
)
```

### Setting Token Limit Only

```go
runtime, err := agentruntime.NewAgentRuntime(
    ctx,
    agentruntime.WithAgent(agent),
    agentruntime.WithOpenAIAPIKey(apiKey),
    agentruntime.WithConversationSummaryTokenLimit(8000), // 8k token limit
)
```

### Detailed Configuration

```go
summaryConfig := config.ConversationSummaryConfig{
    MaxTokens:                   10000,      // Maximum token count
    SummaryTokens:               2000,       // Target tokens per summary
    MinConversationsToSummarize: 8,          // Minimum conversations for summarization
    ModelForSummary:             "gpt-4o",   // Model to use for summarization
}

runtime, err := agentruntime.NewAgentRuntime(
    ctx,
    agentruntime.WithAgent(agent),
    agentruntime.WithOpenAIAPIKey(apiKey),
    agentruntime.WithConversationSummary(summaryConfig),
)
```

## Configuration Options

| Option                        | Default       | Description                                    |
| ----------------------------- | ------------- | ---------------------------------------------- |
| `MaxTokens`                   | 100,000       | Maximum tokens for entire conversation history |
| `SummaryTokens`               | 2,000         | Target token count for each summary            |
| `MinConversationsToSummarize` | 10            | Minimum conversations to trigger summarization |
| `ModelForSummary`             | "gpt-4o-mini" | LLM model used for summary generation          |

### Token Provider Options

| Provider      | Description                               | Required Setup                           |
| ------------- | ----------------------------------------- | ---------------------------------------- |
| `"auto"`      | Auto-detect from model name (recommended) | None                                     |
| `"openai"`    | Use OpenAI tiktoken                       | None                                     |
| `"anthropic"` | Use Anthropic count_tokens API            | `ANTHROPIC_API_KEY` environment variable |

## Usage Example

See the complete example in `examples/conversation_summary_example.go`.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/habiliai/agentruntime"
    "github.com/habiliai/agentruntime/config"
    "github.com/habiliai/agentruntime/engine"
    "github.com/habiliai/agentruntime/entity"
)

func main() {
    ctx := context.Background()

    agent := entity.Agent{
        Name:        "ConversationBot",
        Role:        "Assistant",
        Description: "A bot that demonstrates conversation summarization",
        Prompt:      "You are a helpful assistant.",
        ModelName:   "gpt-4o-mini",
    }

    // Configure conversation summarization
    runtime, err := agentruntime.NewAgentRuntime(
        ctx,
        agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
        agentruntime.WithAgent(agent),
        agentruntime.WithDefaultConversationSummary(),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer runtime.Close()

    // Test with long conversation history
    longHistory := createLongConversationHistory()

    req := engine.RunRequest{
        History: longHistory,
        ThreadInstruction: "Help the user with their questions.",
    }

    response, err := runtime.Run(ctx, req, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %s\n", response.Text())
}
```

## Operation Process

1. **Token Measurement**: Calculate total token count of current conversation history
2. **Limit Check**: Verify if `MaxTokens` limit is exceeded
3. **Find Split Point**: Determine point to summarize old conversations while keeping recent ones
4. **Generate Summary**: Use LLM to convert old conversations into meaningful summary
5. **Combine**: Include summary and recent conversations in prompt

## Summary Quality

The summary includes the following information:

- **Key Topics**: Core subjects discussed
- **Important Decisions**: Decisions that were made
- **Action Items**: Tasks or to-do items mentioned
- **Context**: Background helpful for future conversations
- **User Preferences**: Revealed user preferences or information

## Performance Considerations

- **Token Counting**: Accurate token measurement using tiktoken library
- **Summary Model**: Cost optimization using efficient models like `gpt-4o-mini`
- **Caching**: Performance improvement by caching summaries for identical conversations
- **Incremental Summarization**: Minimize latency by generating summaries only when needed

## Troubleshooting

### When Summaries Are Not Generated

1. Check if `MaxTokens` setting is sufficiently high (default: 100,000)
2. Verify conversation count exceeds `MinConversationsToSummarize`
3. Ensure API key is properly configured

### When Summary Quality Is Poor

1. Change `ModelForSummary` to a more powerful model (e.g., "gpt-4o")
2. Increase `SummaryTokens` for more detailed summaries
3. Add summarization guidelines to agent's system prompt

### Performance Issues

1. Reduce `SummaryTokens` to shorten summary length
2. Increase `MinConversationsToSummarize` to reduce summarization frequency
3. Use faster models (e.g., "gpt-4o-mini")

## Related Documentation

- **[Multi-Provider Token Counting](./multi-provider-token-counting.md)**: Detailed guide for model provider-specific token calculation
- **[Token Architecture](./token-architecture.md)**: Design principles of the token management system

## API Reference

For detailed API documentation, see GoDoc:

```bash
go doc github.com/habiliai/agentruntime/engine.ConversationSummarizer
go doc github.com/habiliai/agentruntime/engine.TokenCounter
go doc github.com/habiliai/agentruntime/engine.OpenAITokenCounter
go doc github.com/habiliai/agentruntime/engine.AnthropicTokenCounter
go doc github.com/habiliai/agentruntime/config.ConversationSummaryConfig
```

### Key Methods

```go
// Count tokens in text
func (cs *ConversationSummarizer) CountTokens(text string) int

// Count tokens in files (images, PDFs, etc.)
func (cs *ConversationSummarizer) CountFileTokens(contentType, data string) int

// Count tokens in conversation history (text only)
func (cs *ConversationSummarizer) CountConversationTokens(conversations []Conversation) int

// Count tokens in current request files
func (cs *ConversationSummarizer) CountRequestFilesTokens(files []File) int

// Process conversation history (summarization + truncation, considering request files)
func (cs *ConversationSummarizer) ProcessConversationHistory(ctx context.Context, conversations []Conversation, requestFiles []File) (*ConversationHistoryResult, error)

// Get the token counter being used
func (cs *ConversationSummarizer) GetTokenCounter() TokenCounter
```
