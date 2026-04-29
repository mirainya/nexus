package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/pipeline"
)

type LLMExtract struct{}

func (p *LLMExtract) Name() string { return "llm_extract" }

func (p *LLMExtract) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	prompt := renderPrompt(cfg.PromptContent, cfg.PromptVariables)
	if prompt == "" {
		return fmt.Errorf("llm_extract: prompt template not configured")
	}

	provider, _ := cfg.Config["provider"].(string)
	model, _ := cfg.Config["model"].(string)

	userContent := pctx.Document.Content
	var extras []string
	var visuals []pipeline.EntityData
	for _, e := range pctx.Entities {
		if e.Type == "person_visual" {
			visuals = append(visuals, e)
		}
	}
	if len(visuals) > 0 {
		if b, err := json.Marshal(visuals); err == nil {
			extras = append(extras, "视觉观察数据（必须合并到对应 person 实体）：\n"+string(b))
		}
	}
	if pctx.SourceImageURL != "" {
		extras = append(extras, "图片来源："+pctx.SourceImageURL)
	}
	if pctx.Extras != nil {
		if scene, ok := pctx.Extras["scene"]; ok {
			if b, err := json.Marshal(scene); err == nil {
				extras = append(extras, "场景信息："+string(b))
			}
		}
		if assess, ok := pctx.Extras["image_assessment"]; ok {
			if b, err := json.Marshal(assess); err == nil {
				extras = append(extras, "图片评估："+string(b))
			}
		}
	}
	if len(extras) > 0 {
		userContent = userContent + "\n\n参考信息：\n" + strings.Join(extras, "\n")
	}

	resp, err := llm.G.Chat(ctx, llm.Request{
		Provider: provider,
		Model:    model,
		Messages: []llm.Message{
			{Role: "system", Content: prompt},
			buildUserMessage(pctx, userContent),
		},
	})
	if err != nil {
		return err
	}

	updateStepLog(pctx, "llm_extract", resp)

	var result struct {
		Entities  []pipeline.EntityData   `json:"entities"`
		Relations []pipeline.RelationData `json:"relations"`
		Summary   string                  `json:"summary"`
	}
	if err := json.Unmarshal([]byte(extractJSON(resp.Content)), &result); err != nil {
		return fmt.Errorf("parse llm output: %w", err)
	}

	pctx.Entities = append(pctx.Entities, result.Entities...)
	pctx.Relations = append(pctx.Relations, result.Relations...)
	if result.Summary != "" {
		pctx.Summary = result.Summary
	}
	return nil
}

func renderPrompt(template string, vars map[string]any) string {
	if template == "" {
		return ""
	}
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return result
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func updateStepLog(pctx *pipeline.ProcessorContext, name string, resp *llm.Response) {
	for i := len(pctx.StepLogs) - 1; i >= 0; i-- {
		if pctx.StepLogs[i].Processor == name {
			pctx.StepLogs[i].Tokens = resp.Usage.TotalTokens
			pctx.StepLogs[i].Cost = llm.G.CalcCost(resp.Provider, resp.Usage)
			return
		}
	}
}
