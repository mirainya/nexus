package service

import (
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type PromptService struct{ db *gorm.DB }

func NewPromptService(db *gorm.DB) *PromptService { return &PromptService{db: db} }

func (s *PromptService) Create(p *model.PromptTemplate) error {
	return s.db.Create(p).Error
}

func (s *PromptService) GetByID(id uint) (*model.PromptTemplate, error) {
	var p model.PromptTemplate
	err := s.db.First(&p, id).Error
	return &p, err
}

func (s *PromptService) List() ([]model.PromptTemplate, error) {
	var list []model.PromptTemplate
	err := s.db.Order("id DESC").Find(&list).Error
	return list, err
}

func (s *PromptService) Update(p *model.PromptTemplate) error {
	p.Version++
	return s.db.Save(p).Error
}

func (s *PromptService) Delete(id uint) error {
	return s.db.Delete(&model.PromptTemplate{}, id).Error
}
