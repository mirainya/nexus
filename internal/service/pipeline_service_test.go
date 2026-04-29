package service

import (
	"testing"

	"github.com/mirainya/nexus/internal/model"
)

func TestPipelineService_Create(t *testing.T) {
	svc := NewPipelineService(testDB)
	p := &model.Pipeline{Name: "test-create", Active: true}
	if err := svc.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID == 0 {
		t.Error("expected ID > 0")
	}
}

func TestPipelineService_GetByID(t *testing.T) {
	svc := NewPipelineService(testDB)
	p := &model.Pipeline{Name: "test-get"}
	testDB.Create(p)

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "test-get" {
		t.Errorf("expected name 'test-get', got %q", got.Name)
	}
}

func TestPipelineService_GetByID_NotFound(t *testing.T) {
	svc := NewPipelineService(testDB)
	_, err := svc.GetByID(99999)
	if err == nil {
		t.Error("expected error for nonexistent pipeline")
	}
}

func TestPipelineService_List(t *testing.T) {
	svc := NewPipelineService(testDB)
	list, err := svc.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list == nil {
		t.Error("expected non-nil list")
	}
}

func TestPipelineService_Update(t *testing.T) {
	svc := NewPipelineService(testDB)
	p := &model.Pipeline{Name: "before-update"}
	testDB.Create(p)

	p.Name = "after-update"
	if err := svc.Update(p); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := svc.GetByID(p.ID)
	if got.Name != "after-update" {
		t.Errorf("expected 'after-update', got %q", got.Name)
	}
}

func TestPipelineService_Delete(t *testing.T) {
	svc := NewPipelineService(testDB)
	p := &model.Pipeline{Name: "to-delete"}
	testDB.Create(p)

	if err := svc.Delete(p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := svc.GetByID(p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestPipelineService_StepCRUD(t *testing.T) {
	svc := NewPipelineService(testDB)
	p := &model.Pipeline{Name: "step-test"}
	testDB.Create(p)

	// Create steps
	s1 := &model.PipelineStep{PipelineID: p.ID, ProcessorType: "ocr", SortOrder: 0}
	s2 := &model.PipelineStep{PipelineID: p.ID, ProcessorType: "llm_extract", SortOrder: 1}
	if err := svc.CreateStep(s1); err != nil {
		t.Fatalf("create step 1: %v", err)
	}
	if err := svc.CreateStep(s2); err != nil {
		t.Fatalf("create step 2: %v", err)
	}

	// Verify steps loaded via GetByID
	got, _ := svc.GetByID(p.ID)
	if len(got.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(got.Steps))
	}

	// Update step
	s1.ProcessorType = "face"
	if err := svc.UpdateStep(s1); err != nil {
		t.Fatalf("update step: %v", err)
	}

	// Reorder
	if err := svc.ReorderSteps(p.ID, []uint{s2.ID, s1.ID}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	got, _ = svc.GetByID(p.ID)
	if got.Steps[0].ID != s2.ID {
		t.Errorf("expected step %d first after reorder, got %d", s2.ID, got.Steps[0].ID)
	}

	// Delete step
	if err := svc.DeleteStep(s1.ID); err != nil {
		t.Fatalf("delete step: %v", err)
	}
	got, _ = svc.GetByID(p.ID)
	if len(got.Steps) != 1 {
		t.Errorf("expected 1 step after delete, got %d", len(got.Steps))
	}
}
