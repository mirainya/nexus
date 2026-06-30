package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/errors"
	"gorm.io/gorm"
)

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}
		if key == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if key == "" {
			resp.Unauthorized(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		var apiKey model.APIKey
		if err := db.Where("key = ? AND active = ?", key, true).First(&apiKey).Error; err != nil {
			resp.Unauthorized(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			resp.Unauthorized(c, errors.WithMessage(errors.ErrUnauthorized, "api key expired"))
			c.Abort()
			return
		}

		c.Set("api_key_id", apiKey.ID)
		c.Set("api_key_name", apiKey.Name)
		c.Next()
	}
}
