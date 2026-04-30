package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type SearchHandler struct {
	svc *service.SearchService
}

func NewSearchHandler(svc *service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search godoc
// @Summary 语义搜索
// @Description 基于向量和关键词的混合搜索，返回相关文档和实体
// @Tags 搜索
// @Accept json
// @Produce json
// @Param request body service.SearchRequest true "搜索请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Failure 500 {object} resp.Response
// @Security ApiKeyAuth
// @Router /v1/search [post]
func (h *SearchHandler) Search(c *gin.Context) {
	var req service.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	req.TenantID = getTenantID(c)
	result, err := h.svc.Search(c.Request.Context(), req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, result)
}
