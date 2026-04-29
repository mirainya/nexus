package llm

import "context"

// DoubaoProvider uses OpenAI-compatible API format.
type DoubaoProvider struct {
	inner *OpenAIProvider
}

func NewDoubao(apiKey, baseURL string) *DoubaoProvider {
	return &DoubaoProvider{inner: NewOpenAI(apiKey, baseURL)}
}

func (p *DoubaoProvider) Name() string { return "doubao" }

func (p *DoubaoProvider) Chat(ctx context.Context, req Request) (*Response, error) {
	return p.inner.Chat(ctx, req)
}

func (p *DoubaoProvider) Embedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	return p.inner.Embedding(ctx, req)
}

func (p *DoubaoProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	return p.inner.ListModels(ctx)
}
