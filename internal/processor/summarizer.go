package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type Summarizer struct{}

func (p *Summarizer) Name() string { return "summarizer" }

func (p *Summarizer) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	text := pctx.RawText
	if text == "" {
		text = pctx.Document.Content
	}
	if text == "" {
		return fmt.Errorf("summarizer: no text to summarize")
	}

	prompt := cfg.PromptContent
	if prompt == "" {
		prompt = "请对以下文本生成简洁的摘要，保留关键信息：\n\n{{text}}"
	}
	prompt = renderPrompt(prompt, cfg.PromptVariables)

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)
	maxTokens := 1024
	if v, ok := cfg.Config["max_tokens"].(float64); ok && v > 0 {
		maxTokens = int(v)
	}

	req := llm.Request{
		Provider:  provider,
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			buildUserMessage(pctx, text),
		},
	}

	resp, err := doChat(ctx, pctx, req)
	if err != nil {
		return fmt.Errorf("summarizer: %w", err)
	}

	pctx.Summary = resp.Content
	return nil
}
