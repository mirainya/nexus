package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/pipeline"
)

type Router struct{}

func (r *Router) Name() string { return "router" }

func (r *Router) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	routesVal, ok := cfg.Config["routes"]
	if !ok {
		return fmt.Errorf("router: routes is required in config")
	}
	routes, ok := routesVal.([]any)
	if !ok {
		return fmt.Errorf("router: routes must be an array")
	}

	for _, item := range routes {
		route, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cond, _ := route["condition"].(string)
		if !pipeline.EvalCondition(cond, pctx) {
			continue
		}

		steps, ok := route["steps"].([]any)
		if !ok {
			return fmt.Errorf("router: matched route has no steps")
		}

		for i, stepItem := range steps {
			step, ok := stepItem.(map[string]any)
			if !ok {
				continue
			}
			procType, _ := step["processor_type"].(string)
			if procType == "" {
				return fmt.Errorf("router: step %d missing processor_type", i)
			}

			proc, err := pipeline.Get(procType)
			if err != nil {
				return fmt.Errorf("router: step %d: %w", i, err)
			}

			stepCfg := pipeline.StepConfig{
				ProcessorType: procType,
			}
			if configMap, ok := step["config"].(map[string]any); ok {
				stepCfg.Config = configMap
			}
			if prompt, ok := step["prompt_content"].(string); ok {
				stepCfg.PromptContent = prompt
			}

			if err := proc.Process(ctx, pctx, stepCfg); err != nil {
				return fmt.Errorf("router: step %d (%s): %w", i, procType, err)
			}
		}
		return nil
	}

	return nil
}
