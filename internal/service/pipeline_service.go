package service

import (
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type PipelineService struct{}

func NewPipelineService() *PipelineService { return &PipelineService{} }

func (s *PipelineService) Create(p *model.Pipeline) error {
	return model.DB().Create(p).Error
}

func (s *PipelineService) GetByID(id uint) (*model.Pipeline, error) {
	var p model.Pipeline
	err := model.DB().Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Preload("Steps.PromptTemplate").First(&p, id).Error
	return &p, err
}

func (s *PipelineService) List() ([]model.Pipeline, error) {
	var list []model.Pipeline
	err := model.DB().Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Find(&list).Error
	return list, err
}

func (s *PipelineService) Update(p *model.Pipeline) error {
	return model.DB().Save(p).Error
}

func (s *PipelineService) Delete(id uint) error {
	return model.DB().Transaction(func(tx *gorm.DB) error {
		tx.Where("pipeline_id = ?", id).Delete(&model.PipelineStep{})
		return tx.Delete(&model.Pipeline{}, id).Error
	})
}

func (s *PipelineService) CreateStep(step *model.PipelineStep) error {
	return model.DB().Create(step).Error
}

func (s *PipelineService) UpdateStep(step *model.PipelineStep) error {
	return model.DB().Save(step).Error
}

func (s *PipelineService) DeleteStep(id uint) error {
	return model.DB().Delete(&model.PipelineStep{}, id).Error
}

func (s *PipelineService) ReorderSteps(pipelineID uint, stepIDs []uint) error {
	return model.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range stepIDs {
			if err := tx.Model(&model.PipelineStep{}).Where("id = ? AND pipeline_id = ?", id, pipelineID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
