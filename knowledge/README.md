# Knowledge Package

The knowledge package provides a flexible and extensible knowledge management system with built-in RAG (Retrieval-Augmented Generation) capabilities.

## Features

- **Flexible Storage**: Pluggable Store interface allows custom implementations
- **Vector Search**: Built-in embedding and similarity search
- **Reranking**: Optional LLM-based reranking for improved relevance
- **Query Rewriting**: Multiple strategies to improve search accuracy
- **Agent Isolation**: Each agent has its own knowledge namespace
- **Batch Processing**: Efficient batch operations for embeddings and reranking

## Architecture

```
┌─────────────────┐
│ KnowledgeService │
└────────┬────────┘
         │
    ┌────┴─────┬──────────┬───────────┐
    │          │          │           │
┌───▼──┐  ┌───▼────┐ ┌───▼────┐ ┌────▼──────┐
│ Store │  │Embedder│ │Reranker│ │QueryRewriter│
└───────┘  └────────┘ └────────┘ └────────────┘
```

## Store Interface

The Store interface defines the contract for knowledge storage:

```go
type Store interface {
    Store(ctx context.Context, agentName string, items []KnowledgeItem) error
    Search(ctx context.Context, agentName string, embedding []float32, limit int) ([]KnowledgeSearchResult, error)
    GetByAgent(ctx context.Context, agentName string) ([]KnowledgeItem, error)
    DeleteByAgent(ctx context.Context, agentName string) error
}
```

### Built-in Implementation

- **SqliteStore**: Production-ready SQLite-based storage with vector search capabilities

### Custom Implementation

To create a custom Store:

```go
type MyCustomStore struct {
    // your fields
}

func (s *MyCustomStore) Store(ctx context.Context, agentName string, items []KnowledgeItem) error {
    // implementation
}

// ... implement other methods

// Use with KnowledgeService
store := &MyCustomStore{}
service := knowledge.NewServiceWithStore(ctx, config, logger, genkit, store)
```

## Query Rewriting

Query rewriting improves search accuracy by transforming user queries into more search-friendly formats.

### Available Strategies

#### 1. HyDE (Hypothetical Document Embeddings)

Generates a hypothetical answer to the query and searches using both the original query and the generated answer.

**Best for**: Questions where the answer format differs significantly from the question.

**Example**:

- Query: "How does Redis handle persistence?"
- Generated: "Redis handles persistence through two main mechanisms: RDB snapshots and AOF..."
- Result: Better matches with technical documentation

#### 2. Query Expansion

Expands queries with synonyms and related terms to catch documents using different terminology.

**Best for**: Technical topics with multiple terms for the same concept.

**Example**:

- Query: "Python debugging"
- Expanded: "Python debugging pdb breakpoint stack trace debugger"
- Result: Finds documents mentioning any debugging-related terms

#### 3. Multi-Strategy

Combines both HyDE and Query Expansion for comprehensive coverage.

**Best for**: Critical queries where recall is more important than cost.

#### 4. None

Disables query rewriting, using only the original query.

**Best for**: When queries are already well-formed or to minimize costs.

### Configuration

```yaml
knowledge:
  # Enable query rewriting
  queryRewriteEnabled: true

  # Choose strategy: "hyde", "expansion", "multi", "none"
  queryRewriteStrategy: hyde

  # LLM model for rewriting (defaults to rerankModel if not set)
  queryRewriteModel: gpt-4o-mini
```

### Custom Query Rewriter

Implement the QueryRewriter interface:

```go
type MyRewriter struct{}

func (r *MyRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
    // Your rewriting logic
    return []string{query, rewrittenQuery}, nil
}
```

## Reranking

Reranking uses an LLM to score and reorder search results based on relevance to the query.

### Configuration

```yaml
knowledge:
  # Enable reranking
  rerankEnabled: true

  # LLM model for reranking
  rerankModel: gpt-4o-mini

  # Number of final results
  rerankTopK: 5

  # Retrieve N times more candidates for reranking
  retrievalFactor: 3

  # Use batch processing (more efficient)
  useBatchRerank: true
```

## Usage Example

```go
// Create service with default SQLite store
service, err := knowledge.NewService(ctx, config, logger, genkit)

// Or with custom store
customStore := NewMyCustomStore()
service, err := knowledge.NewServiceWithStore(ctx, config, logger, genkit, customStore)

// Store knowledge
items := []knowledge.KnowledgeItem{
    {Content: "Redis persistence uses RDB and AOF"},
    {Content: "Python debugging with pdb"},
}
err = service.StoreKnowledge(ctx, "myAgent", items)

// Retrieve relevant knowledge
results, err := service.RetrieveRelevantKnowledge(ctx, "myAgent", "How to debug Python?", 5)
```

## Best Practices

1. **Choose the Right Strategy**:

   - Use HyDE for question-answering scenarios
   - Use Expansion for technical documentation
   - Use Multi for critical applications

2. **Balance Cost vs Quality**:

   - Query rewriting adds LLM calls
   - Reranking adds additional processing
   - Monitor token usage

3. **Test Different Configurations**:

   - Try different strategies with your data
   - Measure retrieval quality
   - Optimize based on your use case

4. **Custom Stores**:
   - Implement atomic operations
   - Handle concurrent access
   - Consider caching for performance
