package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/internal/sse"
	"github.com/mirainya/nexus/pkg/cache"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JobService struct {
	engine      *pipeline.Engine
	asynqClient *asynq.Client
}

func NewJobService(client *asynq.Client) *JobService {
	return &JobService{engine: pipeline.NewEngine(), asynqClient: client}
}

type JobSubmitRequest struct {
	Content     string         `json:"content"`
	Type        string         `json:"type" binding:"required"`
	SourceURL   string         `json:"source_url"`
	PipelineID  uint           `json:"pipeline_id" binding:"required"`
	CallbackURL string         `json:"callback_url"`
	SkipCache   bool           `json:"skip_cache"`
	Metadata    map[string]any `json:"metadata"`
}

func (s *JobService) computeContentHash(req JobSubmitRequest, pipelineUpdatedAt time.Time) string {
	key := req.Type + "|" + req.SourceURL + "|" + req.Content + "|" +
		fmt.Sprintf("%d", req.PipelineID) + "|" + pipelineUpdatedAt.UTC().Format(time.RFC3339)
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)
}

func (s *JobService) Submit(req JobSubmitRequest) (*model.Job, error) {
	var p model.Pipeline
	if err := model.DB().Select("id, updated_at").First(&p, req.PipelineID).Error; err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	contentHash := s.computeContentHash(req, p.UpdatedAt)

	if !req.SkipCache {
		var cached model.Job
		err := model.DB().Where("content_hash = ? AND status = ?", contentHash, "completed").
			Order("created_at DESC").First(&cached).Error
		if err == nil {
			return &cached, nil
		}
	}

	metaJSON, _ := json.Marshal(req.Metadata)
	doc := model.Document{
		UUID:      uuid.New().String(),
		Type:      req.Type,
		Content:   req.Content,
		SourceURL: req.SourceURL,
		Metadata:  metaJSON,
		Status:    "pending",
	}
	if err := model.DB().Create(&doc).Error; err != nil {
		return nil, err
	}

	job := model.Job{
		UUID:        uuid.New().String(),
		DocumentID:  doc.ID,
		PipelineID:  req.PipelineID,
		Status:      "pending",
		ContentHash: contentHash,
		CallbackURL: req.CallbackURL,
	}
	if err := model.DB().Create(&job).Error; err != nil {
		return nil, err
	}

	if err := s.enqueue(job.ID); err != nil {
		return &job, err
	}

	return &job, nil
}

func (s *JobService) GetByUUID(uuid string) (*model.Job, error) {
	var job model.Job
	if err := model.DB().Preload("StepLogs", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("uuid = ?", uuid).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *JobService) Execute(ctx context.Context, jobID uint) error {
	var job model.Job
	if err := model.DB().First(&job, jobID).Error; err != nil {
		return err
	}
	if job.Status != "pending" {
		logger.Info("skip job with non-pending status", zap.Uint("job_id", job.ID), zap.String("status", job.Status))
		return nil
	}

	claimed := model.DB().Model(&model.Job{}).
		Where("id = ? AND status = ?", jobID, "pending").
		Updates(map[string]any{"status": "running", "error": ""})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected == 0 {
		logger.Info("job already claimed or finished", zap.Uint("job_id", jobID))
		return nil
	}

	var doc model.Document
	if err := model.DB().First(&doc, job.DocumentID).Error; err != nil {
		return err
	}

	var p model.Pipeline
	if err := model.DB().Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Preload("Steps.PromptTemplate").First(&p, job.PipelineID).Error; err != nil {
		return err
	}

	var meta map[string]any
	if doc.Metadata != nil {
		if err := json.Unmarshal([]byte(doc.Metadata), &meta); err != nil {
			logger.Warn("failed to unmarshal doc metadata", zap.Uint("doc_id", doc.ID), zap.Error(err))
		}
	}

	pctx := &pipeline.ProcessorContext{
		Document: pipeline.DocumentData{
			ID:        doc.UUID,
			Type:      doc.Type,
			Content:   doc.Content,
			SourceURL: doc.SourceURL,
			Metadata:  meta,
		},
	}

	totalSteps := len(p.Steps)
	model.DB().Model(&job).Update("total_steps", totalSteps)

	if err := s.engine.Run(ctx, &p, pctx,
		pipeline.WithStartFrom(job.CurrentStep),
		pipeline.WithOnStepStart(func(stepOrder int, processorType string) {
			now := time.Now()
			model.DB().Create(&model.JobStepLog{
				JobID:         job.ID,
				StepOrder:     stepOrder,
				ProcessorType: processorType,
				Status:        "running",
				StartedAt:     &now,
			})
			model.DB().Model(&job).Update("current_step", stepOrder)
			sse.Default().Publish(job.UUID, sse.Event{
				Type:      "step_start",
				Step:      stepOrder,
				Processor: processorType,
			})
		}),
		pipeline.WithOnStepEnd(func(stepOrder int, processorType string, stepErr error, log pipeline.StepLog) {
			now := time.Now()
			updates := map[string]any{
				"finished_at": &now,
				"tokens":      log.Tokens,
				"cost":        log.Cost,
			}
			if stepErr != nil {
				updates["status"] = "failed"
				updates["error"] = stepErr.Error()
			} else {
				updates["status"] = "completed"
			}
			model.DB().Model(&model.JobStepLog{}).
				Where("job_id = ? AND step_order = ?", job.ID, stepOrder).
				Updates(updates)
			model.DB().Model(&job).Update("current_step", stepOrder+1)

			evt := sse.Event{
				Type:      "step_end",
				Step:      stepOrder,
				Processor: processorType,
			}
			if stepErr != nil {
				evt.Error = stepErr.Error()
			}
			sse.Default().Publish(job.UUID, evt)
		}),
	); err != nil {
		if errors.Is(err, pipeline.ErrPartial) {
			result, _ := json.Marshal(pipeline.BuildResult(pctx))
			model.DB().Model(&job).Updates(map[string]any{"status": "partial", "result": result})
			if persistErr := s.persistResults(pctx, doc.ID); persistErr != nil {
				model.DB().Model(&job).Updates(map[string]any{"error": "persist results: " + persistErr.Error()})
			}
			sse.Default().Publish(job.UUID, sse.Event{Type: "completed", Data: "partial"})
			return nil
		}
		model.DB().Model(&job).Updates(map[string]any{"status": "failed", "error": err.Error()})
		sse.Default().Publish(job.UUID, sse.Event{Type: "failed", Error: err.Error()})
		return err
	}

	if err := s.persistResults(pctx, doc.ID); err != nil {
		model.DB().Model(&job).Updates(map[string]any{"status": "failed", "error": "persist results: " + err.Error()})
		sse.Default().Publish(job.UUID, sse.Event{Type: "failed", Error: "persist results: " + err.Error()})
		return err
	}

	result, _ := json.Marshal(pipeline.BuildResult(pctx))
	model.DB().Model(&job).Updates(map[string]any{"status": "completed", "result": result})
	model.DB().Model(&doc).Update("status", "completed")
	sse.Default().Publish(job.UUID, sse.Event{Type: "completed", Data: pipeline.BuildResult(pctx)})

	if cache.Available() && job.ContentHash != "" {
		cacheKey := "nexus:parse:" + job.ContentHash
		cache.Set(ctx, cacheKey, string(result), 24*time.Hour)
	}

	return nil
}

func (s *JobService) Retry(jobID uint) (*model.Job, error) {
	var job model.Job
	if err := model.DB().First(&job, jobID).Error; err != nil {
		return nil, err
	}
	if job.Status != "failed" {
		return nil, errors.New("only failed jobs can be retried")
	}

	model.DB().Model(&job).Updates(map[string]any{"status": "pending", "error": ""})

	if err := s.enqueue(job.ID); err != nil {
		return &job, err
	}
	return &job, nil
}

func (s *JobService) RecoverStalled() error {
	if s.asynqClient == nil {
		return nil
	}

	var jobs []model.Job
	if err := model.DB().Where("status IN ?", []string{"pending", "running"}).Find(&jobs).Error; err != nil {
		return err
	}
	if err := model.DB().Model(&model.Job{}).
		Where("status = ?", "running").
		Updates(map[string]any{"status": "pending"}).Error; err != nil {
		return err
	}
	for _, job := range jobs {
		if err := s.enqueueRecovered(job.ID); err != nil {
			return err
		}
	}
	if len(jobs) > 0 {
		logger.Info("recovered stalled jobs", zap.Int("count", len(jobs)))
	}
	return nil
}

func (s *JobService) enqueue(jobID uint) error {
	if s.asynqClient == nil {
		return nil
	}
	payload, _ := json.Marshal(map[string]uint{"job_id": jobID})
	task := asynq.NewTask("pipeline:execute", payload)
	_, err := s.asynqClient.Enqueue(task, asynq.TaskID(fmt.Sprintf("pipeline:execute:%d", jobID)))
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}

func (s *JobService) enqueueRecovered(jobID uint) error {
	if s.asynqClient == nil {
		return nil
	}
	payload, _ := json.Marshal(map[string]uint{"job_id": jobID})
	task := asynq.NewTask("pipeline:execute", payload)
	_, err := s.asynqClient.Enqueue(task)
	return err
}

func (s *JobService) List(page, pageSize int, status string) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64
	q := model.DB().Model(&model.Job{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).
		Preload("StepLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).Find(&jobs).Error
	return jobs, total, err
}

type RecommendItem struct {
	DocumentID uint     `json:"document_id"`
	SourceURL  string   `json:"source_url"`
	Content    string   `json:"content"`
	Scene      string   `json:"scene"`
	Score      float64  `json:"score"`
	Reason     string   `json:"reason"`
	Tags       []string `json:"tags,omitempty"`
}

func (s *JobService) RecommendByScene(scene string, limit int) ([]RecommendItem, error) {
	if limit <= 0 {
		limit = 20
	}

	var jobs []model.Job
	err := model.DB().
		Joins("JOIN documents ON documents.id = jobs.document_id").
		Where("documents.type = ? AND jobs.status = ?", "image", "completed").
		Where("jobs.result IS NOT NULL").
		Where("jobs.result::jsonb -> 'extras' -> 'image_assessment' -> 'use_cases' IS NOT NULL").
		Preload("Document").
		Limit(limit * 5).
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}

	var items []RecommendItem
	for _, job := range jobs {
		if job.Result == nil {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal(job.Result, &result); err != nil {
			continue
		}
		extras, _ := result["extras"].(map[string]any)
		if extras == nil {
			continue
		}
		assessment, _ := extras["image_assessment"].(map[string]any)
		if assessment == nil {
			continue
		}
		useCases, _ := assessment["use_cases"].([]any)
		for _, uc := range useCases {
			ucMap, _ := uc.(map[string]any)
			if ucMap == nil {
				continue
			}
			ucScene, _ := ucMap["scene"].(string)
			suitable, _ := ucMap["suitable"].(bool)
			if ucScene == scene && suitable {
				score, _ := ucMap["score"].(float64)
				reason, _ := ucMap["reason"].(string)
				var tags []string
				if rawTags, ok := ucMap["tags"].([]any); ok {
					for _, t := range rawTags {
						if s, ok := t.(string); ok {
							tags = append(tags, s)
						}
					}
				}
				items = append(items, RecommendItem{
					DocumentID: job.DocumentID,
					SourceURL:  job.Document.SourceURL,
					Content:    job.Document.Content,
					Scene:      scene,
					Score:      score,
					Reason:     reason,
					Tags:       tags,
				})
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *JobService) persistResults(pctx *pipeline.ProcessorContext, sourceID uint) error {
	return model.DB().Transaction(func(tx *gorm.DB) error {
		entityNameToID := make(map[string]uint)

		for _, e := range pctx.Entities {
			aliasesJSON, _ := json.Marshal(e.Aliases)
			attrsJSON, _ := json.Marshal(e.Attributes)
			evidenceJSON, _ := json.Marshal(e.Evidence)

			// Check if entity_align marked this as an existing entity
			var existingID uint
			if e.Attributes != nil {
				if eid, ok := e.Attributes["existing_id"]; ok {
					switch v := eid.(type) {
					case float64:
						existingID = uint(v)
					case json.Number:
						n, _ := v.Int64()
						existingID = uint(n)
					}
				}
			}

			if existingID > 0 {
				// Update existing entity
				tx.Model(&model.Entity{}).Where("id = ?", existingID).Updates(map[string]any{
					"attributes": attrsJSON,
					"confidence": e.Confidence,
					"evidence":   evidenceJSON,
				})
				// Merge aliases
				var existing model.Entity
				if tx.First(&existing, existingID).Error == nil {
					var oldAliases, newAliases []string
					json.Unmarshal(existing.Aliases, &oldAliases)
					seen := make(map[string]bool)
					for _, a := range oldAliases {
						seen[a] = true
						newAliases = append(newAliases, a)
					}
					for _, a := range e.Aliases {
						if !seen[a] {
							newAliases = append(newAliases, a)
						}
					}
					merged, _ := json.Marshal(newAliases)
					tx.Model(&existing).Update("aliases", merged)
				}
				entityNameToID[e.Name] = existingID
			} else {
				entity := model.Entity{
					UUID:       uuid.New().String(),
					Type:       e.Type,
					Name:       e.Name,
					Aliases:    aliasesJSON,
					Attributes: attrsJSON,
					Confidence: e.Confidence,
					SourceID:   sourceID,
					Evidence:   evidenceJSON,
				}
				if err := tx.Create(&entity).Error; err != nil {
					return err
				}
				entityNameToID[e.Name] = entity.ID

				originalJSON, _ := json.Marshal(e)
				review := model.Review{
					EntityID:     &entity.ID,
					Status:       "pending",
					OriginalData: originalJSON,
				}
				if err := tx.Create(&review).Error; err != nil {
					return err
				}
			}
		}

		for _, r := range pctx.Relations {
			fromID, fromOK := entityNameToID[r.From]
			toID, toOK := entityNameToID[r.To]
			if !fromOK || !toOK {
				continue
			}
			metaJSON, _ := json.Marshal(r.Metadata)
			rel := model.Relation{
				UUID:         uuid.New().String(),
				FromEntityID: fromID,
				ToEntityID:   toID,
				Type:         r.Type,
				Metadata:     metaJSON,
				Confidence:   r.Confidence,
				SourceID:     sourceID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
