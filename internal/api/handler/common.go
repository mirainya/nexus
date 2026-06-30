package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

// getAPIKeyID returns the calling API key's id (set by APIKeyAuth middleware),
// or nil for admin JWT requests. Services treat nil as "no key filter" (admin sees all).
func getAPIKeyID(c *gin.Context) *uint {
	if v, ok := c.Get("api_key_id"); ok {
		if id, ok := v.(uint); ok {
			return &id
		}
	}
	return nil
}
