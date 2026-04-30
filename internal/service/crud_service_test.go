package service

import (
	"testing"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
)

// --- EntityService Tests ---

func TestEntityService_List(t *testing.T) {
	svc := NewEntityService(testDB)

	testDB.Create(&model.Entity{UUID: "ent-list-1", Type: "person", Name: "Alice", SourceID: 1})
	testDB.Create(&model.Entity{UUID: "ent-list-2", Type: "org", Name: "Acme", SourceID: 1})

	config.C.Database.Driver = "sqlite"
	list, total, err := svc.List("", "", 1, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total < 2 {
		t.Errorf("expected total >= 2, got %d", total)
	}
	if len(list) == 0 {
		t.Error("expected non-empty list")
	}
}

func TestEntityService_List_FilterByType(t *testing.T) {
	svc := NewEntityService(testDB)
	config.C.Database.Driver = "sqlite"

	testDB.Create(&model.Entity{UUID: "ent-filter-1", Type: "person", Name: "Bob", SourceID: 1})
	testDB.Create(&model.Entity{UUID: "ent-filter-2", Type: "location", Name: "Tokyo", SourceID: 1})

	list, _, err := svc.List("location", "", 1, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, e := range list {
		if e.Type != "location" {
			t.Errorf("expected type 'location', got %q", e.Type)
		}
	}
}

func TestEntityService_List_Keyword(t *testing.T) {
	svc := NewEntityService(testDB)
	config.C.Database.Driver = "sqlite"

	testDB.Create(&model.Entity{UUID: "ent-kw-1", Type: "person", Name: "Charlie", SourceID: 1})

	list, _, err := svc.List("", "Charlie", 1, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) == 0 {
		t.Error("expected to find Charlie")
	}
}

func TestEntityService_GetByID(t *testing.T) {
	svc := NewEntityService(testDB)
	e := model.Entity{UUID: "ent-get-1", Type: "person", Name: "Dave", SourceID: 1}
	testDB.Create(&e)

	got, err := svc.GetByID(e.ID, 0)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Dave" {
		t.Errorf("expected 'Dave', got %q", got.Name)
	}
}

func TestEntityService_GetRelations(t *testing.T) {
	svc := NewEntityService(testDB)

	e1 := model.Entity{UUID: "ent-rel-1", Type: "person", Name: "Eve", SourceID: 1}
	e2 := model.Entity{UUID: "ent-rel-2", Type: "person", Name: "Frank", SourceID: 1}
	testDB.Create(&e1)
	testDB.Create(&e2)
	testDB.Create(&model.Relation{UUID: "rel-1", FromEntityID: e1.ID, ToEntityID: e2.ID, Type: "knows", SourceID: 1})

	rels, err := svc.GetRelations(e1.ID, 0)
	if err != nil {
		t.Fatalf("get relations: %v", err)
	}
	if len(rels) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(rels))
	}
	if rels[0].Type != "knows" {
		t.Errorf("expected type 'knows', got %q", rels[0].Type)
	}
}

// --- PromptService Tests ---

func TestPromptService_Create(t *testing.T) {
	svc := NewPromptService(testDB)
	p := &model.PromptTemplate{Name: "test-prompt", Content: "Extract entities from: {{text}}"}
	if err := svc.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID == 0 {
		t.Error("expected ID > 0")
	}
}

func TestPromptService_GetByID(t *testing.T) {
	svc := NewPromptService(testDB)
	p := &model.PromptTemplate{Name: "get-prompt", Content: "hello"}
	testDB.Create(p)

	got, err := svc.GetByID(p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Content != "hello" {
		t.Errorf("expected content 'hello', got %q", got.Content)
	}
}

func TestPromptService_List(t *testing.T) {
	svc := NewPromptService(testDB)
	list, err := svc.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list == nil {
		t.Error("expected non-nil list")
	}
}

func TestPromptService_Update(t *testing.T) {
	svc := NewPromptService(testDB)
	p := &model.PromptTemplate{Name: "update-prompt", Content: "v1", Version: 1}
	testDB.Create(p)

	p.Content = "v2"
	if err := svc.Update(p); err != nil {
		t.Fatalf("update: %v", err)
	}
	if p.Version != 2 {
		t.Errorf("expected version 2, got %d", p.Version)
	}

	got, _ := svc.GetByID(p.ID)
	if got.Content != "v2" {
		t.Errorf("expected content 'v2', got %q", got.Content)
	}
}

func TestPromptService_Delete(t *testing.T) {
	svc := NewPromptService(testDB)
	p := &model.PromptTemplate{Name: "delete-prompt", Content: "bye"}
	testDB.Create(p)

	if err := svc.Delete(p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := svc.GetByID(p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// --- ReviewService Tests ---

func TestReviewService_List(t *testing.T) {
	svc := NewReviewService(testDB)

	testDB.Create(&model.Review{Status: "pending"})
	testDB.Create(&model.Review{Status: "approved"})

	list, total, err := svc.List("", 1, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total < 2 {
		t.Errorf("expected total >= 2, got %d", total)
	}
	if len(list) == 0 {
		t.Error("expected non-empty list")
	}
}

func TestReviewService_List_FilterByStatus(t *testing.T) {
	svc := NewReviewService(testDB)

	list, _, err := svc.List("pending", 1, 100, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, r := range list {
		if r.Status != "pending" {
			t.Errorf("expected status 'pending', got %q", r.Status)
		}
	}
}

func TestReviewService_Approve(t *testing.T) {
	svc := NewReviewService(testDB)
	r := model.Review{Status: "pending"}
	testDB.Create(&r)

	if err := svc.Approve(r.ID, "admin", 0); err != nil {
		t.Fatalf("approve: %v", err)
	}

	var updated model.Review
	testDB.First(&updated, r.ID)
	if updated.Status != "approved" {
		t.Errorf("expected 'approved', got %q", updated.Status)
	}
	if updated.Reviewer != "admin" {
		t.Errorf("expected reviewer 'admin', got %q", updated.Reviewer)
	}
}

func TestReviewService_Reject(t *testing.T) {
	svc := NewReviewService(testDB)
	r := model.Review{Status: "pending"}
	testDB.Create(&r)

	if err := svc.Reject(r.ID, "admin", 0); err != nil {
		t.Fatalf("reject: %v", err)
	}

	var updated model.Review
	testDB.First(&updated, r.ID)
	if updated.Status != "rejected" {
		t.Errorf("expected 'rejected', got %q", updated.Status)
	}
}

func TestReviewService_Modify(t *testing.T) {
	svc := NewReviewService(testDB)
	r := model.Review{Status: "pending"}
	testDB.Create(&r)

	data := map[string]any{"name": "corrected"}
	if err := svc.Modify(r.ID, "admin", data, 0); err != nil {
		t.Fatalf("modify: %v", err)
	}

	var updated model.Review
	testDB.First(&updated, r.ID)
	if updated.Status != "modified" {
		t.Errorf("expected 'modified', got %q", updated.Status)
	}
}
