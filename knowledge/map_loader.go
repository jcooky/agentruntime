package knowledge

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

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
			Contents: []mcp.Content{
				mcp.NewTextContent(content),
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
