package processor

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
)

type ContextLoader struct{}

func (p *ContextLoader) Name() string { return "context_loader" }

func (p *ContextLoader) Process(_ context.Context, pctx *pipeline.ProcessorContext, _ pipeline.StepConfig) error {
	content := pctx.Document.Content
	if pctx.RawText != "" {
		content = pctx.RawText
	}

	var entities []model.Entity
	q := pctx.DB.Limit(200)
	if pctx.TenantID > 0 {
		q = q.Where("tenant_id = ?", pctx.TenantID)
	}
	q.Find(&entities)

	var matched []map[string]any
	for _, e := range entities {
		if nameInContent(e.Name, content) || aliasInContent(e.Aliases, content) {
			matched = append(matched, map[string]any{
				"id":         e.ID,
				"type":       e.Type,
				"name":       e.Name,
				"aliases":    json.RawMessage(e.Aliases),
				"attributes": json.RawMessage(e.Attributes),
				"confidence": e.Confidence,
				"confirmed":  e.Confirmed,
			})
		}
	}

	if pctx.Extras == nil {
		pctx.Extras = make(map[string]any)
	}
	pctx.Extras["existing_entities"] = matched
	return nil
}

func nameInContent(name, content string) bool {
	return name != "" && strings.Contains(content, name)
}

func aliasInContent(aliasesJSON []byte, content string) bool {
	if len(aliasesJSON) == 0 {
		return false
	}
	var aliases []string
	if json.Unmarshal(aliasesJSON, &aliases) != nil {
		return false
	}
	for _, a := range aliases {
		if a != "" && strings.Contains(content, a) {
			return true
		}
	}
	return false
}
