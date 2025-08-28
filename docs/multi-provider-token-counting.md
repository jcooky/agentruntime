# Multi-Provider Token Counting

AgentRuntime supports optimized token calculation for various AI model providers. Since each provider uses different tokenization algorithms, we provide dedicated implementations for accurate token calculation per provider.

## Supported Providers

| Provider      | Token Calculation Method | Supported Models              |
| ------------- | ------------------------ | ----------------------------- |
| **OpenAI**    | tiktoken (cl100k_base)   | GPT-4, GPT-3.5, GPT-4o series |
| **Anthropic** | count_tokens API         | Claude 3, Claude 4 series     |

## TokenCounter Interface

```go
type TokenCounter interface {
    CountTokens(text string) int
    CountFileTokens(contentType, data string) int
    CountConversationTokens(conversations []Conversation) int
    CountRequestFilesTokens(files []File) int
    ProviderName() string
}
```

## Usage Methods

### 1. Auto-Detection (Recommended)

The simplest approach is to automatically select the appropriate token counter based on the model name:

```go
cfg := config.ConversationSummaryConfig{
    MaxTokens:       10000,
    SummaryTokens:   2000,
    ModelForSummary: "gpt-4o-mini",    // Automatically selects OpenAI counter
    TokenProvider:   "auto",           // Auto-detection
}

summarizer, err := engine.NewConversationSummarizer(genkit, cfg)
```

### 2. Explicit Provider Specification

You can explicitly specify a particular provider:

```go
cfg := config.ConversationSummaryConfig{
    MaxTokens:       10000,
    SummaryTokens:   2000,
    ModelForSummary: "claude-3-5-sonnet",
    TokenProvider:   "anthropic",      // Explicitly specify Anthropic counter
}
```

### 3. Direct Token Counter Creation

For more granular control, you can create token counters directly:

```go
// OpenAI token counter
openaiCounter, err := engine.NewOpenAITokenCounter()
if err != nil {
    log.Fatal(err)
}

// Anthropic token counter
anthropicCounter, err := engine.NewAnthropicTokenCounter("claude-3-5-sonnet-20241022")
if err != nil {
    log.Fatal(err)
}

// Create summarizer with custom token counter
summarizer, err := engine.NewConversationSummarizerWithTokenCounter(genkit, cfg, openaiCounter)
```

## Provider-Specific Features

### OpenAI TokenCounter

- **Tokenization**: Uses tiktoken's `cl100k_base` encoding
- **Image Tokens**: Applies OpenAI Vision model's official algorithm
- **Advantages**:
  - Fast local computation
  - No API calls required
  - Accurate GPT model token calculation

```go
counter, err := engine.NewOpenAITokenCounter()
textTokens := counter.CountTokens("Hello, world!")
imageTokens := counter.CountFileTokens("image/jpeg", base64Data)
```

### Anthropic TokenCounter

- **Tokenization**: Uses Anthropic's `count_tokens` API
- **Image Tokens**: Real token calculation for Claude Vision models
- **Advantages**:
  - Most accurate token calculation for Claude models
  - Same tokenization logic as actual models
- **Considerations**:
  - API key required: `ANTHROPIC_API_KEY` environment variable
  - API call latency
  - API call costs to consider

```go
// ANTHROPIC_API_KEY environment variable required
counter, err := engine.NewAnthropicTokenCounter("claude-3-5-sonnet-20241022")
textTokens := counter.CountTokens("Hello, world!")
```

## Token Factory Usage

Automatic token counter selection based on model name:

```go
factory := engine.NewDefaultTokenCounterFactory()

// Auto-selection by model name
counter, err := factory.CreateTokenCounterForModel("gpt-4o-mini")  // OpenAI
counter, err := factory.CreateTokenCounterForModel("claude-3-5-sonnet")  // Anthropic

// Direct selection by provider name
counter, err := factory.CreateTokenCounter("openai")
counter, err := factory.CreateTokenCounter("anthropic")
```

## AgentRuntime Configuration

Setting up token counting providers when initializing AgentRuntime:

```go
// Using auto-detection (recommended)
runtime, err := agentruntime.NewAgentRuntime(
    agentruntime.WithConversationSummary(config.ConversationSummaryConfig{
        MaxTokens:       100000,
        ModelForSummary: "gpt-4o-mini",
        TokenProvider:   "auto",  // Auto-select based on model
    }),
)

// Explicitly specifying Anthropic provider
runtime, err := agentruntime.NewAgentRuntime(
    agentruntime.WithConversationSummary(config.ConversationSummaryConfig{
        MaxTokens:       100000,
        ModelForSummary: "claude-3-5-sonnet",
        TokenProvider:   "anthropic",
    }),
)
```

## Environment Setup

### OpenAI

- No additional setup required (uses local tiktoken)

### Anthropic

```bash
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

## Performance Considerations

### OpenAI TokenCounter

- ✅ **Fast**: Local computation
- ✅ **Free**: No API calls
- ✅ **Offline**: No internet connection required

### Anthropic TokenCounter

- ⚠️ **Slower**: API calls required
- ⚠️ **Cost**: Per-API-call charges
- ⚠️ **Online**: Internet connection required
- ✅ **Accurate**: Same tokenization as Claude models

## Fallback Mechanism

Falls back to estimation when API calls fail:

```go
// Automatically returns estimated value if Anthropic API fails
tokens := anthropicCounter.CountTokens("text")
// API failure → EstimateTokens("text") called
```

## Example Code

See the complete example in `examples/multi_provider_token_counting.go`:

```bash
go run ./examples/multi_provider_token_counting.go
```

## Model Mapping

| Model Pattern | Provider                | Examples                         |
| ------------- | ----------------------- | -------------------------------- |
| `gpt*`        | openai                  | gpt-4, gpt-3.5-turbo, gpt-4o     |
| `claude*`     | anthropic               | claude-3-5-sonnet, claude-4-opus |
| `grok*`       | xai (openai compatible) | grok-1                           |
| Others        | openai (default)        | Unknown models                   |

This multi-provider token counting enables accurate token management tailored to each AI model's characteristics.
