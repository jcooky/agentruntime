package knowledge

import (
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/mokiat/gog"
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
		Content       Content        `json:"contents,omitempty"`
		Embeddings    []float32      `json:"embeddings,omitempty"`
		EmbeddingText string         `json:"embedding_text,omitempty"`
		Metadata      map[string]any `json:"metadata,omitempty"`
	}

	KnowledgeSearchResult struct {
		*Document `json:",inline"`
		Score     float32 `json:"score,omitzero"`
	}

	Content struct {
		Type string `json:"type,omitempty"`

		Text     string `json:"text,omitempty"`
		Image    string `json:"data,omitempty"`
		MIMEType string `json:"mimeType,omitempty"`
	}
)

const (
	SourceTypeMap = "map"
	SourceTypePDF = "pdf"

	ContentTypeText  = "text"
	ContentTypeImage = "image"
)

func (d *Document) ToDoc() (*ai.Document, error) {
	doc := &ai.Document{
		Metadata: gog.Merge(d.Metadata, map[string]any{
			"id": d.ID,
		}),
	}

	if d.Content.Type == ContentTypeText {
		doc.Content = append(doc.Content, ai.NewTextPart(d.Content.Text))
	} else if d.Content.Type == ContentTypeImage {
		doc.Content = append(doc.Content, ai.NewMediaPart(d.Content.MIMEType, d.Content.Image))
	} else {
		return nil, fmt.Errorf("unknown content type: %s", d.Content.Type)
	}

	return doc, nil
}
