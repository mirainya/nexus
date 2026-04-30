package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type TenantService struct{ db *gorm.DB }

func NewTenantService(db *gorm.DB) *TenantService { return &TenantService{db: db} }

type TenantCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type TenantUpdateRequest struct {
	Name   string `json:"name"`
	Active *bool  `json:"active"`
}

type TenantResponse struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func toTenantResponse(t *model.Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		UUID:      t.UUID,
		Name:      t.Name,
		Active:    t.Active,
		CreatedAt: t.CreatedAt,
	}
}

func (s *TenantService) Create(req TenantCreateRequest) (*TenantResponse, error) {
	t := model.Tenant{
		UUID:   uuid.New().String(),
		Name:   req.Name,
		Active: true,
	}
	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}
	resp := toTenantResponse(&t)
	return &resp, nil
}

func (s *TenantService) List() ([]TenantResponse, error) {
	var tenants []model.Tenant
	if err := s.db.Order("id").Find(&tenants).Error; err != nil {
		return nil, err
	}
	result := make([]TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		result = append(result, toTenantResponse(&t))
	}
	return result, nil
}

func (s *TenantService) Update(id uint, req TenantUpdateRequest) (*TenantResponse, error) {
	var t model.Tenant
	if err := s.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if len(updates) > 0 {
		if err := s.db.Model(&t).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	s.db.First(&t, id)
	resp := toTenantResponse(&t)
	return &resp, nil
}

func (s *TenantService) Delete(id uint) error {
	return s.db.Delete(&model.Tenant{}, id).Error
}
