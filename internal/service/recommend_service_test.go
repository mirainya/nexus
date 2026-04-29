package service

import (
	"encoding/json"
	"testing"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/datatypes"
)

func TestRecommendService_Empty(t *testing.T) {
	svc := NewRecommendService(testDB)
	items, err := svc.ByScene("banner", 10)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestRecommendService_ByScene(t *testing.T) {
	svc := NewRecommendService(testDB)

	doc := model.Document{UUID: "rec-doc-1", Type: "image", Content: "img.jpg", SourceURL: "https://example.com/img.jpg", Status: "completed"}
	testDB.Create(&doc)

	result := map[string]any{
		"extras": map[string]any{
			"image_assessment": map[string]any{
				"use_cases": []any{
					map[string]any{
						"scene":    "banner",
						"suitable": true,
						"score":    0.9,
						"reason":   "good composition",
						"tags":     []any{"wide", "colorful"},
					},
					map[string]any{
						"scene":    "avatar",
						"suitable": false,
						"score":    0.2,
						"reason":   "too wide",
					},
				},
			},
		},
	}
	resultJSON, _ := json.Marshal(result)

	job := model.Job{
		UUID:       "rec-job-1",
		DocumentID: doc.ID,
		PipelineID: 1,
		Status:     "completed",
		Result:     datatypes.JSON(resultJSON),
	}
	testDB.Create(&job)

	items, err := svc.ByScene("banner", 10)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Score != 0.9 {
		t.Errorf("expected score 0.9, got %f", items[0].Score)
	}
	if items[0].SourceURL != "https://example.com/img.jpg" {
		t.Errorf("expected source URL, got %q", items[0].SourceURL)
	}
	if len(items[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(items[0].Tags))
	}

	// avatar scene should not match (suitable=false)
	items2, _ := svc.ByScene("avatar", 10)
	if len(items2) != 0 {
		t.Errorf("expected 0 items for avatar, got %d", len(items2))
	}
}

func TestRecommendService_SortedByScore(t *testing.T) {
	svc := NewRecommendService(testDB)

	for i, score := range []float64{0.5, 0.9, 0.7} {
		doc := model.Document{UUID: "rec-sort-doc-" + string(rune('a'+i)), Type: "image", Content: "img.jpg", Status: "completed"}
		testDB.Create(&doc)

		result := map[string]any{
			"extras": map[string]any{
				"image_assessment": map[string]any{
					"use_cases": []any{
						map[string]any{"scene": "social", "suitable": true, "score": score, "reason": "ok"},
					},
				},
			},
		}
		resultJSON, _ := json.Marshal(result)
		job := model.Job{UUID: "rec-sort-job-" + string(rune('a'+i)), DocumentID: doc.ID, PipelineID: 1, Status: "completed", Result: datatypes.JSON(resultJSON)}
		testDB.Create(&job)
	}

	items, _ := svc.ByScene("social", 10)
	if len(items) < 3 {
		t.Fatalf("expected >= 3 items, got %d", len(items))
	}
	for i := 1; i < len(items); i++ {
		if items[i].Score > items[i-1].Score {
			t.Errorf("items not sorted by score: %f > %f at index %d", items[i].Score, items[i-1].Score, i)
		}
	}
}
