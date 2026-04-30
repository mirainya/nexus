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

func getTenantID(c *gin.Context) uint {
	if v, ok := c.Get("tenant_id"); ok {
		switch tid := v.(type) {
		case uint:
			return tid
		case float64:
			return uint(tid)
		}
	}
	return 0
}
