package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

func QuotaCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyID, exists := c.Get("api_key_id")
		if !exists {
			c.Next()
			return
		}

		var apiKey model.APIKey
		if err := db.First(&apiKey, apiKeyID).Error; err != nil {
			c.Next()
			return
		}

		today := time.Now().Format("2006-01-02")
		month := time.Now().Format("2006-01")

		if apiKey.DailyLimit > 0 || apiKey.DailyTokens > 0 {
			var daily model.APIUsage
			db.Where("api_key_id = ? AND date = ?", apiKey.ID, today).First(&daily)

			if apiKey.DailyLimit > 0 && daily.Requests >= apiKey.DailyLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "daily request limit exceeded"})
				c.Abort()
				return
			}
			if apiKey.DailyTokens > 0 && daily.Tokens >= apiKey.DailyTokens {
				c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "daily token limit exceeded"})
				c.Abort()
				return
			}
		}

		if apiKey.MonthlyLimit > 0 || apiKey.MonthlyTokens > 0 {
			var monthly struct {
				Requests int
				Tokens   int64
			}
			db.Model(&model.APIUsage{}).
				Select("COALESCE(SUM(requests), 0) as requests, COALESCE(SUM(tokens), 0) as tokens").
				Where("api_key_id = ? AND date LIKE ?", apiKey.ID, month+"%").
				Scan(&monthly)

			if apiKey.MonthlyLimit > 0 && monthly.Requests >= apiKey.MonthlyLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "monthly request limit exceeded"})
				c.Abort()
				return
			}
			if apiKey.MonthlyTokens > 0 && monthly.Tokens >= apiKey.MonthlyTokens {
				c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "monthly token limit exceeded"})
				c.Abort()
				return
			}
		}

		// Tenant-level quota
		if tenantID, ok := c.Get("tenant_id"); ok {
			if tid, ok := tenantID.(uint); ok && tid > 0 {
				var tenant model.Tenant
				if err := db.First(&tenant, tid).Error; err == nil && (tenant.MonthlyRequestLimit > 0 || tenant.MonthlyTokenLimit > 0) {
					var tenantMonthly struct {
						Requests int
						Tokens   int64
					}
					db.Model(&model.APIUsage{}).
						Select("COALESCE(SUM(requests), 0) as requests, COALESCE(SUM(tokens), 0) as tokens").
						Where("api_key_id IN (SELECT id FROM api_keys WHERE tenant_id = ?) AND date LIKE ?", tid, month+"%").
						Scan(&tenantMonthly)

					if tenant.MonthlyRequestLimit > 0 && tenantMonthly.Requests >= tenant.MonthlyRequestLimit {
						c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "tenant monthly request limit exceeded"})
						c.Abort()
						return
					}
					if tenant.MonthlyTokenLimit > 0 && tenantMonthly.Tokens >= tenant.MonthlyTokenLimit {
						c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "tenant monthly token limit exceeded"})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}
