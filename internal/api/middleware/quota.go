package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/model"
)

func QuotaCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyID, exists := c.Get("api_key_id")
		if !exists {
			c.Next()
			return
		}

		var apiKey model.APIKey
		if err := model.DB().First(&apiKey, apiKeyID).Error; err != nil {
			c.Next()
			return
		}

		if apiKey.DailyLimit == 0 && apiKey.MonthlyLimit == 0 &&
			apiKey.DailyTokens == 0 && apiKey.MonthlyTokens == 0 {
			c.Next()
			return
		}

		today := time.Now().Format("2006-01-02")
		month := time.Now().Format("2006-01")

		if apiKey.DailyLimit > 0 || apiKey.DailyTokens > 0 {
			var daily model.APIUsage
			model.DB().Where("api_key_id = ? AND date = ?", apiKey.ID, today).First(&daily)

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
			model.DB().Model(&model.APIUsage{}).
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

		c.Next()
	}
}
