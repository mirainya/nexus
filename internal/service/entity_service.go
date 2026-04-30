package service

import (
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"gorm.io/gorm"
)

type EntityService struct{ db *gorm.DB }

func NewEntityService(db *gorm.DB) *EntityService { return &EntityService{db: db} }

func (s *EntityService) List(entityType string, keyword string, page, pageSize int, tenantID uint) ([]model.Entity, int64, error) {
	var list []model.Entity
	var total int64
	q := s.db.Model(&model.Entity{})
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if entityType != "" {
		q = q.Where("type = ?", entityType)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		if config.C.Database.Driver == "sqlite" {
			q = q.Where("name LIKE ? OR aliases LIKE ?", like, like)
		} else {
			q = q.Where("name ILIKE ? OR aliases::text ILIKE ?", like, like)
		}
	}
	q.Count(&total)
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *EntityService) GetByID(id uint, tenantID uint) (*model.Entity, error) {
	var e model.Entity
	q := s.db.Where("id = ?", id)
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.First(&e).Error
	return &e, err
}

func (s *EntityService) GetRelations(entityID uint, tenantID uint) ([]model.Relation, error) {
	var list []model.Relation
	q := s.db.Where("from_entity_id = ? OR to_entity_id = ?", entityID, entityID)
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	err := q.Preload("FromEntity").Preload("ToEntity").Find(&list).Error
	return list, err
}
