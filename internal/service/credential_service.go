package service

import (
	"fmt"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/crypto"
)

type CredentialService struct{}

func NewCredentialService() *CredentialService { return &CredentialService{} }

type CredentialCreateRequest struct {
	APIKeyID     uint   `json:"api_key_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	ProviderType string `json:"provider_type" binding:"required"`
	BaseURL      string `json:"base_url" binding:"required"`
	APIKey       string `json:"api_key" binding:"required"`
	DefaultModel string `json:"default_model"`
}

type CredentialUpdateRequest struct {
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	DefaultModel string `json:"default_model"`
	Active       *bool  `json:"active"`
}

type CredentialResponse struct {
	ID           uint   `json:"id"`
	APIKeyID     uint   `json:"api_key_id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	MaskedKey    string `json:"api_key"`
	DefaultModel string `json:"default_model"`
	Active       bool   `json:"active"`
}

func toResponse(c *model.Credential, maskedKey string) CredentialResponse {
	return CredentialResponse{
		ID:           c.ID,
		APIKeyID:     c.APIKeyID,
		Name:         c.Name,
		ProviderType: c.ProviderType,
		BaseURL:      c.BaseURL,
		MaskedKey:    maskedKey,
		DefaultModel: c.DefaultModel,
		Active:       c.Active,
	}
}

func (s *CredentialService) Create(req CredentialCreateRequest) (*CredentialResponse, error) {
	encrypted, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt api key: %w", err)
	}

	cred := model.Credential{
		APIKeyID:     req.APIKeyID,
		Name:         req.Name,
		ProviderType: req.ProviderType,
		BaseURL:      req.BaseURL,
		EncryptedKey: encrypted,
		DefaultModel: req.DefaultModel,
		Active:       true,
	}
	if err := model.DB().Create(&cred).Error; err != nil {
		return nil, err
	}

	resp := toResponse(&cred, crypto.MaskKey(req.APIKey))
	return &resp, nil
}

func (s *CredentialService) List(apiKeyID uint) ([]CredentialResponse, error) {
	var creds []model.Credential
	q := model.DB().Order("id DESC")
	if apiKeyID > 0 {
		q = q.Where("api_key_id = ?", apiKeyID)
	}
	if err := q.Find(&creds).Error; err != nil {
		return nil, err
	}

	result := make([]CredentialResponse, 0, len(creds))
	for _, c := range creds {
		masked := "****"
		if plain, err := crypto.Decrypt(c.EncryptedKey); err == nil {
			masked = crypto.MaskKey(plain)
		}
		result = append(result, toResponse(&c, masked))
	}
	return result, nil
}

func (s *CredentialService) Update(id uint, req CredentialUpdateRequest) (*CredentialResponse, error) {
	var cred model.Credential
	if err := model.DB().First(&cred, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ProviderType != "" {
		updates["provider_type"] = req.ProviderType
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.DefaultModel != "" {
		updates["default_model"] = req.DefaultModel
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if req.APIKey != "" {
		encrypted, err := crypto.Encrypt(req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt api key: %w", err)
		}
		updates["encrypted_key"] = encrypted
	}

	if len(updates) > 0 {
		if err := model.DB().Model(&cred).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	model.DB().First(&cred, id)
	masked := "****"
	if plain, err := crypto.Decrypt(cred.EncryptedKey); err == nil {
		masked = crypto.MaskKey(plain)
	}
	resp := toResponse(&cred, masked)
	return &resp, nil
}

func (s *CredentialService) Delete(id uint) error {
	return model.DB().Delete(&model.Credential{}, id).Error
}

func (s *CredentialService) GetDecrypted(id uint) (*model.Credential, string, error) {
	var cred model.Credential
	if err := model.DB().First(&cred, id).Error; err != nil {
		return nil, "", err
	}
	if !cred.Active {
		return nil, "", fmt.Errorf("credential %d is inactive", id)
	}
	plain, err := crypto.Decrypt(cred.EncryptedKey)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt credential: %w", err)
	}
	return &cred, plain, nil
}
