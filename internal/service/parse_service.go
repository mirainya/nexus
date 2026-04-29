package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"gorm.io/gorm"
)

type ParseService struct {
	engine *pipeline.Engine
}

func NewParseService() *ParseService {
	return &ParseService{engine: pipeline.NewEngine()}
}

type ParseRequest struct {
	Content    string         `json:"content"`
	Type       string         `json:"type" binding:"required"`
	SourceURL  string         `json:"source_url"`
	PipelineID uint           `json:"pipeline_id"`
	Metadata   map[string]any `json:"metadata"`
}

func (s *ParseService) Parse(ctx context.Context, req ParseRequest) (map[string]any, error) {
	p, err := s.loadPipeline(req.PipelineID)
	if err != nil {
		return nil, err
	}

	pctx := &pipeline.ProcessorContext{
		Document: pipeline.DocumentData{
			ID:        uuid.New().String(),
			Type:      req.Type,
			Content:   req.Content,
			SourceURL: req.SourceURL,
			Metadata:  req.Metadata,
		},
	}

	if err := s.engine.Run(ctx, p, pctx); err != nil {
		return nil, err
	}

	return pipeline.BuildResult(pctx), nil
}

func (s *ParseService) loadPipeline(id uint) (*model.Pipeline, error) {
	var p model.Pipeline
	if err := model.DB().Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Preload("Steps.PromptTemplate").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
