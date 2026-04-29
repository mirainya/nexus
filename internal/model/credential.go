package model

type Credential struct {
	BaseModel
	APIKeyID     uint   `gorm:"index" json:"api_key_id"`
	Name         string `gorm:"type:varchar(100)" json:"name"`
	ProviderType string `gorm:"type:varchar(50)" json:"provider_type"`
	BaseURL      string `gorm:"type:varchar(500)" json:"base_url"`
	EncryptedKey string `gorm:"type:text" json:"-"`
	DefaultModel string `gorm:"type:varchar(100)" json:"default_model"`
	Active       bool   `gorm:"default:true" json:"active"`
}
