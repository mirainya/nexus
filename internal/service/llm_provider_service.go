package service

import (
	"context"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type LLMProviderService struct{}

func NewLLMProviderService() *LLMProviderService { return &LLMProviderService{} }

func (s *LLMProviderService) List() ([]model.LLMProvider, error) {
	var list []model.LLMProvider
	err := model.DB().Order("id ASC").Find(&list).Error
	return list, err
}

func (s *LLMProviderService) Create(p *model.LLMProvider) error {
	if err := model.DB().Create(p).Error; err != nil {
		return err
	}
	llm.G.LoadFromDB()
	return nil
}

func (s *LLMProviderService) Update(p *model.LLMProvider) error {
	updates := map[string]any{
		"display_name":    p.DisplayName,
		"base_url":        p.BaseURL,
		"default_model":   p.DefaultModel,
		"input_price":     p.InputPrice,
		"output_price":    p.OutputPrice,
		"max_concurrency": p.MaxConcurrency,
		"active":          p.Active,
	}
	if p.APIKey != "" {
		updates["api_key"] = p.APIKey
	}
	if err := model.DB().Model(&model.LLMProvider{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
		return err
	}
	llm.G.LoadFromDB()
	return nil
}

func (s *LLMProviderService) Delete(id uint) error {
	if err := model.DB().Delete(&model.LLMProvider{}, id).Error; err != nil {
		return err
	}
	llm.G.LoadFromDB()
	return nil
}

func (s *LLMProviderService) SetDefault(id uint) error {
	err := model.DB().Transaction(func(tx *gorm.DB) error {
		tx.Model(&model.LLMProvider{}).Where("is_default = ?", true).Update("is_default", false)
		return tx.Model(&model.LLMProvider{}).Where("id = ?", id).Update("is_default", true).Error
	})
	if err != nil {
		return err
	}
	llm.G.LoadFromDB()
	return nil
}

func (s *LLMProviderService) ListModels(ctx context.Context, providerName string) ([]llm.ModelInfo, error) {
	return llm.G.ListModels(ctx, providerName)
}
