package knowledge

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/mokiat/gog"
)

type (
	Knowledge struct {
		ID        string         `json:"id"`
		Source    Source         `json:"source"`
		Metadata  map[string]any `json:"metadata"`
		Documents []*Document    `json:"documents"`
	}

	Source struct {
		Title    string     `json:"title"`
		URL      *string    `json:"url"`
		Filename *string    `json:"filename"`
		Type     SourceType `json:"type"`
	}

	SourceType string

	Document struct {
		ID            string         `json:"id"`
		Content       Content        `json:"content"`
		Embeddings    []float32      `json:"embeddings"`
		EmbeddingText string         `json:"embeddingText"`
		Metadata      map[string]any `json:"metadata"`
	}

	KnowledgeSearchResult struct {
		*Document `json:",inline"`
		Score     float32 `json:"score"`
	}

	Content struct {
		Type string `json:"type"`

		Text     string `json:"text,omitempty"`
		Image    string `json:"data,omitempty"`
		MIMEType string `json:"mimeType,omitempty"`
	}
)

const (
	SourceTypeMap = "map"
	SourceTypePDF = "pdf"
	SourceTypeURL = "url"

	ContentTypeText  = "text"
	ContentTypeImage = "image"
)

func (d *Document) ToDoc() (*ai.Document, error) {
	doc := &ai.Document{
		Metadata: gog.Merge(d.Metadata, map[string]any{
			"id": d.ID,
		}),
	}

	switch d.Content.Type {
	case ContentTypeText:
		doc.Content = append(doc.Content, ai.NewTextPart(d.Content.Text))
	case ContentTypeImage:
		doc.Content = append(doc.Content, ai.NewMediaPart(d.Content.MIMEType, d.Content.Image))
	default:
		return nil, fmt.Errorf("unknown content type: %s", d.Content.Type)
	}

	return doc, nil
}
