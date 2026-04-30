package model

type LLMProvider struct {
	BaseModel
	Name           string  `gorm:"type:varchar(50);uniqueIndex" json:"name"`
	DisplayName    string  `gorm:"type:varchar(100)" json:"display_name"`
	BaseURL        string  `gorm:"type:varchar(500)" json:"base_url"`
	EncryptedKey   string  `gorm:"type:text" json:"-"`
	DefaultModel   string  `gorm:"type:varchar(100)" json:"default_model"`
	InputPrice     float64 `gorm:"default:0" json:"input_price"`
	OutputPrice    float64 `gorm:"default:0" json:"output_price"`
	MaxConcurrency int     `gorm:"default:10" json:"max_concurrency"`
	Active         bool    `json:"active"`
	IsDefault      bool    `gorm:"default:false" json:"is_default"`
}
