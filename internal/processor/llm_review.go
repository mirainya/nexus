package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type LLMReview struct{}

func (p *LLMReview) Name() string { return "llm_review" }

func (p *LLMReview) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("llm_review: prompt template not configured")
	}

	entitiesJSON, _ := json.Marshal(pctx.Entities)
	relationsJSON, _ := json.Marshal(pctx.Relations)
	content := fmt.Sprintf("原始内容：\n%s\n\n已提取实体：\n%s\n\n已提取关系：\n%s",
		pctx.Document.Content, string(entitiesJSON), string(relationsJSON))

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			buildUserMessage(pctx, content),
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "llm_review", resp)

	var result struct {
		Entities  []pipeline.EntityData   `json:"entities"`
		Relations []pipeline.RelationData `json:"relations"`
	}
	if err := json.Unmarshal([]byte(extractJSON(resp.Content)), &result); err != nil {
		return fmt.Errorf("parse review output: %w", err)
	}

	var preserved []pipeline.EntityData
	for _, e := range pctx.Entities {
		if e.Type == "person_visual" {
			preserved = append(preserved, e)
		}
	}
	pctx.Entities = append(preserved, result.Entities...)
	pctx.Relations = result.Relations
	return nil
}
