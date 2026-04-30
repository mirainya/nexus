package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/errors"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			if t := c.Query("token"); t != "" {
				auth = "Bearer " + t
			}
		}
		if !strings.HasPrefix(auth, "Bearer ") {
			resp.Unauthorized(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.C.Server.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			resp.Unauthorized(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			resp.Unauthorized(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])
		if tid, ok := claims["tenant_id"].(float64); ok {
			c.Set("tenant_id", uint(tid))
		}
		c.Next()
	}
}

func GenerateToken(userID uint, username string, tenantID *uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	if tenantID != nil {
		claims["tenant_id"] = *tenantID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.Server.JWTSecret))
}
