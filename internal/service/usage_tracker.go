package service

import (
	"time"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UsageTracker struct{ db *gorm.DB }

func NewUsageTracker(db *gorm.DB) *UsageTracker { return &UsageTracker{db: db} }

func (t *UsageTracker) Track(apiKeyID *uint, tokens int) {
	if apiKeyID == nil || tokens <= 0 {
		return
	}
	today := time.Now().Format("2006-01-02")

	var sql string
	if config.C.Database.Driver == "sqlite" {
		sql = `INSERT INTO api_usages (api_key_id, date, requests, tokens, created_at, updated_at)
			VALUES (?, ?, 1, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT(api_key_id, date) DO UPDATE SET
				requests = api_usages.requests + 1,
				tokens = api_usages.tokens + excluded.tokens,
				updated_at = CURRENT_TIMESTAMP`
	} else {
		sql = `INSERT INTO api_usages (api_key_id, date, requests, tokens, created_at, updated_at)
			VALUES (?, ?, 1, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (api_key_id, date) DO UPDATE SET
				requests = api_usages.requests + 1,
				tokens = api_usages.tokens + EXCLUDED.tokens,
				updated_at = CURRENT_TIMESTAMP`
	}

	if err := t.db.Exec(sql, *apiKeyID, today, tokens).Error; err != nil {
		logger.Warn("usage track failed", zap.Uint("api_key_id", *apiKeyID), zap.Error(err))
		t.trackFallback(*apiKeyID, today, tokens)
	}
}

func (t *UsageTracker) trackFallback(apiKeyID uint, today string, tokens int) {
	var usage model.APIUsage
	result := t.db.Where("api_key_id = ? AND date = ?", apiKeyID, today).First(&usage)
	if result.Error != nil {
		usage = model.APIUsage{APIKeyID: apiKeyID, Date: today, Requests: 1, Tokens: int64(tokens)}
		t.db.Create(&usage)
	} else {
		t.db.Model(&usage).Updates(map[string]any{
			"requests": gorm.Expr("requests + 1"),
			"tokens":   gorm.Expr("tokens + ?", tokens),
		})
	}
}
