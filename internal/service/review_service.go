package service

import (
	"encoding/json"
	"fmt"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type ReviewService struct{ db *gorm.DB }

func NewReviewService(db *gorm.DB) *ReviewService { return &ReviewService{db: db} }

func (s *ReviewService) List(status string, page, pageSize int, tenantID uint) ([]model.Review, int64, error) {
	var list []model.Review
	var total int64
	q := s.db.Model(&model.Review{})
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *ReviewService) Approve(id uint, reviewer string, tenantID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var review model.Review
		q := tx.Where("id = ?", id)
		if tenantID > 0 {
			q = q.Where("tenant_id = ?", tenantID)
		}
		if err := q.First(&review).Error; err != nil {
			return err
		}
		if review.Status != "pending" {
			return fmt.Errorf("review is already %s", review.Status)
		}

		if err := tx.Model(&review).Updates(map[string]any{
			"status": "approved", "reviewer": reviewer,
		}).Error; err != nil {
			return err
		}

		if review.DocumentID != nil {
			tx.Model(&model.Entity{}).
				Where("source_id = ? AND confirmed = false AND deleted_at IS NULL", *review.DocumentID).
				Update("confirmed", true)
		} else if review.EntityID != nil {
			tx.Model(&model.Entity{}).Where("id = ?", *review.EntityID).Update("confirmed", true)
		}
		return nil
	})
}

func (s *ReviewService) Reject(id uint, reviewer string, tenantID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var review model.Review
		q := tx.Where("id = ?", id)
		if tenantID > 0 {
			q = q.Where("tenant_id = ?", tenantID)
		}
		if err := q.First(&review).Error; err != nil {
			return err
		}
		if review.Status != "pending" {
			return fmt.Errorf("review is already %s", review.Status)
		}

		if err := tx.Model(&review).Updates(map[string]any{
			"status": "rejected", "reviewer": reviewer,
		}).Error; err != nil {
			return err
		}

		if review.DocumentID != nil {
			var entityIDs []uint
			var items []struct {
				EntityID float64 `json:"entity_id"`
			}
			if json.Unmarshal(review.OriginalData, &items) == nil {
				for _, item := range items {
					if uint(item.EntityID) > 0 {
						entityIDs = append(entityIDs, uint(item.EntityID))
					}
				}
			}
			if len(entityIDs) == 0 {
				tx.Model(&model.Entity{}).
					Where("source_id = ? AND confirmed = false AND deleted_at IS NULL", *review.DocumentID).
					Pluck("id", &entityIDs)
			}

			if len(entityIDs) > 0 {
				tx.Where("id IN ?", entityIDs).Delete(&model.Entity{})
				tx.Where("from_entity_id IN ? OR to_entity_id IN ?", entityIDs, entityIDs).Delete(&model.Relation{})
			}
		} else if review.EntityID != nil {
			tx.Delete(&model.Entity{}, *review.EntityID)
			tx.Where("from_entity_id = ? OR to_entity_id = ?", *review.EntityID, *review.EntityID).Delete(&model.Relation{})
		}
		return nil
	})
}

func (s *ReviewService) Modify(id uint, reviewer string, data map[string]any, tenantID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var review model.Review
		q := tx.Where("id = ?", id)
		if tenantID > 0 {
			q = q.Where("tenant_id = ?", tenantID)
		}
		if err := q.First(&review).Error; err != nil {
			return err
		}
		if review.Status != "pending" {
			return fmt.Errorf("review is already %s", review.Status)
		}

		modified, _ := json.Marshal(data)
		if err := tx.Model(&review).Updates(map[string]any{
			"status": "modified", "reviewer": reviewer, "modified_data": modified,
		}).Error; err != nil {
			return err
		}

		entities, _ := data["entities"].([]any)
		for _, item := range entities {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			eid, _ := entry["entity_id"].(float64)
			if eid == 0 {
				continue
			}
			updates := map[string]any{"confirmed": true}
			if name, ok := entry["name"].(string); ok && name != "" {
				updates["name"] = name
			}
			if typ, ok := entry["type"].(string); ok && typ != "" {
				updates["type"] = typ
			}
			if aliases, ok := entry["aliases"]; ok {
				aliasJSON, _ := json.Marshal(aliases)
				updates["aliases"] = aliasJSON
			}
			if attrs, ok := entry["attributes"]; ok {
				attrsJSON, _ := json.Marshal(attrs)
				updates["attributes"] = attrsJSON
			}
			tx.Model(&model.Entity{}).Where("id = ?", uint(eid)).Updates(updates)
		}

		return nil
	})
}
