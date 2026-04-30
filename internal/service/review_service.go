package service

import (
	"encoding/json"

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

func (s *ReviewService) Approve(id uint, reviewer string) error {
	return s.db.Model(&model.Review{}).Where("id = ?", id).
		Updates(map[string]any{"status": "approved", "reviewer": reviewer}).Error
}

func (s *ReviewService) Reject(id uint, reviewer string) error {
	return s.db.Model(&model.Review{}).Where("id = ?", id).
		Updates(map[string]any{"status": "rejected", "reviewer": reviewer}).Error
}

func (s *ReviewService) Modify(id uint, reviewer string, data map[string]any) error {
	modified, _ := json.Marshal(data)
	return s.db.Model(&model.Review{}).Where("id = ?", id).
		Updates(map[string]any{"status": "modified", "reviewer": reviewer, "modified_data": modified}).Error
}
