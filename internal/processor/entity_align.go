package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type EntityAlign struct{}

func (p *EntityAlign) Name() string { return "entity_align" }

func (p *EntityAlign) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	existing, _ := pctx.Extras["existing_entities"]
	if existing == nil && len(pctx.Entities) == 0 {
		return nil
	}

	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("entity_align: prompt template not configured")
	}

	newJSON, _ := json.Marshal(pctx.Entities)
	existingJSON, _ := json.Marshal(existing)
	content := fmt.Sprintf("新提取的实体：\n%s\n\n已有图谱实体：\n%s", string(newJSON), string(existingJSON))

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "entity_align", resp)

	var result struct {
		Entities []pipeline.EntityData `json:"entities"`
	}
	if err := json.Unmarshal([]byte(extractJSON(resp.Content)), &result); err != nil {
		return fmt.Errorf("parse align output: %w", err)
	}

	pctx.Entities = result.Entities
	return nil
}
