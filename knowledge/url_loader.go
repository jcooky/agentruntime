package knowledge

import (
	"context"
	"fmt"
	"strings"
	"time"

	firecrawl "github.com/mendableai/firecrawl-go"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
)

func (s *service) IndexKnowledgeFromURL(ctx context.Context, id string, inputUrl string, crawlParams firecrawl.CrawlParams) (*Knowledge, error) {
	if s.embedder == nil {
		return nil, errors.New("embedder is not available - check OpenAI API key configuration. Knowledge indexing requires a valid OpenAI API key")
	}

	if s.firecrawlConfig == nil {
		return nil, errors.New("firecrawl config is not available - check FireCrawl configuration")
	}

	// First, delete existing knowledge for this ID
	if id != "" {
		if err := s.DeleteKnowledge(ctx, id); err != nil {
			return nil, errors.Wrapf(err, "failed to delete existing knowledge")
		}
	}

	// Validate FireCrawl configuration
	if err := s.firecrawlConfig.Validate(); err != nil {
		return nil, errors.Wrap(err, "FireCrawl configuration is invalid - check FIRECRAWL_API_KEY environment variable")
	}

	// Initialize FireCrawl client
	client, err := firecrawl.NewFirecrawlApp(s.firecrawlConfig.APIKey, s.firecrawlConfig.APIUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create FireCrawl client")
	}

	// crawlParams is now required - no need to check for nil

	s.logger.Info("Starting to crawl website", "url", inputUrl, "maxDepth", *crawlParams.MaxDepth, "limit", *crawlParams.Limit)
	startTime := time.Now()

	// Start the crawl job and wait for completion
	// CrawlURL is synchronous and will poll internally
	crawlResult, err := client.CrawlURL(inputUrl, &crawlParams, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to crawl URL: %s", inputUrl)
	}

	if crawlResult.Status == "failed" {
		return nil, errors.Errorf("crawl failed for URL: %s", inputUrl)
	}

	s.logger.Info("Crawl completed", "url", inputUrl, "duration", time.Since(startTime), "pagesFound", len(crawlResult.Data))

	if len(crawlResult.Data) == 0 {
		return nil, errors.Errorf("no content retrieved from URL: %s", inputUrl)
	}

	// Create knowledge object
	knowledge := &Knowledge{
		ID: id,
		Source: Source{
			Title: fmt.Sprintf("Website: %s", inputUrl),
			Type:  SourceTypeURL,
			URL:   &inputUrl,
		},
		Metadata: map[string]any{
			"url":          inputUrl,
			"crawled_at":   time.Now().Format(time.RFC3339),
			"pages_count":  len(crawlResult.Data),
			"crawl_depth":  *crawlParams.MaxDepth,
			"total":        crawlResult.Total,
			"completed":    crawlResult.Completed,
			"credits_used": crawlResult.CreditsUsed,
		},
	}

	// Process all crawled pages into documents
	documents := ProcessKnowledgeFromCrawl(crawlResult)
	if len(documents) == 0 {
		return nil, errors.Errorf("no content found at URL: %s", inputUrl)
	}

	knowledge.Documents = documents

	// Generate embeddings for all documents
	s.logger.Info("Generating embeddings for documents", "count", len(knowledge.Documents))

	embeddings, err := s.embedder.Embed(ctx, gog.Map(knowledge.Documents, func(d *Document) string {
		return d.EmbeddingText
	})...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate embeddings")
	}

	if len(embeddings) != len(knowledge.Documents) {
		return nil, errors.Errorf("embedding count mismatch: got %d, expected %d", len(embeddings), len(knowledge.Documents))
	}

	// Assign embeddings to documents
	for i := range knowledge.Documents {
		knowledge.Documents[i].Embeddings = embeddings[i]
	}

	s.logger.Info("Generated embeddings successfully", "count", len(embeddings))

	// Store all items
	if err := s.store.Store(ctx, knowledge); err != nil {
		return nil, errors.Wrapf(err, "failed to store knowledge")
	}

	return knowledge, nil
}

// ProcessKnowledgeFromCrawl converts crawled website content into indexable documents
func ProcessKnowledgeFromCrawl(crawlResult *firecrawl.CrawlStatusResponse) []*Document {
	documents := make([]*Document, 0)

	for pageIdx, page := range crawlResult.Data {
		// Get the page URL and title
		pageUrl := ""
		pageTitle := ""

		if page.Metadata != nil {
			if page.Metadata.SourceURL != nil && *page.Metadata.SourceURL != "" {
				pageUrl = *page.Metadata.SourceURL
			}

			if page.Metadata.Title != nil && *page.Metadata.Title != "" {
				pageTitle = *page.Metadata.Title
			}
		}

		// Get markdown content for embedding text
		markdownContent := page.Markdown
		if markdownContent == "" {
			markdownContent = page.HTML // Fall back to HTML if markdown is not available
		}

		if markdownContent == "" && page.Screenshot == "" {
			continue // Skip pages with no content or screenshot
		}

		// Create a descriptive prefix for context
		contextPrefix := ""
		if pageTitle != "" {
			contextPrefix = fmt.Sprintf("[Page: %s]\n", pageTitle)
		}
		if pageUrl != "" {
			contextPrefix += fmt.Sprintf("[URL: %s]\n\n", pageUrl)
		}

		// If we have a screenshot, create a single document with the screenshot
		if page.Screenshot != "" {
			documentId := fmt.Sprintf("page_%d_screenshot", pageIdx)

			// Use full markdown content for embedding (no chunking needed with screenshots)
			embeddingText := contextPrefix + markdownContent

			document := &Document{
				ID: documentId,
				Content: Content{
					Type:     ContentTypeImage,
					Image:    page.Screenshot, // Base64 encoded screenshot
					MIMEType: "image/png",     // Screenshots are typically PNG
				},
				EmbeddingText: embeddingText,
				Metadata: map[string]any{
					"page_index":     pageIdx,
					"source_url":     pageUrl,
					"page_title":     pageTitle,
					"has_screenshot": true,
					"markdown_text":  markdownContent, // Store the text for reference
				},
			}
			documents = append(documents, document)
		} else {
			// If no screenshot, fall back to text chunking
			chunks := chunkMarkdownContent(markdownContent, 2000) // 2000 chars per chunk

			for chunkIdx, chunk := range chunks {
				documentId := fmt.Sprintf("page_%d_chunk_%d", pageIdx, chunkIdx)

				embeddingText := contextPrefix + chunk

				document := &Document{
					ID: documentId,
					Content: Content{
						Type: ContentTypeText,
						Text: chunk,
					},
					EmbeddingText: embeddingText,
					Metadata: map[string]any{
						"page_index":     pageIdx,
						"chunk_index":    chunkIdx,
						"total_chunks":   len(chunks),
						"source_url":     pageUrl,
						"page_title":     pageTitle,
						"has_screenshot": false,
					},
				}
				documents = append(documents, document)
			}
		}
	}

	return documents
}

// chunkMarkdownContent splits markdown content into smaller chunks
func chunkMarkdownContent(content string, maxChunkSize int) []string {
	if len(content) <= maxChunkSize {
		return []string{content}
	}

	var chunks []string

	// First, try to split by markdown headers (##, ###, etc.)
	sections := splitByMarkdownHeaders(content)

	for _, section := range sections {
		if len(section) <= maxChunkSize {
			chunks = append(chunks, section)
		} else {
			// If a section is still too large, split it further
			subChunks := chunkByParagraphs(section, maxChunkSize)
			chunks = append(chunks, subChunks...)
		}
	}

	// If no chunks were created (no headers found), fall back to paragraph splitting
	if len(chunks) == 0 {
		chunks = chunkByParagraphs(content, maxChunkSize)
	}

	return chunks
}

// splitByMarkdownHeaders splits content by markdown headers
func splitByMarkdownHeaders(content string) []string {
	lines := strings.Split(content, "\n")
	var sections []string
	var currentSection []string

	for _, line := range lines {
		// Check if line is a markdown header
		if strings.HasPrefix(line, "#") && (strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") ||
			strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### ")) {
			// Start a new section
			if len(currentSection) > 0 {
				sections = append(sections, strings.Join(currentSection, "\n"))
				currentSection = []string{}
			}
		}
		currentSection = append(currentSection, line)
	}

	// Add the last section
	if len(currentSection) > 0 {
		sections = append(sections, strings.Join(currentSection, "\n"))
	}

	return sections
}

// chunkByParagraphs splits content by paragraphs while respecting max chunk size
func chunkByParagraphs(content string, maxChunkSize int) []string {
	var chunks []string
	paragraphs := strings.Split(content, "\n\n")

	var currentChunk []string
	currentSize := 0

	for _, paragraph := range paragraphs {
		paragraphSize := len(paragraph)

		// If adding this paragraph would exceed the limit
		if currentSize+paragraphSize+2 > maxChunkSize && len(currentChunk) > 0 {
			// Save current chunk
			chunks = append(chunks, strings.Join(currentChunk, "\n\n"))
			currentChunk = []string{}
			currentSize = 0
		}

		// If a single paragraph is larger than max size, split it further
		if paragraphSize > maxChunkSize {
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.Join(currentChunk, "\n\n"))
				currentChunk = []string{}
				currentSize = 0
			}

			// Split large paragraph by sentences or words
			subChunks := splitLargeParagraph(paragraph, maxChunkSize)
			chunks = append(chunks, subChunks...)
		} else {
			currentChunk = append(currentChunk, paragraph)
			currentSize += paragraphSize + 2 // +2 for "\n\n"
		}
	}

	// Add remaining content
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, "\n\n"))
	}

	return chunks
}

// splitLargeParagraph splits a large paragraph into smaller chunks
func splitLargeParagraph(paragraph string, maxChunkSize int) []string {
	var chunks []string
	runes := []rune(paragraph)

	for i := 0; i < len(runes); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		// Try to find a good break point
		if end < len(runes) {
			// Look for sentence end
			for j := end; j > i+maxChunkSize/2; j-- {
				if j > 0 && (runes[j-1] == '.' || runes[j-1] == '!' || runes[j-1] == '?') &&
					j < len(runes) && (runes[j] == ' ' || runes[j] == '\n') {
					end = j
					break
				}
			}

			// If no sentence break, look for word boundary
			if end == i+maxChunkSize {
				for j := end; j > i+maxChunkSize*3/4; j-- {
					if j > 0 && runes[j-1] != ' ' && runes[j] == ' ' {
						end = j
						break
					}
				}
			}
		}

		chunk := strings.TrimSpace(string(runes[i:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}
