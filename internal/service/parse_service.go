package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
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
	SkipCache  bool           `json:"skip_cache"`
	Metadata   map[string]any `json:"metadata"`
}

func (s *ParseService) computeCacheKey(req ParseRequest, pipelineUpdatedAt time.Time) string {
	key := "parse:" + req.Type + "|" + req.SourceURL + "|" + req.Content + "|" +
		fmt.Sprintf("%d", req.PipelineID) + "|" + pipelineUpdatedAt.UTC().Format(time.RFC3339)
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("nexus:parse:%x", h)
}

func (s *ParseService) Parse(ctx context.Context, req ParseRequest) (map[string]any, error) {
	p, err := s.loadPipeline(req.PipelineID)
	if err != nil {
		return nil, err
	}

	if !req.SkipCache && cache.Available() {
		cacheKey := s.computeCacheKey(req, p.UpdatedAt)
		cached, err := cache.Get(ctx, cacheKey)
		if err == nil {
			var result map[string]any
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result, nil
			}
		} else if err != redis.Nil {
			// log but don't fail on cache errors
		}
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

	result := pipeline.BuildResult(pctx)

	if cache.Available() {
		cacheKey := s.computeCacheKey(req, p.UpdatedAt)
		if b, err := json.Marshal(result); err == nil {
			cache.Set(ctx, cacheKey, string(b), 24*time.Hour)
		}
	}

	return result, nil
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
