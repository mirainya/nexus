package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
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

	resp, err := doEmbedding(ctx, pctx, llm.EmbeddingRequest{
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

	if pctx.VectorDB != nil && len(resp.Embedding) > 0 {
		vec32 := make([]float32, len(resp.Embedding))
		for i, v := range resp.Embedding {
			vec32[i] = float32(v)
		}
		meta := map[string]any{
			"doc_id":   pctx.Document.ID,
			"doc_type": pctx.Document.Type,
		}
		collection := config.C.Milvus.Collection
		if collection == "" {
			collection = "nexus_embeddings"
		}
		if err := pctx.VectorDB.Insert(collection, pctx.Document.ID, vec32, meta); err != nil {
			logger.Warn("failed to insert vector", zap.String("doc_id", pctx.Document.ID), zap.Error(err))
		}
	}

	return nil
}
