package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mokiat/gog"
	"github.com/pkg/errors"
)

// IndexKnowledge indexes knowledge documents for an agent
func (s *service) IndexKnowledgeFromMap(ctx context.Context, id string, input []map[string]any) (*Knowledge, error) {
	if s.embedder == nil {
		// Return error instead of silently failing - this indicates a configuration issue
		return nil, errors.New("embedder is not available - check OpenAI API key configuration. Knowledge indexing requires a valid OpenAI API key")
	}

	// First, delete existing knowledge for this agent
	if id != "" {
		if err := s.DeleteKnowledge(ctx, id); err != nil {
			return nil, errors.Wrapf(err, "failed to delete existing knowledge")
		}
	}

	knowledge := &Knowledge{
		ID: id,
		Source: Source{
			Title: "Map",
			Type:  SourceTypeMap,
		},
	}

	// Process knowledge into text chunks
	knowledge.Documents = ProcessKnowledgeFromMap(input)
	if len(knowledge.Documents) == 0 {
		return nil, errors.Errorf("no documents found for knowledge %s", id)
	}

	// Extract text content for embedding
	embeddingTexts := make([]string, len(knowledge.Documents))
	for i, chunk := range knowledge.Documents {
		embeddingTexts[i] = chunk.EmbeddingText
	}

	// Generate embeddings
	embeddings, err := s.embedder.Embed(ctx, gog.Map(knowledge.Documents, func(d *Document) string {
		return d.EmbeddingText
	})...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate embeddings")
	}

	if len(embeddings) != len(knowledge.Documents) {
		return nil, errors.Errorf("embedding count mismatch: got %d, expected %d", len(embeddings), len(knowledge.Documents))
	}

	// Create knowledge items for storage
	for i := range knowledge.Documents {
		knowledge.Documents[i].Embeddings = embeddings[i]
	}

	// Store all items
	if err := s.store.Store(ctx, knowledge); err != nil {
		return nil, errors.Wrapf(err, "failed to store knowledge")
	}

	return knowledge, nil
}

// processKnowledge converts knowledge maps into indexable text chunks
func ProcessKnowledgeFromMap(data []map[string]any) []*Document {
	documents := make([]*Document, 0, len(data))
	for _, item := range data {
		// Convert the knowledge item to a searchable text representation
		content := ExtractTextFromMap(item)
		if content == "" {
			continue
		}

		documents = append(documents, &Document{
			Content: Content{
				Type: ContentTypeText,
				Text: content,
			},
			Metadata:      item,
			EmbeddingText: content,
		})
	}

	return documents
}

// extractTextFromKnowledge extracts searchable text from a knowledge map
func ExtractTextFromMap(item map[string]any) string {
	var textParts []string

	// Common text fields to extract (in priority order)
	textFields := []string{"content", "description", "title", "summary", "text", "name"}

	// First, look for standard text fields
	var foundStandardFields []string
	for _, field := range textFields {
		if value, exists := item[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				foundStandardFields = append(foundStandardFields, str)
			}
		}
	}

	// If we found standard text fields, use them
	if len(foundStandardFields) > 0 {
		textParts = foundStandardFields
	} else {
		// If no standard text fields found, try to extract from all string values
		// Sort keys for deterministic ordering
		var keys []string
		for k := range item {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := item[key]
			if str, ok := value.(string); ok && str != "" {
				textParts = append(textParts, fmt.Sprintf("%s: %s", key, str))
			}
		}
	}

	return strings.Join(textParts, " ")
}
