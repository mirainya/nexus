package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	config.C = &config.Config{}
	logger.Init()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}
	db.AutoMigrate(
		&model.User{},
		&model.APIKey{},
		&model.PromptTemplate{},
		&model.Pipeline{},
		&model.PipelineStep{},
		&model.Document{},
		&model.Entity{},
		&model.Relation{},
		&model.Job{},
		&model.JobStepLog{},
		&model.Review{},
		&model.LLMProvider{},
	)
	model.SetDB(db)

	os.Exit(m.Run())
}

func seedPipeline(t *testing.T) model.Pipeline {
	t.Helper()
	p := model.Pipeline{Name: t.Name(), Active: true}
	if err := model.DB().Create(&p).Error; err != nil {
		t.Fatalf("seed pipeline: %v", err)
	}
	return p
}

func TestJobService_Submit(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job, err := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "hello world",
		PipelineID: p.ID,
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if job.UUID == "" {
		t.Error("expected UUID to be set")
	}
	if job.Status != "pending" {
		t.Errorf("expected status pending, got %s", job.Status)
	}
	if job.ContentHash == "" {
		t.Error("expected content hash to be set")
	}

	var doc model.Document
	if err := model.DB().First(&doc, job.DocumentID).Error; err != nil {
		t.Fatalf("document not created: %v", err)
	}
	if doc.Content != "hello world" {
		t.Errorf("expected doc content 'hello world', got %q", doc.Content)
	}
}

func TestJobService_Submit_CacheHit(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job1, err := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "cache test",
		PipelineID: p.ID,
	})
	if err != nil {
		t.Fatalf("submit 1: %v", err)
	}
	model.DB().Model(job1).Update("status", "completed")

	job2, err := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "cache test",
		PipelineID: p.ID,
	})
	if err != nil {
		t.Fatalf("submit 2: %v", err)
	}
	if job2.ID != job1.ID {
		t.Errorf("expected cache hit (same job ID %d), got new job %d", job1.ID, job2.ID)
	}
}

func TestJobService_Submit_SkipCache(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job1, err := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "skip cache test",
		PipelineID: p.ID,
	})
	if err != nil {
		t.Fatalf("submit 1: %v", err)
	}
	model.DB().Model(job1).Update("status", "completed")

	job2, err := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "skip cache test",
		PipelineID: p.ID,
		SkipCache:  true,
	})
	if err != nil {
		t.Fatalf("submit 2: %v", err)
	}
	if job2.ID == job1.ID {
		t.Error("expected new job when skip_cache=true")
	}
}

func TestJobService_GetByUUID(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job, _ := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "get test",
		PipelineID: p.ID,
	})

	found, err := svc.GetByUUID(job.UUID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if found.ID != job.ID {
		t.Errorf("expected job %d, got %d", job.ID, found.ID)
	}
}

func TestJobService_GetByUUID_NotFound(t *testing.T) {
	svc := NewJobService(nil)
	_, err := svc.GetByUUID("nonexistent-uuid")
	if err == nil {
		t.Error("expected error for nonexistent UUID")
	}
}

func TestJobService_Retry_OnlyFailed(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job, _ := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "retry test",
		PipelineID: p.ID,
	})

	_, err := svc.Retry(job.ID)
	if err == nil {
		t.Error("expected error when retrying non-failed job")
	}

	model.DB().Model(job).Update("status", "failed")
	retried, err := svc.Retry(job.ID)
	if err != nil {
		t.Fatalf("retry failed job: %v", err)
	}
	if retried.Status != "pending" {
		t.Errorf("expected status pending after retry, got %s", retried.Status)
	}
}

func TestJobService_List(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	for i := 0; i < 5; i++ {
		svc.Submit(JobSubmitRequest{
			Type:       "text",
			Content:    "list test " + string(rune('a'+i)),
			PipelineID: p.ID,
		})
	}

	jobs, total, err := svc.List(1, 3, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs per page, got %d", len(jobs))
	}
	if total < 5 {
		t.Errorf("expected total >= 5, got %d", total)
	}
}

func TestJobService_List_FilterByStatus(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	job, _ := svc.Submit(JobSubmitRequest{
		Type:       "text",
		Content:    "filter test",
		PipelineID: p.ID,
	})
	model.DB().Model(job).Update("status", "completed")

	jobs, _, err := svc.List(1, 100, "completed")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, j := range jobs {
		if j.Status != "completed" {
			t.Errorf("expected all completed, got %s", j.Status)
		}
	}
}

func TestJobService_PersistResults(t *testing.T) {
	p := seedPipeline(t)
	svc := NewJobService(nil)

	doc := model.Document{UUID: "persist-test", Type: "text", Content: "test", Status: "pending"}
	model.DB().Create(&doc)

	_ = p

	pctx := &pipeline.ProcessorContext{
		Entities: []pipeline.EntityData{
			{Type: "person", Name: "Alice", Confidence: 0.95},
			{Type: "person", Name: "Bob", Confidence: 0.90},
		},
		Relations: []pipeline.RelationData{
			{From: "Alice", To: "Bob", Type: "knows", Confidence: 0.85},
		},
	}

	err := svc.persistResults(pctx, doc.ID)
	if err != nil {
		t.Fatalf("persistResults: %v", err)
	}

	var entities []model.Entity
	model.DB().Where("source_id = ?", doc.ID).Find(&entities)
	if len(entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(entities))
	}

	var relations []model.Relation
	model.DB().Where("source_id = ?", doc.ID).Find(&relations)
	if len(relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(relations))
	}
	if relations[0].Type != "knows" {
		t.Errorf("expected relation type 'knows', got %q", relations[0].Type)
	}

	var reviews []model.Review
	model.DB().Where("entity_id IN ?", []uint{entities[0].ID, entities[1].ID}).Find(&reviews)
	if len(reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(reviews))
	}
}

func TestJobService_PersistResults_ExistingEntity(t *testing.T) {
	doc := model.Document{UUID: "persist-existing-test", Type: "text", Content: "test", Status: "pending"}
	model.DB().Create(&doc)

	svc := NewJobService(nil)

	aliasesJSON, _ := json.Marshal([]string{"Robert"})
	existing := model.Entity{
		UUID:       "existing-entity",
		Type:       "person",
		Name:       "Bob",
		Aliases:    aliasesJSON,
		Confidence: 0.80,
		SourceID:   doc.ID,
	}
	model.DB().Create(&existing)

	pctx := &pipeline.ProcessorContext{
		Entities: []pipeline.EntityData{
			{
				Type:       "person",
				Name:       "Bob",
				Aliases:    []string{"Bobby"},
				Confidence: 0.95,
				Attributes: map[string]any{"existing_id": float64(existing.ID)},
			},
		},
	}

	err := svc.persistResults(pctx, doc.ID)
	if err != nil {
		t.Fatalf("persistResults: %v", err)
	}

	var updated model.Entity
	model.DB().First(&updated, existing.ID)
	if updated.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", updated.Confidence)
	}

	var aliases []string
	json.Unmarshal(updated.Aliases, &aliases)
	found := false
	for _, a := range aliases {
		if a == "Bobby" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected alias 'Bobby' to be merged, got %v", aliases)
	}
}
