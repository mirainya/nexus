package service

import (
	"encoding/json"
	"sort"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type RecommendService struct{ db *gorm.DB }

func NewRecommendService(db *gorm.DB) *RecommendService { return &RecommendService{db: db} }

type RecommendItem struct {
	DocumentID uint     `json:"document_id"`
	SourceURL  string   `json:"source_url"`
	Content    string   `json:"content"`
	Scene      string   `json:"scene"`
	Score      float64  `json:"score"`
	Reason     string   `json:"reason"`
	Tags       []string `json:"tags,omitempty"`
}

func (s *RecommendService) ByScene(scene string, limit int) ([]RecommendItem, error) {
	if limit <= 0 {
		limit = 20
	}

	var jobs []model.Job
	err := s.db.
		Joins("JOIN documents ON documents.id = jobs.document_id").
		Where("documents.type = ? AND jobs.status = ?", "image", "completed").
		Where("jobs.result IS NOT NULL").
		Where("jobs.result LIKE ?", "%image_assessment%").
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
