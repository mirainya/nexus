package processor

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"gorm.io/gorm"
)

type SubPipeline struct{}

func (p *SubPipeline) Name() string { return "sub_pipeline" }

func (p *SubPipeline) Process(ctx context.Context, pctx *pipeline.ProcessorContext, cfg pipeline.StepConfig) error {
	pidVal, ok := cfg.Config["pipeline_id"]
	if !ok {
		return fmt.Errorf("sub_pipeline: pipeline_id is required in config")
	}
	pid, ok := pidVal.(float64)
	if !ok {
		return fmt.Errorf("sub_pipeline: pipeline_id must be a number")
	}

	var target model.Pipeline
	if err := pctx.DB.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order")
	}).First(&target, uint(pid)).Error; err != nil {
		return fmt.Errorf("sub_pipeline: load pipeline %d: %w", uint(pid), err)
	}

	engine := pipeline.NewEngine(pipeline.WithPromptLoader(func(id uint) (*model.PromptTemplate, error) {
		var pt model.PromptTemplate
		if err := pctx.DB.First(&pt, id).Error; err != nil {
			return nil, err
		}
		return &pt, nil
	}))

	return engine.Run(ctx, &target, pctx)
}
