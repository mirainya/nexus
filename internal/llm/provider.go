package llm

import "context"

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type Request struct {
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Response struct {
	Content  string `json:"content"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Usage    Usage  `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, req Request) (*Response, error)
}

type EmbeddingRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Input    string `json:"input"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	Usage     Usage     `json:"usage"`
}

type EmbeddingProvider interface {
	Embedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
}

type ModelInfo struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by,omitempty"`
}

type ModelLister interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
