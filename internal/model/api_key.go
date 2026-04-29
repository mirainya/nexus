package model

import "time"

type APIKey struct {
	BaseModel
	Name          string     `gorm:"type:varchar(100)" json:"name"`
	Key           string     `gorm:"type:varchar(64);uniqueIndex" json:"key"`
	Active        bool       `gorm:"default:true" json:"active"`
	ExpiresAt     *time.Time `json:"expires_at"`
	DailyLimit    int        `gorm:"default:0" json:"daily_limit"`
	MonthlyLimit  int        `gorm:"default:0" json:"monthly_limit"`
	DailyTokens   int64      `gorm:"default:0" json:"daily_tokens"`
	MonthlyTokens int64      `gorm:"default:0" json:"monthly_tokens"`
}
