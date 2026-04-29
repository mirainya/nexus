package service

import (
	"time"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type UsageTracker struct{ db *gorm.DB }

func NewUsageTracker(db *gorm.DB) *UsageTracker { return &UsageTracker{db: db} }

func (t *UsageTracker) Track(apiKeyID *uint, tokens int) {
	if apiKeyID == nil || tokens <= 0 {
		return
	}
	today := time.Now().Format("2006-01-02")
	var usage model.APIUsage
	result := t.db.Where("api_key_id = ? AND date = ?", *apiKeyID, today).First(&usage)
	if result.Error != nil {
		usage = model.APIUsage{APIKeyID: *apiKeyID, Date: today, Requests: 1, Tokens: int64(tokens)}
		t.db.Create(&usage)
	} else {
		t.db.Model(&usage).Updates(map[string]any{
			"requests": usage.Requests + 1,
			"tokens":   usage.Tokens + int64(tokens),
		})
	}
}
