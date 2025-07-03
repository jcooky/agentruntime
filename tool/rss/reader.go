package rss

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

// RSS Tool structure
type RSSReader struct {
	parser *gofeed.Parser
}

// RSS item structure (format to pass to AI Agent)
type FeedItem struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Published   time.Time `json:"published"`
	Author      string    `json:"author,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
}

// Initialize RSS Reader
func NewRSSReader() *RSSReader {
	return &RSSReader{
		parser: gofeed.NewParser(),
	}
}

// Method to read feed
func (r *RSSReader) ReadFeed(ctx context.Context, feedURL string) ([]FeedItem, error) {
	// Set timeout using Context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	feed, err := r.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	items := make([]FeedItem, 0, len(feed.Items))
	for _, item := range feed.Items {
		feedItem := FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Categories:  item.Categories,
		}

		// Parse publication time
		if item.PublishedParsed != nil {
			feedItem.Published = *item.PublishedParsed
		}

		// Author information
		if item.Author != nil {
			feedItem.Author = item.Author.Name
		}

		items = append(items, feedItem)
	}

	return items, nil
}

// Read multiple feeds simultaneously
func (r *RSSReader) ReadMultipleFeeds(ctx context.Context, feedURLs []string) map[string][]FeedItem {
	results := make(map[string][]FeedItem)
	ch := make(chan struct {
		url   string
		items []FeedItem
		err   error
	}, len(feedURLs))

	// Parallel processing with goroutines
	for _, url := range feedURLs {
		go func(feedURL string) {
			items, err := r.ReadFeed(ctx, feedURL)
			ch <- struct {
				url   string
				items []FeedItem
				err   error
			}{url: feedURL, items: items, err: err}
		}(url)
	}

	// Collect results
	for i := 0; i < len(feedURLs); i++ {
		result := <-ch
		if result.err == nil {
			results[result.url] = result.items
		}
	}

	return results
}
