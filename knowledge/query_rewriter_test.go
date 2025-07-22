package knowledge

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpQueryRewriter(t *testing.T) {
	rewriter := NewNoOpQueryRewriter()
	ctx := t.Context()

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "simple query",
			query: "How does Redis handle persistence?",
		},
		{
			name:  "complex query",
			query: "What are the differences between RDB and AOF in Redis?",
		},
		{
			name:  "empty query",
			query: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := rewriter.Rewrite(ctx, tt.query)
			require.NoError(t, err)
			assert.Len(t, results, 1)
			assert.Equal(t, tt.query, results[0])
		})
	}
}

func TestMultiStrategyRewriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create multiple mock rewriters
	rewriter1 := &mockRewriter{
		responses: map[string][]string{
			"test query": {"test query", "rewritten1"},
		},
	}
	rewriter2 := &mockRewriter{
		responses: map[string][]string{
			"test query": {"test query", "rewritten2", "rewritten3"},
		},
	}

	multiRewriter := NewMultiStrategyRewriter(rewriter1, rewriter2)
	ctx := context.Background()

	results, err := multiRewriter.Rewrite(ctx, "test query")
	require.NoError(t, err)

	// Should have all unique results
	assert.Len(t, results, 4) // original + 3 rewritten
	assert.Contains(t, results, "test query")
	assert.Contains(t, results, "rewritten1")
	assert.Contains(t, results, "rewritten2")
	assert.Contains(t, results, "rewritten3")
}

func TestMultiStrategyRewriterDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create rewriters that return duplicate results
	rewriter1 := &mockRewriter{
		responses: map[string][]string{
			"test": {"test", "duplicate", "unique1"},
		},
	}
	rewriter2 := &mockRewriter{
		responses: map[string][]string{
			"test": {"test", "duplicate", "unique2"},
		},
	}

	multiRewriter := NewMultiStrategyRewriter(rewriter1, rewriter2)
	ctx := context.Background()

	results, err := multiRewriter.Rewrite(ctx, "test")
	require.NoError(t, err)

	// Should deduplicate results
	assert.Len(t, results, 4) // test, duplicate, unique1, unique2

	// Count occurrences
	counts := make(map[string]int)
	for _, r := range results {
		counts[r]++
	}

	// Each should appear exactly once
	assert.Equal(t, 1, counts["test"])
	assert.Equal(t, 1, counts["duplicate"])
	assert.Equal(t, 1, counts["unique1"])
	assert.Equal(t, 1, counts["unique2"])
}

func TestCreateQueryRewriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		strategy string
		wantType string
	}{
		{
			strategy: "none",
			wantType: "*knowledge.NoOpQueryRewriter",
		},
		{
			strategy: "",
			wantType: "*knowledge.NoOpQueryRewriter",
		},
		{
			strategy: "unknown",
			wantType: "*knowledge.NoOpQueryRewriter",
		},
		// Note: We can't test HyDE, expansion, and multi without a real genkit instance
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			rewriter := CreateQueryRewriter(nil, tt.strategy, "")
			actualType := reflect.TypeOf(rewriter).String()
			assert.Equal(t, tt.wantType, actualType)
		})
	}
}

// mockRewriter for testing
type mockRewriter struct {
	responses map[string][]string
	err       error
}

func (m *mockRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	if results, ok := m.responses[query]; ok {
		return results, nil
	}
	return []string{query}, nil
}
