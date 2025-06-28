# Knowledge Package

The knowledge package provides a flexible and extensible system for storing and retrieving knowledge with semantic search capabilities.

## Architecture

The package is designed with a clean separation of concerns:

- **KnowledgeService**: Handles business logic (text processing, embedding generation, reranking)
- **Store**: Interface for storage operations (can be implemented by different backends)
- **SqliteStore**: Default implementation using SQLite with sqlite-vec extension

## Usage

### Using Default SQLite Store

```go
import "github.com/habiliai/agentruntime/knowledge"

// Create service with default SQLite store
service, err := knowledge.NewService(ctx, config, logger, genkit)
```

### Using Custom Store

```go
// Implement your own Store
type MyCustomStore struct {
    // your implementation
}

func (s *MyCustomStore) Store(ctx context.Context, items []KnowledgeItem) error {
    // your implementation
}

func (s *MyCustomStore) Search(ctx context.Context, agentName string, queryEmbedding []float32, limit int) ([]KnowledgeSearchResult, error) {
    // your implementation
}

// ... implement other methods

// Use custom store
customStore := &MyCustomStore{}
service, err := knowledge.NewServiceWithStore(ctx, config, logger, genkit, customStore)

// Or with AgentRuntime
runtime := agentruntime.New(
    agentruntime.WithStore(customStore),
)
```

## Store Interface

The `Store` interface defines the contract for any knowledge storage backend:

```go
type Store interface {
    // Store knowledge items atomically
    Store(ctx context.Context, items []KnowledgeItem) error

    // Perform semantic search
    Search(ctx context.Context, agentName string, queryEmbedding []float32, limit int) ([]KnowledgeSearchResult, error)

    // Get all knowledge for an agent
    GetByAgent(ctx context.Context, agentName string) ([]KnowledgeItem, error)

    // Delete all knowledge for an agent
    DeleteByAgent(ctx context.Context, agentName string) error

    // Close the store
    Close() error
}
```

## Implementing Custom Stores

To implement a custom knowledge store (e.g., for Pinecone, Qdrant, Weaviate):

1. Implement the `Store` interface
2. Handle both metadata and vector storage
3. Ensure atomic operations in `Store`
4. Convert similarity scores appropriately in `Search`

Example implementations can be found in:

- `sqlite_store.go` - SQLite with sqlite-vec extension
