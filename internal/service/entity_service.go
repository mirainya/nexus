package service

import (
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
)

type EntityService struct{}

func NewEntityService() *EntityService { return &EntityService{} }

func (s *EntityService) List(entityType string, keyword string, page, pageSize int) ([]model.Entity, int64, error) {
	var list []model.Entity
	var total int64
	q := model.DB().Model(&model.Entity{})
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
	err := model.DB().First(&e, id).Error
	return &e, err
}

func (s *EntityService) GetRelations(entityID uint) ([]model.Relation, error) {
	var list []model.Relation
	err := model.DB().Where("from_entity_id = ? OR to_entity_id = ?", entityID, entityID).
		Preload("FromEntity").Preload("ToEntity").Find(&list).Error
	return list, err
}
