package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type Classifier struct{}

func (p *Classifier) Name() string { return "classifier" }

func (p *Classifier) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("classifier: prompt template not configured")
	}

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			buildUserMessage(pctx, pctx.Document.Content),
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "classifier", resp)

	var result map[string]any
	if err := parseJSON(extractJSON(resp.Content), &result); err != nil {
		return err
	}
	if pctx.Extras == nil {
		pctx.Extras = make(map[string]any)
	}
	pctx.Extras["classification"] = result
	return nil
}
