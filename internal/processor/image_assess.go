package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type ImageAssess struct{}

func (p *ImageAssess) Name() string { return "image_assess" }

func (p *ImageAssess) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("image_assess: prompt template not configured")
	}

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	imageURL := pctx.SourceImageURL
	if imageURL == "" {
		imageURL = pctx.Document.SourceURL
	}
	if imageURL == "" {
		imageURL = urlRe.FindString(pctx.Document.Content)
	}

	var userMsg llm.Message
	if imageURL != "" {
		userMsg = llm.Message{
			Role: "user",
			Content: []llm.ContentPart{
				{Type: "image_url", ImageURL: &llm.ImageURL{URL: imageURL}},
			},
		}
	} else {
		userMsg = llm.Message{Role: "user", Content: "请根据以上要求进行评估。"}
	}

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			userMsg,
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "image_assess", resp)

	var result map[string]any
	if err := parseJSON(extractJSON(resp.Content), &result); err != nil {
		return fmt.Errorf("parse image_assess output: %w", err)
	}

	if pctx.Extras == nil {
		pctx.Extras = make(map[string]any)
	}
	pctx.Extras["image_assessment"] = result
	return nil
}
