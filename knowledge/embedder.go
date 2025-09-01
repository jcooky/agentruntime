package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type (
	EmbeddingTaskType string

	Embedder struct {
		client *http.Client
		apiKey string
	}
)

const (
	EmbeddingTaskTypeDocument EmbeddingTaskType = "search_document"
	EmbeddingTaskTypeQuery    EmbeddingTaskType = "search_query"

	NomicEmbedderTextEndpoint  = "https://api-atlas.nomic.ai/v1/embedding/text"
	NomicEmbedderImageEndpoint = "https://api-atlas.nomic.ai/v1/embedding/image"

	NomicVisionEmbedderModel = "nomic-embed-vision-v1.5"
	NomicTextEmbedderModel   = "nomic-embed-text-v1.5"
)

func (e *EmbeddingTaskType) String() string {
	return string(*e)
}

func NewEmbedder(apiKey string) Embedder {
	return Embedder{client: http.DefaultClient, apiKey: apiKey}
}

func (e *Embedder) EmbedTexts(ctx context.Context, taskType EmbeddingTaskType, texts ...string) ([][]float32, error) {
	var requestBody bytes.Buffer
	if err := json.NewEncoder(&requestBody).Encode(struct {
		TaskType string   `json:"task_type"`
		Model    string   `json:"model"`
		Texts    []string `json:"texts"`
	}{
		TaskType: taskType.String(),
		Model:    NomicTextEmbedderModel,
		Texts:    texts,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to encode request body")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, NomicEmbedderTextEndpoint, &requestBody)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to embed text")
	}

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response")
	}

	return response.Embeddings, nil
}

func (e *Embedder) EmbedImageUrls(ctx context.Context, imageUrls ...string) ([][]float32, error) {
	// Create form data
	formData := url.Values{}
	formData.Set("model", NomicVisionEmbedderModel)
	for _, imageUrl := range imageUrls {
		formData.Add("urls", imageUrl)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, NomicEmbedderImageEndpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("failed to embed image: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response")
	}

	return response.Embeddings, nil
}

func (e *Embedder) EmbedImageFiles(ctx context.Context, mimeType string, imageFiles ...[]byte) ([][]float32, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add model field
	err := writer.WriteField("model", NomicVisionEmbedderModel)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write model field")
	}

	// Add image files
	for i, imageFile := range imageFiles {
		var filename string
		switch mimeType {
		case "image/jpeg", "image/jpg":
			filename = "image%d.jpg"
		case "image/png":
			filename = "image%d.png"
		case "image/gif":
			filename = "image%d.gif"
		case "image/webp":
			filename = "image%d.webp"
		}
		part, err := writer.CreateFormFile("images", filename)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create form file %d", i)
		}
		_, err = io.Copy(part, bytes.NewReader(imageFile))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to copy image data %d", i)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to close multipart writer")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, NomicEmbedderImageEndpoint, &requestBody)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("failed to embed image files: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrapf(err, "failed to decode response")
	}

	return response.Embeddings, nil
}

func (e *Embedder) GetEmbedSize() int {
	return 768
}
