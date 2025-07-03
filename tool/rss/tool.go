package rss

import (
	"context"
	"strings"
)

// Define tool parameters
type ReadRSSParams struct {
	URL   string `json:"url" description:"RSS feed URL to read"`
	Limit *int   `json:"limit,omitempty" description:"Maximum number of items to return (default: no limit)"`
}

type SearchRSSParams struct {
	URLs     []string `json:"urls" description:"List of RSS feed URLs to search"`
	Query    string   `json:"query" description:"Search query"`
	MaxItems *int     `json:"max_items,omitempty" description:"Maximum items per feed (default: no limit)"`
}

type SearchRSSResult struct {
	Source string   `json:"source" description:"Source URL of the RSS feed"`
	Item   FeedItem `json:"item" description:"Item from the RSS feed"`
}

// Read single RSS feed
func ReadRSS[ctxT context.Context](ctx ctxT, params ReadRSSParams) (reply struct {
	FeedURL string     `json:"feed_url" description:"URL of the RSS feed"`
	Items   []FeedItem `json:"items" description:"RSS feed items"`
	Count   int        `json:"count" description:"Number of items in the RSS feed"`
}, err error) {
	reader := NewRSSReader()
	reply.Items, err = reader.ReadFeed(ctx, params.URL)
	if err != nil {
		return
	}

	limit := 0
	if params.Limit != nil {
		limit = *params.Limit
	}

	// Apply limit
	if limit > 0 && len(reply.Items) > limit {
		reply.Items = reply.Items[:limit]
	}

	reply.FeedURL = params.URL
	reply.Count = len(reply.Items)

	return reply, nil
}

// Search in multiple RSS feeds
func SearchRSS[ctxT context.Context](ctx ctxT, params SearchRSSParams) (reply struct {
	Query   string            `json:"query" description:"Search query"`
	Results []SearchRSSResult `json:"results" description:"List of items from the RSS feeds"`
	Count   int               `json:"count" description:"Number of items in the RSS feeds"`
}, err error) {
	reader := NewRSSReader()

	maxItems := 0
	if params.MaxItems != nil {
		maxItems = *params.MaxItems
	}

	reply.Results = make([]SearchRSSResult, 0, maxItems)
	for _, url := range params.URLs {
		items, err := reader.ReadFeed(ctx, url)
		if err != nil {
			continue // Move to next feed if error occurs
		}

		// Filter by search query
		stopLoop := false
		for _, item := range items {
			if maxItems > 0 && len(reply.Results) >= maxItems {
				stopLoop = true
				break
			}

			if containsQuery(item, params.Query) {
				reply.Results = append(reply.Results, SearchRSSResult{
					Source: url,
					Item:   item,
				})
			}
		}

		if stopLoop {
			break
		}
	}

	reply.Query = params.Query
	reply.Count = len(reply.Results)

	return reply, nil
}

func containsQuery(item FeedItem, query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(item.Title), query) ||
		strings.Contains(strings.ToLower(item.Description), query)
}
