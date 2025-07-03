package rss_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/habiliai/agentruntime/tool/rss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock RSS feed data for tool tests
const mockRSSFeedForTool = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Tech News</title>
    <link>https://technews.com</link>
    <description>Latest technology news</description>
    <item>
      <title>AI Breakthrough in Machine Learning</title>
      <link>https://technews.com/ai-breakthrough</link>
      <description>Revolutionary AI technology shows promise in machine learning applications</description>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <category>AI</category>
    </item>
    <item>
      <title>New Programming Language Released</title>
      <link>https://technews.com/new-language</link>
      <description>Developers are excited about the new programming language features</description>
      <pubDate>Tue, 02 Jan 2024 12:00:00 GMT</pubDate>
      <category>Programming</category>
    </item>
    <item>
      <title>Database Performance Optimization</title>
      <link>https://technews.com/database-perf</link>
      <description>Learn how to optimize database performance for better results</description>
      <pubDate>Wed, 03 Jan 2024 12:00:00 GMT</pubDate>
      <category>Database</category>
    </item>
  </channel>
</rss>`

const mockRSSFeedForTool2 = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Science Daily</title>
    <link>https://sciencedaily.com</link>
    <description>Daily science news</description>
    <item>
      <title>Machine Learning in Healthcare</title>
      <link>https://sciencedaily.com/ml-healthcare</link>
      <description>How machine learning is transforming healthcare industry</description>
      <pubDate>Thu, 04 Jan 2024 12:00:00 GMT</pubDate>
      <category>Healthcare</category>
    </item>
    <item>
      <title>Climate Change Research</title>
      <link>https://sciencedaily.com/climate</link>
      <description>Latest findings in climate change research and solutions</description>
      <pubDate>Fri, 05 Jan 2024 12:00:00 GMT</pubDate>
      <category>Climate</category>
    </item>
  </channel>
</rss>`

func TestReadRSS_Success(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool))
	}))
	defer server.Close()

	ctx := context.Background()
	params := rss.ReadRSSParams{
		URL: server.URL,
	}

	result, err := rss.ReadRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, server.URL, result.FeedURL)
	assert.Equal(t, 3, result.Count)
	assert.Len(t, result.Items, 3)
	assert.Equal(t, "AI Breakthrough in Machine Learning", result.Items[0].Title)
	assert.Equal(t, "New Programming Language Released", result.Items[1].Title)
	assert.Equal(t, "Database Performance Optimization", result.Items[2].Title)
}

func TestReadRSS_WithLimit(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool))
	}))
	defer server.Close()

	ctx := context.Background()
	limit := 2
	params := rss.ReadRSSParams{
		URL:   server.URL,
		Limit: &limit,
	}

	result, err := rss.ReadRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, server.URL, result.FeedURL)
	assert.Equal(t, 2, result.Count)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, "AI Breakthrough in Machine Learning", result.Items[0].Title)
	assert.Equal(t, "New Programming Language Released", result.Items[1].Title)
}

func TestReadRSS_Error(t *testing.T) {
	ctx := context.Background()
	params := rss.ReadRSSParams{
		URL: "invalid-url",
	}

	_, err := rss.ReadRSS(ctx, params)

	assert.Error(t, err)
}

func TestSearchRSS_Success(t *testing.T) {
	// Create mock HTTP servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool2))
	}))
	defer server2.Close()

	ctx := context.Background()
	params := rss.SearchRSSParams{
		URLs:  []string{server1.URL, server2.URL},
		Query: "machine learning",
	}

	result, err := rss.SearchRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "machine learning", result.Query)
	assert.Equal(t, 2, result.Count)
	assert.Len(t, result.Results, 2)

	// Check that both matching items are found
	foundTitles := make([]string, 0)
	for _, item := range result.Results {
		foundTitles = append(foundTitles, item.Item.Title)
	}
	assert.Contains(t, foundTitles, "AI Breakthrough in Machine Learning")
	assert.Contains(t, foundTitles, "Machine Learning in Healthcare")
}

func TestSearchRSS_WithMaxItems(t *testing.T) {
	// Create mock HTTP servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool2))
	}))
	defer server2.Close()

	ctx := context.Background()
	maxItems := 1
	params := rss.SearchRSSParams{
		URLs:     []string{server1.URL, server2.URL},
		Query:    "machine learning",
		MaxItems: &maxItems,
	}

	result, err := rss.SearchRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "machine learning", result.Query)
	assert.Equal(t, 1, result.Count)
	assert.Len(t, result.Results, 1)
}

func TestSearchRSS_NoMatches(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSFeedForTool))
	}))
	defer server.Close()

	ctx := context.Background()
	params := rss.SearchRSSParams{
		URLs:  []string{server.URL},
		Query: "nonexistent topic",
	}

	result, err := rss.SearchRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "nonexistent topic", result.Query)
	assert.Equal(t, 0, result.Count)
	assert.Empty(t, result.Results)
}

func TestSearchRSS_ErrorFeed(t *testing.T) {
	// Create mock HTTP server that returns error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer errorServer.Close()

	ctx := context.Background()
	params := rss.SearchRSSParams{
		URLs:  []string{errorServer.URL},
		Query: "machine learning",
	}

	result, err := rss.SearchRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "machine learning", result.Query)
	assert.Equal(t, 0, result.Count)
	assert.Empty(t, result.Results)
}

func TestSearchRSS_EmptyURLs(t *testing.T) {
	ctx := context.Background()
	params := rss.SearchRSSParams{
		URLs:  []string{},
		Query: "machine learning",
	}

	result, err := rss.SearchRSS(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, "machine learning", result.Query)
	assert.Equal(t, 0, result.Count)
	assert.Empty(t, result.Results)
}
