package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/internal/sse"
	"github.com/mirainya/nexus/pkg/cache"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"github.com/mirainya/nexus/pkg/vectordb"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JobService struct {
	db          *gorm.DB
	engine      *pipeline.Engine
	asynqClient *asynq.Client
	hub         *sse.Hub
	gw          *llm.Gateway
	persister   *ResultPersister
	tracker     *UsageTracker
	webhook     *WebhookService
}

func NewJobService(db *gorm.DB, client *asynq.Client, hub *sse.Hub, gw *llm.Gateway) *JobService {
	return &JobService{
		db:          db,
		engine:      pipeline.NewEngine(),
		asynqClient: client,
		hub:         hub,
		gw:          gw,
		persister:   NewResultPersister(db),
		tracker:     NewUsageTracker(db),
		webhook:     NewWebhookService(db),
	}
}

type JobSubmitRequest struct {
	Content      string         `json:"content"`
	Type         string         `json:"type" binding:"required"`
	SourceURL    string         `json:"source_url"`
	PipelineID   uint           `json:"pipeline_id" binding:"required"`
	CallbackURL  string         `json:"callback_url"`
	SkipCache    bool           `json:"skip_cache"`
	Metadata     map[string]any `json:"metadata"`
	CredentialID *uint          `json:"credential_id"`
	APIKeyID     *uint          `json:"-"`
}

func (s *JobService) computeContentHash(req JobSubmitRequest, pipelineUpdatedAt time.Time) string {
	key := req.Type + "|" + req.SourceURL + "|" + req.Content + "|" +
		fmt.Sprintf("%d", req.PipelineID) + "|" + pipelineUpdatedAt.UTC().Format(time.RFC3339)
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)
}

func (s *JobService) Submit(req JobSubmitRequest) (*model.Job, error) {
	var p model.Pipeline
	if err := s.db.Select("id, updated_at").First(&p, req.PipelineID).Error; err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	if req.CredentialID != nil {
		var cred model.Credential
		if err := s.db.First(&cred, *req.CredentialID).Error; err != nil {
			return nil, fmt.Errorf("credential not found: %w", err)
		}
		if req.APIKeyID != nil && cred.APIKeyID != *req.APIKeyID {
			return nil, errors.New("credential does not belong to this api key")
		}
		if !cred.Active {
			return nil, errors.New("credential is inactive")
		}
	}

	contentHash := s.computeContentHash(req, p.UpdatedAt)

	if !req.SkipCache {
		var cached model.Job
		err := s.db.Where("content_hash = ? AND status = ?", contentHash, "completed").
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
	if err := s.db.Create(&doc).Error; err != nil {
		return nil, err
	}

	job := model.Job{
		UUID:         uuid.New().String(),
		DocumentID:   doc.ID,
		PipelineID:   req.PipelineID,
		Status:       "pending",
		ContentHash:  contentHash,
		CallbackURL:  req.CallbackURL,
		APIKeyID:     req.APIKeyID,
		CredentialID: req.CredentialID,
	}
	if err := s.db.Create(&job).Error; err != nil {
		return nil, err
	}

	if err := s.enqueue(job.ID); err != nil {
		return &job, err
	}

	return &job, nil
}

func (s *JobService) GetByUUID(uuid string) (*model.Job, error) {
	var job model.Job
	if err := s.db.Preload("StepLogs", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("uuid = ?", uuid).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *JobService) Execute(ctx context.Context, jobID uint) error {
	var job model.Job
	if err := s.db.First(&job, jobID).Error; err != nil {
		return err
	}
	if job.Status != "pending" {
		logger.Info("skip job with non-pending status", zap.Uint("job_id", job.ID), zap.String("status", job.Status))
		return nil
	}

	claimed := s.db.Model(&model.Job{}).
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
	if err := s.db.First(&doc, job.DocumentID).Error; err != nil {
		return err
	}

	var p model.Pipeline
	if err := s.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
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
		LLM: s.gw,
		DB:  s.db,
	}

	if vectordb.Available() {
		pctx.VectorDB = vectordb.Default()
	}

	if job.CredentialID != nil {
		credSvc := NewCredentialService(s.db)
		cred, apiKey, err := credSvc.GetDecrypted(*job.CredentialID)
		if err != nil {
			if dbErr := s.db.Model(&job).Updates(map[string]any{"status": "failed", "error": "credential error: " + err.Error()}).Error; dbErr != nil {
				logger.Warn("failed to update job status", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}
			return err
		}
		pctx.LLMOverride = &pipeline.LLMOverrideConfig{
			ProviderType: cred.ProviderType,
			APIKey:       apiKey,
			BaseURL:      cred.BaseURL,
			Model:        cred.DefaultModel,
		}
	}

	totalSteps := len(p.Steps)
	if dbErr := s.db.Model(&job).Update("total_steps", totalSteps).Error; dbErr != nil {
		logger.Warn("failed to update total_steps", zap.Uint("job_id", job.ID), zap.Error(dbErr))
	}

	if err := s.engine.Run(ctx, &p, pctx,
		pipeline.WithStartFrom(job.CurrentStep),
		pipeline.WithOnStepStart(func(stepOrder int, processorType string) {
			now := time.Now()
			if dbErr := s.db.Create(&model.JobStepLog{
				JobID:         job.ID,
				StepOrder:     stepOrder,
				ProcessorType: processorType,
				Status:        "running",
				StartedAt:     &now,
			}).Error; dbErr != nil {
				logger.Warn("failed to create step log", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}
			if dbErr := s.db.Model(&job).Update("current_step", stepOrder).Error; dbErr != nil {
				logger.Warn("failed to update current_step", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}
			s.hub.Publish(job.UUID, sse.Event{
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
			if dbErr := s.db.Model(&model.JobStepLog{}).
				Where("job_id = ? AND step_order = ?", job.ID, stepOrder).
				Updates(updates).Error; dbErr != nil {
				logger.Warn("failed to update step log", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}
			if dbErr := s.db.Model(&job).Update("current_step", stepOrder+1).Error; dbErr != nil {
				logger.Warn("failed to update current_step", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}

			s.tracker.Track(job.APIKeyID, log.Tokens)

			evt := sse.Event{
				Type:      "step_end",
				Step:      stepOrder,
				Processor: processorType,
			}
			if stepErr != nil {
				evt.Error = stepErr.Error()
			}
			s.hub.Publish(job.UUID, evt)
		}),
	); err != nil {
		if errors.Is(err, pipeline.ErrPartial) {
			result, _ := json.Marshal(pipeline.BuildResult(pctx))
			if dbErr := s.db.Model(&job).Updates(map[string]any{"status": "partial", "result": result}).Error; dbErr != nil {
				logger.Warn("failed to update job status", zap.Uint("job_id", job.ID), zap.Error(dbErr))
			}
			if persistErr := s.persister.Persist(pctx, doc.ID); persistErr != nil {
				if dbErr := s.db.Model(&job).Updates(map[string]any{"error": "persist results: " + persistErr.Error()}).Error; dbErr != nil {
					logger.Warn("failed to update job error", zap.Uint("job_id", job.ID), zap.Error(dbErr))
				}
			}
			s.hub.Publish(job.UUID, sse.Event{Type: "completed", Data: "partial"})
			s.fireWebhook(job, "job.partial", pipeline.BuildResult(pctx), "")
			return nil
		}
		if dbErr := s.db.Model(&job).Updates(map[string]any{"status": "failed", "error": err.Error()}).Error; dbErr != nil {
			logger.Warn("failed to update job status", zap.Uint("job_id", job.ID), zap.Error(dbErr))
		}
		s.hub.Publish(job.UUID, sse.Event{Type: "failed", Error: err.Error()})
		s.fireWebhook(job, "job.failed", nil, err.Error())
		return err
	}

	if err := s.persister.Persist(pctx, doc.ID); err != nil {
		if dbErr := s.db.Model(&job).Updates(map[string]any{"status": "failed", "error": "persist results: " + err.Error()}).Error; dbErr != nil {
			logger.Warn("failed to update job status", zap.Uint("job_id", job.ID), zap.Error(dbErr))
		}
		s.hub.Publish(job.UUID, sse.Event{Type: "failed", Error: "persist results: " + err.Error()})
		s.fireWebhook(job, "job.failed", nil, "persist results: "+err.Error())
		return err
	}

	result, _ := json.Marshal(pipeline.BuildResult(pctx))
	if dbErr := s.db.Model(&job).Updates(map[string]any{"status": "completed", "result": result}).Error; dbErr != nil {
		logger.Warn("failed to update job status", zap.Uint("job_id", job.ID), zap.Error(dbErr))
	}
	if dbErr := s.db.Model(&doc).Update("status", "completed").Error; dbErr != nil {
		logger.Warn("failed to update doc status", zap.Uint("doc_id", doc.ID), zap.Error(dbErr))
	}
	s.hub.Publish(job.UUID, sse.Event{Type: "completed", Data: pipeline.BuildResult(pctx)})
	s.fireWebhook(job, "job.completed", pipeline.BuildResult(pctx), "")

	if cache.Available() && job.ContentHash != "" {
		cacheKey := "nexus:parse:" + job.ContentHash
		cache.Set(ctx, cacheKey, string(result), 24*time.Hour)
	}

	return nil
}

func (s *JobService) Retry(jobID uint) (*model.Job, error) {
	var job model.Job
	if err := s.db.First(&job, jobID).Error; err != nil {
		return nil, err
	}
	if job.Status != "failed" {
		return nil, errors.New("only failed jobs can be retried")
	}

	s.db.Model(&job).Updates(map[string]any{"status": "pending", "error": ""})

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
	if err := s.db.Where("status IN ?", []string{"pending", "running"}).Find(&jobs).Error; err != nil {
		return err
	}
	if err := s.db.Model(&model.Job{}).
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

func (s *JobService) fireWebhook(job model.Job, event string, result any, errMsg string) {
	if job.CallbackURL == "" {
		return
	}
	go s.webhook.Send(job.CallbackURL, WebhookPayload{
		Event:     event,
		JobID:     job.ID,
		JobUUID:   job.UUID,
		Status:    job.Status,
		Result:    result,
		Error:     errMsg,
		Timestamp: time.Now().Unix(),
	}, config.C.Server.JWTSecret)
}

func (s *JobService) List(page, pageSize int, status string) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64
	q := s.db.Model(&model.Job{})
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
