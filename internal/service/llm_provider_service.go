package service

import (
	"context"
	"fmt"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/crypto"
	"gorm.io/gorm"
)

type LLMProviderService struct {
	db *gorm.DB
	gw *llm.Gateway
}

func NewLLMProviderService(db *gorm.DB, gw *llm.Gateway) *LLMProviderService {
	return &LLMProviderService{db: db, gw: gw}
}

type LLMProviderCreateRequest struct {
	Name           string  `json:"name" binding:"required"`
	DisplayName    string  `json:"display_name"`
	BaseURL        string  `json:"base_url"`
	APIKey         string  `json:"api_key" binding:"required"`
	DefaultModel   string  `json:"default_model"`
	InputPrice     float64 `json:"input_price"`
	OutputPrice    float64 `json:"output_price"`
	MaxConcurrency int     `json:"max_concurrency"`
	Active         bool    `json:"active"`
	IsDefault      bool    `json:"is_default"`
}

type LLMProviderUpdateRequest struct {
	DisplayName    string  `json:"display_name"`
	BaseURL        string  `json:"base_url"`
	APIKey         string  `json:"api_key"`
	DefaultModel   string  `json:"default_model"`
	InputPrice     float64 `json:"input_price"`
	OutputPrice    float64 `json:"output_price"`
	MaxConcurrency int     `json:"max_concurrency"`
	Active         bool    `json:"active"`
}

type LLMProviderResponse struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	DisplayName    string  `json:"display_name"`
	BaseURL        string  `json:"base_url"`
	MaskedKey      string  `json:"api_key"`
	DefaultModel   string  `json:"default_model"`
	InputPrice     float64 `json:"input_price"`
	OutputPrice    float64 `json:"output_price"`
	MaxConcurrency int     `json:"max_concurrency"`
	Active         bool    `json:"active"`
	IsDefault      bool    `json:"is_default"`
}

func toLLMProviderResponse(p *model.LLMProvider) LLMProviderResponse {
	masked := "****"
	if plain, err := crypto.Decrypt(p.EncryptedKey); err == nil && plain != "" {
		masked = crypto.MaskKey(plain)
	}
	return LLMProviderResponse{
		ID:             p.ID,
		Name:           p.Name,
		DisplayName:    p.DisplayName,
		BaseURL:        p.BaseURL,
		MaskedKey:      masked,
		DefaultModel:   p.DefaultModel,
		InputPrice:     p.InputPrice,
		OutputPrice:    p.OutputPrice,
		MaxConcurrency: p.MaxConcurrency,
		Active:         p.Active,
		IsDefault:      p.IsDefault,
	}
}

func (s *LLMProviderService) List() ([]LLMProviderResponse, error) {
	var list []model.LLMProvider
	if err := s.db.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]LLMProviderResponse, 0, len(list))
	for i := range list {
		result = append(result, toLLMProviderResponse(&list[i]))
	}
	return result, nil
}

func (s *LLMProviderService) Create(req LLMProviderCreateRequest) (*LLMProviderResponse, error) {
	encrypted, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt api key: %w", err)
	}

	p := model.LLMProvider{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		BaseURL:        req.BaseURL,
		EncryptedKey:   encrypted,
		DefaultModel:   req.DefaultModel,
		InputPrice:     req.InputPrice,
		OutputPrice:    req.OutputPrice,
		MaxConcurrency: req.MaxConcurrency,
		Active:         req.Active,
		IsDefault:      req.IsDefault,
	}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	s.gw.LoadFromDB()
	resp := toLLMProviderResponse(&p)
	return &resp, nil
}

func (s *LLMProviderService) Update(id uint, req LLMProviderUpdateRequest) (*LLMProviderResponse, error) {
	updates := map[string]any{
		"display_name":    req.DisplayName,
		"base_url":        req.BaseURL,
		"default_model":   req.DefaultModel,
		"input_price":     req.InputPrice,
		"output_price":    req.OutputPrice,
		"max_concurrency": req.MaxConcurrency,
		"active":          req.Active,
	}
	if req.APIKey != "" {
		encrypted, err := crypto.Encrypt(req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt api key: %w", err)
		}
		updates["encrypted_key"] = encrypted
	}
	if err := s.db.Model(&model.LLMProvider{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.gw.LoadFromDB()

	var p model.LLMProvider
	s.db.First(&p, id)
	resp := toLLMProviderResponse(&p)
	return &resp, nil
}

func (s *LLMProviderService) Delete(id uint) error {
	if err := s.db.Delete(&model.LLMProvider{}, id).Error; err != nil {
		return err
	}
	s.gw.LoadFromDB()
	return nil
}

func (s *LLMProviderService) SetDefault(id uint) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		tx.Model(&model.LLMProvider{}).Where("is_default = ?", true).Update("is_default", false)
		return tx.Model(&model.LLMProvider{}).Where("id = ?", id).Update("is_default", true).Error
	})
	if err != nil {
		return err
	}
	s.gw.LoadFromDB()
	return nil
}

func (s *LLMProviderService) ListModels(ctx context.Context, providerName string) ([]llm.ModelInfo, error) {
	return s.gw.ListModels(ctx, providerName)
}
