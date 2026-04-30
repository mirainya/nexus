package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type ParseHandler struct {
	svc *service.ParseService
}

func NewParseHandler(svc *service.ParseService) *ParseHandler {
	return &ParseHandler{svc: svc}
}

// Parse godoc
// @Summary 同步解析文档
// @Description 同步调用 LLM 解析文档内容，提取实体和关系
// @Tags 解析
// @Accept json
// @Produce json
// @Param request body service.ParseRequest true "解析请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Failure 500 {object} resp.Response
// @Security ApiKeyAuth
// @Router /v1/parse [post]
func (h *ParseHandler) Parse(c *gin.Context) {
	var req service.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	if req.Type == "image" && req.SourceURL == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "source_url is required for image type"))
		return
	}
	if req.Type != "image" && req.Content == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "content is required for text type"))
		return
	}
	if apiKeyID, exists := c.Get("api_key_id"); exists {
		if id, ok := apiKeyID.(uint); ok {
			req.APIKeyID = &id
		}
	}
	result, err := h.svc.Parse(c.Request.Context(), req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, result)
}
