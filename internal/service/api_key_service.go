package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type APIKeyService struct{ db *gorm.DB }

func NewAPIKeyService(db *gorm.DB) *APIKeyService { return &APIKeyService{db: db} }

type APIKeyCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	TenantID      uint   `json:"tenant_id" binding:"required"`
	ExpiresAt     string `json:"expires_at"`
	DailyLimit    int    `json:"daily_limit"`
	MonthlyLimit  int    `json:"monthly_limit"`
	DailyTokens   int64  `json:"daily_tokens"`
	MonthlyTokens int64  `json:"monthly_tokens"`
}

type APIKeyUpdateRequest struct {
	Name          string `json:"name"`
	Active        *bool  `json:"active"`
	ExpiresAt     string `json:"expires_at"`
	DailyLimit    *int   `json:"daily_limit"`
	MonthlyLimit  *int   `json:"monthly_limit"`
	DailyTokens   *int64 `json:"daily_tokens"`
	MonthlyTokens *int64 `json:"monthly_tokens"`
}

type APIKeyResponse struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Key           string     `json:"key"`
	TenantID      uint       `json:"tenant_id"`
	Active        bool       `json:"active"`
	ExpiresAt     *time.Time `json:"expires_at"`
	DailyLimit    int        `json:"daily_limit"`
	MonthlyLimit  int        `json:"monthly_limit"`
	DailyTokens   int64      `json:"daily_tokens"`
	MonthlyTokens int64      `json:"monthly_tokens"`
	CreatedAt     time.Time  `json:"created_at"`
}

type APIKeyUsageResponse struct {
	APIKeyID uint   `json:"api_key_id"`
	Date     string `json:"date"`
	Requests int    `json:"requests"`
	Tokens   int64  `json:"tokens"`
}

func generateKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "nxk_" + hex.EncodeToString(b)
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func toAPIKeyResponse(k *model.APIKey, showFull bool) APIKeyResponse {
	key := maskAPIKey(k.Key)
	if showFull {
		key = k.Key
	}
	return APIKeyResponse{
		ID:            k.ID,
		Name:          k.Name,
		Key:           key,
		TenantID:      k.TenantID,
		Active:        k.Active,
		ExpiresAt:     k.ExpiresAt,
		DailyLimit:    k.DailyLimit,
		MonthlyLimit:  k.MonthlyLimit,
		DailyTokens:   k.DailyTokens,
		MonthlyTokens: k.MonthlyTokens,
		CreatedAt:     k.CreatedAt,
	}
}

func (s *APIKeyService) Create(req APIKeyCreateRequest) (*APIKeyResponse, error) {
	ak := model.APIKey{
		Name:          req.Name,
		Key:           generateKey(),
		Active:        true,
		TenantID:      req.TenantID,
		DailyLimit:    req.DailyLimit,
		MonthlyLimit:  req.MonthlyLimit,
		DailyTokens:   req.DailyTokens,
		MonthlyTokens: req.MonthlyTokens,
	}
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			t, err = time.Parse("2006-01-02", req.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("invalid expires_at format")
			}
		}
		ak.ExpiresAt = &t
	}
	if err := s.db.Create(&ak).Error; err != nil {
		return nil, err
	}
	resp := toAPIKeyResponse(&ak, true)
	return &resp, nil
}

func (s *APIKeyService) List() ([]APIKeyResponse, error) {
	var keys []model.APIKey
	if err := s.db.Order("id DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	result := make([]APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		result = append(result, toAPIKeyResponse(&k, false))
	}
	return result, nil
}

func (s *APIKeyService) Update(id uint, req APIKeyUpdateRequest) (*APIKeyResponse, error) {
	var ak model.APIKey
	if err := s.db.First(&ak, id).Error; err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if req.DailyLimit != nil {
		updates["daily_limit"] = *req.DailyLimit
	}
	if req.MonthlyLimit != nil {
		updates["monthly_limit"] = *req.MonthlyLimit
	}
	if req.DailyTokens != nil {
		updates["daily_tokens"] = *req.DailyTokens
	}
	if req.MonthlyTokens != nil {
		updates["monthly_tokens"] = *req.MonthlyTokens
	}
	if req.ExpiresAt != "" {
		if req.ExpiresAt == "null" {
			updates["expires_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, req.ExpiresAt)
			if err != nil {
				t, err = time.Parse("2006-01-02", req.ExpiresAt)
				if err != nil {
					return nil, fmt.Errorf("invalid expires_at format")
				}
			}
			updates["expires_at"] = &t
		}
	}
	if len(updates) > 0 {
		if err := s.db.Model(&ak).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	s.db.First(&ak, id)
	resp := toAPIKeyResponse(&ak, false)
	return &resp, nil
}

func (s *APIKeyService) Delete(id uint) error {
	return s.db.Delete(&model.APIKey{}, id).Error
}

func (s *APIKeyService) GetUsage(apiKeyID uint, days int) ([]APIKeyUsageResponse, error) {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	var usages []model.APIUsage
	if err := s.db.Where("api_key_id = ? AND date >= ?", apiKeyID, since).
		Order("date DESC").Find(&usages).Error; err != nil {
		return nil, err
	}
	result := make([]APIKeyUsageResponse, 0, len(usages))
	for _, u := range usages {
		result = append(result, APIKeyUsageResponse{
			APIKeyID: u.APIKeyID,
			Date:     u.Date,
			Requests: u.Requests,
			Tokens:   u.Tokens,
		})
	}
	return result, nil
}
