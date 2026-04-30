package pipeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/httputil"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var ErrPartial = errors.New("pipeline completed with skipped steps")

type PromptLoader func(id uint) (*model.PromptTemplate, error)

type Engine struct {
	LoadPrompt PromptLoader
}

func NewEngine(opts ...EngineOption) *Engine {
	e := &Engine{}
	for _, o := range opts {
		o(e)
	}
	return e
}

type EngineOption func(*Engine)

func WithPromptLoader(fn PromptLoader) EngineOption {
	return func(e *Engine) { e.LoadPrompt = fn }
}

// Run executes all steps of a pipeline in order against the given context.
func (e *Engine) Run(ctx context.Context, p *model.Pipeline, pctx *ProcessorContext, opts ...RunOption) error {
	o := &RunOptions{}
	for _, fn := range opts {
		fn(o)
	}

	if o.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.Timeout)
		defer cancel()
	}

	if pctx.Document.Type == "image" && pctx.ImageBase64 == "" {
		imageURL := pctx.Document.SourceURL
		if imageURL == "" {
			imageURL = pctx.Document.Content
		}
		if strings.HasPrefix(imageURL, "http") {
			if b64, err := downloadImageBase64(ctx, imageURL); err != nil {
				logger.Warn("failed to pre-download image, will use URL fallback", zap.Error(err))
			} else {
				pctx.ImageBase64 = b64
				pctx.SourceImageURL = imageURL
			}
		}
	}

	steps := e.filterSteps(p.Steps, o.StartFrom, pctx)
	groups := e.groupSteps(steps)

	hasSkipped := false
	for _, grp := range groups {
		if len(grp) == 1 || grp[0].ParallelGroup == 0 {
			for _, step := range grp {
				skipped, err := e.runStep(ctx, step, pctx, o)
				if err != nil {
					return err
				}
				if skipped {
					hasSkipped = true
				}
			}
			continue
		}

		var mu sync.Mutex
		g, gctx := errgroup.WithContext(ctx)
		var groupSkipped bool
		var groupErr error

		for _, step := range grp {
			step := step
			g.Go(func() error {
				proc, err := Get(step.ProcessorType)
				if err != nil {
					return fmt.Errorf("step %d: %w", step.SortOrder, err)
				}

				mu.Lock()
				cfg := e.buildStepConfig(step, pctx)
				mu.Unlock()

				if o.OnStepStart != nil {
					o.OnStepStart(step.SortOrder, step.ProcessorType)
				}

				err = e.executeWithRetry(gctx, proc, pctx, cfg, step)

				log := StepLog{Processor: step.ProcessorType}
				if err != nil {
					log.Error = err.Error()
					mu.Lock()
					pctx.StepLogs = append(pctx.StepLogs, log)
					mu.Unlock()

					if o.OnStepEnd != nil {
						o.OnStepEnd(step.SortOrder, step.ProcessorType, err, log)
					}

					onError := strings.TrimSpace(step.OnError)
					if onError == "" {
						onError = "stop"
					}
					switch onError {
					case "skip":
						logger.Warn("pipeline step skipped due to error",
							zap.String("processor", step.ProcessorType),
							zap.Int("order", step.SortOrder),
							zap.Error(err))
						mu.Lock()
						groupSkipped = true
						mu.Unlock()
						return nil
					default:
						logger.Error("pipeline step failed",
							zap.String("processor", step.ProcessorType),
							zap.Int("order", step.SortOrder),
							zap.Error(err))
						return fmt.Errorf("step %d (%s): %w", step.SortOrder, step.ProcessorType, err)
					}
				}

				mu.Lock()
				pctx.StepLogs = append(pctx.StepLogs, log)
				mu.Unlock()

				logger.Info("pipeline step completed",
					zap.String("processor", step.ProcessorType))

				if o.OnStepEnd != nil {
					o.OnStepEnd(step.SortOrder, step.ProcessorType, nil, log)
				}
				if o.OnProgress != nil {
					o.OnProgress(step.SortOrder)
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			groupErr = err
		}
		if groupSkipped {
			hasSkipped = true
		}
		if groupErr != nil {
			return groupErr
		}
	}

	if hasSkipped {
		return ErrPartial
	}
	return nil
}

// RunSingle executes a single step config (used for sync parse without a stored pipeline).
func (e *Engine) RunSingle(ctx context.Context, pctx *ProcessorContext, cfg StepConfig) error {
	proc, err := Get(cfg.ProcessorType)
	if err != nil {
		return err
	}
	return proc.Process(ctx, pctx, cfg)
}

func (e *Engine) filterSteps(steps []model.PipelineStep, startFrom int, pctx *ProcessorContext) []model.PipelineStep {
	var out []model.PipelineStep
	for _, step := range steps {
		if step.SortOrder < startFrom {
			continue
		}
		if !e.shouldRun(step, pctx) {
			continue
		}
		out = append(out, step)
	}
	return out
}

func (e *Engine) groupSteps(steps []model.PipelineStep) [][]model.PipelineStep {
	var groups [][]model.PipelineStep
	for _, step := range steps {
		pg := step.ParallelGroup
		if pg > 0 && len(groups) > 0 && groups[len(groups)-1][0].ParallelGroup == pg {
			groups[len(groups)-1] = append(groups[len(groups)-1], step)
		} else {
			groups = append(groups, []model.PipelineStep{step})
		}
	}
	return groups
}

func (e *Engine) runStep(ctx context.Context, step model.PipelineStep, pctx *ProcessorContext, o *RunOptions) (skipped bool, err error) {
	proc, procErr := Get(step.ProcessorType)
	if procErr != nil {
		return false, fmt.Errorf("step %d: %w", step.SortOrder, procErr)
	}

	cfg := e.buildStepConfig(step, pctx)

	if o.OnStepStart != nil {
		o.OnStepStart(step.SortOrder, step.ProcessorType)
	}

	procErr = e.executeWithRetry(ctx, proc, pctx, cfg, step)

	log := StepLog{Processor: step.ProcessorType}
	if procErr != nil {
		log.Error = procErr.Error()
		pctx.StepLogs = append(pctx.StepLogs, log)

		if o.OnStepEnd != nil {
			o.OnStepEnd(step.SortOrder, step.ProcessorType, procErr, log)
		}

		onError := strings.TrimSpace(step.OnError)
		if onError == "" {
			onError = "stop"
		}
		switch onError {
		case "skip":
			logger.Warn("pipeline step skipped due to error",
				zap.String("processor", step.ProcessorType),
				zap.Int("order", step.SortOrder),
				zap.Error(procErr))
			return true, nil
		default:
			logger.Error("pipeline step failed",
				zap.String("processor", step.ProcessorType),
				zap.Int("order", step.SortOrder),
				zap.Error(procErr))
			return false, fmt.Errorf("step %d (%s): %w", step.SortOrder, step.ProcessorType, procErr)
		}
	}

	pctx.StepLogs = append(pctx.StepLogs, log)
	logger.Info("pipeline step completed",
		zap.String("processor", step.ProcessorType))

	if o.OnStepEnd != nil {
		o.OnStepEnd(step.SortOrder, step.ProcessorType, nil, log)
	}
	if o.OnProgress != nil {
		o.OnProgress(step.SortOrder)
	}
	return false, nil
}

func (e *Engine) executeWithRetry(ctx context.Context, proc Processor, pctx *ProcessorContext, cfg StepConfig, step model.PipelineStep) error {
	maxAttempts := 1
	if strings.TrimSpace(step.OnError) == "retry" && step.MaxRetry > 0 {
		maxAttempts = step.MaxRetry + 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			logger.Info("retrying pipeline step",
				zap.String("processor", step.ProcessorType),
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
		lastErr = proc.Process(ctx, pctx, cfg)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (e *Engine) shouldRun(step model.PipelineStep, pctx *ProcessorContext) bool {
	cond := strings.TrimSpace(step.Condition)
	if cond == "" {
		return true
	}
	return EvalCondition(cond, pctx)
}

// EvalCondition evaluates a condition string against the processor context.
// Supported forms:
//
//	type=image
//	classification.category=人物
//	classification.tags contains 照片
//	has:entities
func EvalCondition(cond string, pctx *ProcessorContext) bool {
	cond = strings.TrimSpace(cond)
	if cond == "" {
		return true
	}

	// has:field
	if strings.HasPrefix(cond, "has:") {
		field := strings.TrimPrefix(cond, "has:")
		switch field {
		case "entities":
			return len(pctx.Entities) > 0
		case "relations":
			return len(pctx.Relations) > 0
		case "summary":
			return pctx.Summary != ""
		case "raw_text":
			return pctx.RawText != ""
		}
		_, ok := pctx.Extras[field]
		return ok
	}

	// key contains value
	if parts := strings.SplitN(cond, " contains ", 2); len(parts) == 2 {
		actual := resolveValue(strings.TrimSpace(parts[0]), pctx)
		return strings.Contains(fmt.Sprintf("%v", actual), strings.TrimSpace(parts[1]))
	}

	// key=value
	if parts := strings.SplitN(cond, "=", 2); len(parts) == 2 {
		actual := resolveValue(strings.TrimSpace(parts[0]), pctx)
		return fmt.Sprintf("%v", actual) == strings.TrimSpace(parts[1])
	}

	return true
}

func resolveValue(key string, pctx *ProcessorContext) any {
	if key == "type" {
		return pctx.Document.Type
	}
	// Dot-notation lookup in Extras: "classification.category" → Extras["classification"]["category"]
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		if pctx.Extras != nil {
			return pctx.Extras[key]
		}
		return nil
	}
	root, rest := parts[0], parts[1]
	if pctx.Extras == nil {
		return nil
	}
	val, ok := pctx.Extras[root]
	if !ok {
		return nil
	}
	m, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	subParts := strings.SplitN(rest, ".", 2)
	v := m[subParts[0]]
	if len(subParts) == 1 {
		return v
	}
	if sub, ok := v.(map[string]any); ok {
		return sub[subParts[1]]
	}
	return nil
}

func (e *Engine) buildStepConfig(step model.PipelineStep, pctx *ProcessorContext) StepConfig {
	cfg := StepConfig{
		ProcessorType: step.ProcessorType,
		Condition:     step.Condition,
	}

	if step.Config != nil {
		if err := json.Unmarshal([]byte(step.Config), &cfg.Config); err != nil {
			logger.Warn("failed to unmarshal step config", zap.Int("step", step.SortOrder), zap.Error(err))
		}
	}

	// Check prompt_overrides in config for conditional prompt selection
	tmpl := step.PromptTemplate
	if overrides, ok := cfg.Config["prompt_overrides"]; ok {
		if list, ok := overrides.([]any); ok {
			for _, item := range list {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				cond, _ := m["condition"].(string)
				if EvalCondition(cond, pctx) {
					if tid, ok := m["prompt_template_id"].(float64); ok {
						if e.LoadPrompt != nil {
							if pt, err := e.LoadPrompt(uint(tid)); err == nil {
								tmpl = pt
							}
						}
					}
					break
				}
			}
		}
	}

	if tmpl != nil {
		cfg.PromptContent = tmpl.Content
		if tmpl.Variables != nil {
			json.Unmarshal([]byte(tmpl.Variables), &cfg.PromptVariables)
		}
	}

	return cfg
}

// BuildResult converts ProcessorContext to the standard output format.
func BuildResult(pctx *ProcessorContext) map[string]any {
	var processorsUsed []string
	var totalTokens int
	var totalCost float64
	for _, l := range pctx.StepLogs {
		processorsUsed = append(processorsUsed, l.Processor)
		totalTokens += l.Tokens
		totalCost += l.Cost
	}

	return map[string]any{
		"doc_id":      pctx.Document.ID,
		"source_type": pctx.Document.Type,
		"entities":    pctx.Entities,
		"relations":   pctx.Relations,
		"content": map[string]any{
			"raw_text": pctx.RawText,
			"summary":  pctx.Summary,
		},
		"extras": pctx.Extras,
		"metadata": map[string]any{
			"processors_used": processorsUsed,
			"cost":            map[string]any{"tokens": totalTokens, "usd": totalCost},
		},
	}
}

func downloadImageBase64(ctx context.Context, imageURL string) (string, error) {
	data, ct, err := httputil.SafeGetBody(ctx, imageURL, 20*1024*1024)
	if err != nil {
		return "", err
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
