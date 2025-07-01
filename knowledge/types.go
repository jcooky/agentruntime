package knowledge

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
)

type (
	Knowledge struct {
		ID        string         `json:"id"`
		Source    Source         `json:"source,omitzero"`
		Metadata  map[string]any `json:"metadata,omitzero"`
		Documents []*Document    `json:"documents,omitzero"`
	}

	Source struct {
		Title    string     `json:"title,omitempty"`
		URL      *string    `json:"url,omitempty"`
		Filename *string    `json:"filename,omitempty"`
		Type     SourceType `json:"type,omitempty"`
	}

	SourceType string

	Document struct {
		ID            string         `json:"id,omitzero"`
		Contents      []mcp.Content  `json:"contents,omitempty"`
		Embeddings    []float32      `json:"embeddings,omitempty"`
		EmbeddingText string         `json:"embedding_text,omitempty"`
		Metadata      map[string]any `json:"metadata,omitempty"`
	}

	KnowledgeSearchResult struct {
		*Document `json:",inline"`
		Score     float32 `json:"score,omitzero"`
	}

	ContentStorage struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Data     string `json:"data,omitempty"`
		MIMEType string `json:"mimeType,omitempty"`
		URI      string `json:"uri,omitempty"`
		Blob     string `json:"blob,omitempty"`
	}
)

const (
	SourceTypeMap = "map"
	SourceTypePDF = "pdf"
)

func (c *ContentStorage) FromContent(content mcp.Content) {
	switch v := content.(type) {
	case mcp.TextContent:
		c.Type = v.Type
		c.Text = v.Text
	case mcp.ImageContent:
		c.Type = v.Type
		c.Data = v.Data
		c.MIMEType = v.MIMEType
	case mcp.AudioContent:
		c.Type = v.Type
		c.Data = v.Data
		c.MIMEType = v.MIMEType
	default:
		panic(errors.Errorf("unknown content type: %T", c))
	}
}

func (c *ContentStorage) ToContent() mcp.Content {
	switch c.Type {
	case "text":
		return mcp.TextContent{Type: c.Type, Text: c.Text}
	case "image":
		return mcp.ImageContent{Type: c.Type, Data: c.Data, MIMEType: c.MIMEType}
	case "audio":
		return mcp.AudioContent{Type: c.Type, Data: c.Data, MIMEType: c.MIMEType}
	default:
		panic(errors.Errorf("unknown content type: %s", c.Type))
	}
}

func (d *Document) ToDoc() (*ai.Document, error) {
	doc := &ai.Document{
		Metadata: gog.Merge(d.Metadata, map[string]any{
			"id": d.ID,
		}),
	}

	for _, content := range d.Contents {
		switch c := content.(type) {
		case mcp.TextContent:
			doc.Content = append(doc.Content, ai.NewTextPart(c.Text))
		case mcp.ImageContent:
			doc.Content = append(doc.Content, ai.NewMediaPart(c.MIMEType, c.Data))
		case mcp.AudioContent:
			doc.Content = append(doc.Content, ai.NewMediaPart(c.MIMEType, c.Data))
		default:
			return nil, errors.Errorf("unknown content type: %T", c)
		}
	}

	return doc, nil
}
