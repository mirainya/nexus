package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type OCR struct{}

func (p *OCR) Name() string { return "ocr" }

func (p *OCR) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("ocr: prompt template not configured")
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

	updateStepLog(pctx, "ocr", resp)
	pctx.RawText = resp.Content
	pctx.Document.Content = resp.Content
	return nil
}
