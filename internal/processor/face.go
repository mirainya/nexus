package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type Face struct{}

func (p *Face) Name() string { return "face" }

func (p *Face) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("face: prompt template not configured")
	}

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			buildUserMessage(pctx, ""),
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "face", resp)

	var result struct {
		Entities []pipeline.EntityData `json:"entities"`
		Scene    map[string]any        `json:"scene"`
	}
	if err := parseJSON(extractJSON(resp.Content), &result); err != nil {
		return err
	}
	pctx.Entities = append(pctx.Entities, result.Entities...)
	if result.Scene != nil {
		if pctx.Extras == nil {
			pctx.Extras = make(map[string]any)
		}
		pctx.Extras["scene"] = result.Scene
	}
	return nil
}
