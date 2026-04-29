package service

import "github.com/mirainya/nexus/internal/model"

type PromptService struct{}

func NewPromptService() *PromptService { return &PromptService{} }

func (s *PromptService) Create(p *model.PromptTemplate) error {
	return model.DB().Create(p).Error
}

func (s *PromptService) GetByID(id uint) (*model.PromptTemplate, error) {
	var p model.PromptTemplate
	err := model.DB().First(&p, id).Error
	return &p, err
}

func (s *PromptService) List() ([]model.PromptTemplate, error) {
	var list []model.PromptTemplate
	err := model.DB().Order("id DESC").Find(&list).Error
	return list, err
}

func (s *PromptService) Update(p *model.PromptTemplate) error {
	p.Version++
	return model.DB().Save(p).Error
}

func (s *PromptService) Delete(id uint) error {
	return model.DB().Delete(&model.PromptTemplate{}, id).Error
}
