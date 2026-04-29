package service

import (
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"gorm.io/gorm"
)

type EntityService struct{ db *gorm.DB }

func NewEntityService(db *gorm.DB) *EntityService { return &EntityService{db: db} }

func (s *EntityService) List(entityType string, keyword string, page, pageSize int) ([]model.Entity, int64, error) {
	var list []model.Entity
	var total int64
	q := s.db.Model(&model.Entity{})
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

func (s *EntityService) GetByID(id uint) (*model.Entity, error) {
	var e model.Entity
	err := s.db.First(&e, id).Error
	return &e, err
}

func (s *EntityService) GetRelations(entityID uint) ([]model.Relation, error) {
	var list []model.Relation
	err := s.db.Where("from_entity_id = ? OR to_entity_id = ?", entityID, entityID).
		Preload("FromEntity").Preload("ToEntity").Find(&list).Error
	return list, err
}
