package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type Embedding struct{}

func (p *Embedding) Name() string { return "embedding" }

func (p *Embedding) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)
	if model == "" {
		model = "text-embedding-3-small"
	}

	text := pctx.Document.Content
	if pctx.Summary != "" {
		text = pctx.Summary
	}

	resp, err := llm.G.Embedding(ctx, llm.EmbeddingRequest{
		Provider: provider,
		Model:    model,
		Input:    text,
	})
	if err != nil {
		return fmt.Errorf("embedding: %w", err)
	}

	resolvedProvider := provider
	if resolvedProvider == "" {
		resolvedProvider = resp.Provider
	}
	for i := len(pctx.StepLogs) - 1; i >= 0; i-- {
		if pctx.StepLogs[i].Processor == "embedding" {
			pctx.StepLogs[i].Tokens = resp.Usage.TotalTokens
			pctx.StepLogs[i].Cost = llm.G.CalcCost(resolvedProvider, resp.Usage)
			break
		}
	}

	if pctx.Extras == nil {
		pctx.Extras = make(map[string]any)
	}
	pctx.Extras["embedding"] = resp.Embedding
	return nil
}
