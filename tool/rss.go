package tool

import (
	"strings"
	"text/template"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/tool/rss"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type RSSFeedUrl struct {
	URL         string `json:"url" jsonschema:"description=The URL of the RSS feed"`
	Name        string `json:"name" jsonschema:"description=The name of the RSS feed"`
	Description string `json:"description" jsonschema:"description=The description of the RSS feed"`
}

var (
	searchRSSDescriptionTmpl = template.Must(template.New("search_rss_description").Parse(`Search multiple RSS feeds for items matching a query keyword.

## Allowed RSS Feeds
You can ONLY search through the following pre-configured RSS feeds:
<allowed_rss_feeds>
{{- if .AllowedFeedUrls}}
{{- range .AllowedFeedUrls}}
- [**{{.Name}}**]({{.URL}}): {{.Description}}
{{- end}}
{{- else}}
⚠️ No RSS feeds have been configured for this agent. Please contact the administrator to add RSS feed sources.
{{- end}}
</allowed_rss_feeds>

## How to select feeds
1. **Review the list above** - Each feed shows its name, URL, and what content it provides
2. **Match content to query** - Choose feeds whose descriptions match your search topic
3. **Use exact URLs** - Copy the URLs exactly as shown in the allowed feeds list

## Purpose
- Find specific content across allowed RSS feeds simultaneously
- Filter articles, news, or blog posts by keywords in titles or descriptions
- Monitor pre-configured content sources for relevant information

## When to use
- When you need to find recent articles or posts about a specific topic from allowed sources
- To search across multiple pre-configured news sources, blogs, or content feeds
- When monitoring allowed RSS feeds for mentions of particular keywords or subjects
- To gather content from approved sources for research or analysis

## How it works
- Select RSS feed URLs from the allowed feeds list
- Provide a search query to filter content
- The tool searches through titles and descriptions of all feed items
- Returns matching items along with their source URLs
- Case-insensitive matching across title and description fields
- Processes feeds in parallel for faster results (30-second timeout per feed)

## Parameters
- **urls**: Array of RSS feed URLs to search *(required)*
  - ⚠️ **MUST use URLs from the "Allowed RSS Feeds" list above**
  - Copy URLs exactly as shown (e.g., if the list shows "https://example.com/feed/", use exactly that)
  - You can include multiple URLs to search across several feeds simultaneously
  - Example: If searching for AI news and you have TechCrunch and Ars Technica in your allowed feeds, use both URLs
- **query**: Search keyword or phrase to match *(required)*
  - Searches in both titles and descriptions (case-insensitive)
  - Example: "artificial intelligence", "ChatGPT", "machine learning"
  - Tip: Use specific terms for precise results, or broader terms to catch more articles
- **max_items**: Maximum number of results to return *(optional)*
  - Limits total results across all searched feeds
  - Useful when searching high-volume feeds to avoid overwhelming output
  - Default: no limit (returns all matching items)

## Output format
Returns a JSON object containing:
- **query**: The search query used
- **results**: Array of matching items, each containing:
  - **source**: The RSS feed URL where the item was found
  - **item**: Object with title, description, link, published date, author, and categories
- **count**: Total number of matching items found

## Best practices
- Review the allowed feeds list and their descriptions before searching
- Select feeds that are most likely to contain content about your query topic
- Use specific keywords for better results
- Include multiple relevant allowed feeds for comprehensive coverage
- Set max_items to limit results when dealing with high-volume feeds
- Consider using broader terms if initial search returns no results
- If unsure which feeds to use, consider the feed descriptions and names

## Common use cases
1. **News monitoring**: Search allowed tech news feeds for product launches or company updates
2. **Research**: Gather blog posts about specific technical topics from approved sources
3. **Competitor tracking**: Monitor allowed industry feeds for competitor mentions
4. **Content curation**: Find relevant articles from trusted sources for newsletters or summaries

## Error handling
- If you use a URL not in the allowed list, the search may fail or return no results
- Always verify that the URLs you're using are from the configured allowed feeds`))

	readRSSDescriptionTmpl = template.Must(template.New("read_rss_description").Parse(`Read and retrieve all recent items from a single RSS feed.

## Allowed RSS Feeds
You can ONLY read from the following pre-configured RSS feeds:
<allowed_rss_feeds>
{{- if .AllowedFeedUrls}}
{{- range .AllowedFeedUrls}}
- [**{{.Name}}**]({{.URL}}): {{.Description}}
{{- end}}
{{- else}}
⚠️ No RSS feeds have been configured for this agent. Please contact the administrator to add RSS feed sources.
{{- end}}
</allowed_rss_feeds>

## Purpose
- Fetch all recent articles/posts from a single RSS feed
- Get a complete overview of what's currently published in a specific feed
- Monitor the latest content from a trusted source without filtering

## When to use
- When you need to see all recent content from a specific news source or blog
- To get an overview of what topics are currently being covered by a feed
- When you want the complete feed content before deciding what to focus on
- To check for new updates from a specific source

## How it works
- Provide the URL of a single RSS feed from the allowed list
- The tool fetches and parses the RSS feed content
- Returns all feed items with their metadata (title, description, link, date, etc.)
- Optionally limit the number of items returned

## Parameters
- **url**: RSS feed URL to read *(required)*
  - ⚠️ **MUST be a URL from the "Allowed RSS Feeds" list above**
  - Copy the URL exactly as shown in the allowed feeds list
  - Only one URL can be read at a time (use search_rss for multiple feeds)
- **limit**: Maximum number of items to return *(optional)*
  - Useful for feeds with many items to avoid overwhelming output
  - Returns the most recent items up to the limit
  - Default: no limit (returns all items in the feed)

## Output format
Returns a JSON object containing:
- **feed_url**: The URL of the RSS feed that was read
- **items**: Array of all feed items, each containing:
  - **title**: Article/post title
  - **description**: Summary or full content
  - **link**: URL to the full article
  - **published**: Publication date and time
  - **author**: Author name (if available)
  - **categories**: Array of category tags (if available)
- **count**: Total number of items returned

## Best practices
- Check the feed description to ensure it matches your information needs
- Use the limit parameter for feeds known to have many items
- For topic-specific content, use search_rss instead to filter by keywords
- Consider reading multiple feeds sequentially if you need comprehensive coverage

## Common use cases
1. **Daily news briefing**: Read a trusted news source to see all recent articles
2. **Blog monitoring**: Check what new posts have been published on a specific blog
3. **Feed overview**: Get a sense of what topics are trending in a particular source
4. **Content aggregation**: Collect all items from a feed for further processing

## Differences from search_rss
- **read_rss**: Gets ALL items from ONE feed (no filtering)
- **search_rss**: Searches for SPECIFIC items across MULTIPLE feeds (with keyword filtering)

## Error handling
- If the URL is not in the allowed list, the operation will fail
- Network timeouts after 30 seconds per feed
- Invalid or inaccessible feeds will return an error`))
)

func (m *manager) registerRSSSkill(skill *entity.NativeAgentSkill) error {
	var allowedFeedUrls []RSSFeedUrl
	if err := mapstructure.Decode(skill.Env["allowed_feed_urls"], &allowedFeedUrls); err != nil {
		return errors.WithStack(err)
	}

	{
		description := strings.Builder{}
		if err := searchRSSDescriptionTmpl.Execute(&description, struct {
			AllowedFeedUrls []RSSFeedUrl
		}{
			AllowedFeedUrls: allowedFeedUrls,
		}); err != nil {
			return errors.WithStack(err)
		}

		registerNativeTool(
			m,
			"search_rss",
			description.String(),
			skill,
			rss.SearchRSS[*Context],
		)
	}

	{
		description := strings.Builder{}
		if err := readRSSDescriptionTmpl.Execute(&description, struct {
			AllowedFeedUrls []RSSFeedUrl
		}{
			AllowedFeedUrls: allowedFeedUrls,
		}); err != nil {
			return errors.WithStack(err)
		}

		registerNativeTool(
			m,
			"read_rss",
			description.String(),
			skill,
			rss.ReadRSS[*Context],
		)
	}

	return nil
}
