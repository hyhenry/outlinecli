package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIKey  string
	BaseURL string
	http    *http.Client
}

func New(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) post(endpoint string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/"+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to extract error message from Outline API response
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(raw, &errResp); jsonErr == nil && errResp.Message != "" {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(raw))
	}

	return json.RawMessage(raw), nil
}

// ─── Document types ───────────────────────────────────────────────────────────

type Document struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Text       string `json:"text"`
	URLId      string `json:"urlId"`
	CollectionID string `json:"collectionId"`
	ParentDocumentID string `json:"parentDocumentId,omitempty"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	PublishedAt string `json:"publishedAt,omitempty"`
	ArchivedAt string `json:"archivedAt,omitempty"`
	DeletedAt  string `json:"deletedAt,omitempty"`
}

type Pagination struct {
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	NextPath string `json:"nextPath,omitempty"`
}

// ─── Document methods ─────────────────────────────────────────────────────────

type CreateDocumentParams struct {
	Title            string `json:"title"`
	Text             string `json:"text,omitempty"`
	CollectionID     string `json:"collectionId,omitempty"`
	ParentDocumentID string `json:"parentDocumentId,omitempty"`
	Publish          bool   `json:"publish"`
}

func (c *Client) CreateDocument(params CreateDocumentParams) (*Document, error) {
	raw, err := c.post("documents.create", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) GetDocument(id string) (*Document, error) {
	raw, err := c.post("documents.info", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}

type UpdateDocumentParams struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
	Publish *bool `json:"publish,omitempty"`
}

func (c *Client) UpdateDocument(params UpdateDocumentParams) (*Document, error) {
	raw, err := c.post("documents.update", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) DeleteDocument(id string, permanent bool) error {
	_, err := c.post("documents.delete", map[string]any{
		"id":        id,
		"permanent": permanent,
	})
	return err
}

type ListDocumentsParams struct {
	CollectionID string `json:"collectionId,omitempty"`
	Status       string `json:"statusFilter,omitempty"` // draft, archived, published
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

func (c *Client) ListDocuments(params ListDocumentsParams) ([]Document, *Pagination, error) {
	if params.Limit == 0 {
		params.Limit = 25
	}
	raw, err := c.post("documents.list", params)
	if err != nil {
		return nil, nil, err
	}
	var resp struct {
		Data       []Document `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil, fmt.Errorf("parse response: %w", err)
	}
	return resp.Data, &resp.Pagination, nil
}

type SearchResult struct {
	Ranking  float64  `json:"ranking"`
	Context  string   `json:"context"`
	Document Document `json:"document"`
}

func (c *Client) SearchDocuments(query, collectionID string) ([]SearchResult, error) {
	payload := map[string]any{"query": query}
	if collectionID != "" {
		payload["collectionId"] = collectionID
	}
	raw, err := c.post("documents.search", payload)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []SearchResult `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return resp.Data, nil
}

func (c *Client) ArchiveDocument(id string) (*Document, error) {
	raw, err := c.post("documents.archive", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}

func (c *Client) RestoreDocument(id string) (*Document, error) {
	raw, err := c.post("documents.restore", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Document `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}

// ─── Collection types & methods ───────────────────────────────────────────────

type Collection struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	URLId       string `json:"urlId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func (c *Client) ListCollections() ([]Collection, error) {
	raw, err := c.post("collections.list", map[string]int{"limit": 100})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []Collection `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return resp.Data, nil
}

func (c *Client) GetCollection(id string) (*Collection, error) {
	raw, err := c.post("collections.info", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data Collection `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &resp.Data, nil
}
