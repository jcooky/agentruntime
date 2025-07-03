package rss_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/habiliai/agentruntime/tool/rss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock RSS feed data
const mockRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>Test RSS feed</description>
    <item>
      <title>Test Item 1</title>
      <link>https://example.com/item1</link>
      <description>This is test item 1</description>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <category>Technology</category>
      <category>AI</category>
    </item>
    <item>
      <title>Test Item 2</title>
      <link>https://example.com/item2</link>
      <description>This is test item 2</description>
      <pubDate>Tue, 02 Jan 2024 12:00:00 GMT</pubDate>
      <category>Programming</category>
    </item>
  </channel>
</rss>`

const invalidRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
  <not-rss>This is not a valid RSS feed</not-rss>
</invalid>`

func TestRSSReader_ReadFeed_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(mockRSSFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	feed, err := reader.ReadFeed(ctx, server.URL)

	require.NoError(t, err)
	assert.Len(t, feed, 2)

	// Verify first item
	assert.Equal(t, "Test Item 1", feed[0].Title)
	assert.Equal(t, "https://example.com/item1", feed[0].Link)
	assert.Equal(t, "This is test item 1", feed[0].Description)

	// Check categories - Categories is []string in universal gofeed parser
	assert.Len(t, feed[0].Categories, 2)
	assert.Equal(t, "Technology", feed[0].Categories[0])
	assert.Equal(t, "AI", feed[0].Categories[1])

	// Check published date
	assert.NotNil(t, feed[0].Published)

	// Verify second item
	assert.Equal(t, "Test Item 2", feed[1].Title)
	assert.Equal(t, "https://example.com/item2", feed[1].Link)
	assert.Equal(t, "This is test item 2", feed[1].Description)

	// Check categories for second item
	assert.Len(t, feed[1].Categories, 1)
	assert.Equal(t, "Programming", feed[1].Categories[0])
}

func TestRSSReader_ReadFeed_InvalidURL(t *testing.T) {
	reader := rss.NewRSSReader()
	ctx := context.Background()

	_, err := reader.ReadFeed(ctx, "invalid-url")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse feed")
}

func TestRSSReader_ReadFeed_InvalidRSSContent(t *testing.T) {
	// Create mock HTTP server with invalid RSS content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(invalidRSSFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	_, err := reader.ReadFeed(ctx, server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse feed")
}

func TestRSSReader_ReadFeed_ServerError(t *testing.T) {
	// Create mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	_, err := reader.ReadFeed(ctx, server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse feed")
}

func TestRSSReader_ReadFeed_ContextTimeout(t *testing.T) {
	// Create mock HTTP server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(mockRSSFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := reader.ReadFeed(ctx, server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestRSSReader_ReadMultipleFeeds_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(mockRSSFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	feedURLs := []string{server.URL, server.URL}
	results := reader.ReadMultipleFeeds(ctx, feedURLs)

	assert.Len(t, results, 1)

	// Both feeds should have same content
	for _, feedURL := range feedURLs {
		feed, exists := results[feedURL]
		assert.True(t, exists)
		assert.Len(t, feed, 2)
		assert.Equal(t, "Test Item 1", feed[0].Title)
		assert.Equal(t, "Test Item 2", feed[1].Title)
	}
}

func TestRSSReader_ReadMultipleFeeds_MixedResults(t *testing.T) {
	// Create mock HTTP server that returns valid RSS
	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(mockRSSFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer validServer.Close()

	// Create mock HTTP server that returns error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal Server Error")); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer errorServer.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	feedURLs := []string{validServer.URL, errorServer.URL}
	results := reader.ReadMultipleFeeds(ctx, feedURLs)

	// Only valid feed should be in results
	assert.Len(t, results, 1)

	feed, exists := results[validServer.URL]
	assert.True(t, exists)
	assert.Len(t, feed, 2)

	// Error feed should not be in results
	_, exists = results[errorServer.URL]
	assert.False(t, exists)
}

func TestRSSReader_ReadMultipleFeeds_EmptyURLs(t *testing.T) {
	reader := rss.NewRSSReader()
	ctx := context.Background()

	results := reader.ReadMultipleFeeds(ctx, []string{})

	assert.Empty(t, results)
}

func TestRSSReader_ReadFeed_EmptyFeed(t *testing.T) {
	// Create mock HTTP server with empty RSS feed
	emptyFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>Empty RSS feed</description>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(emptyFeed)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	reader := rss.NewRSSReader()
	ctx := context.Background()

	feed, err := reader.ReadFeed(ctx, server.URL)

	require.NoError(t, err)
	assert.Empty(t, feed)
}
