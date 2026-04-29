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
	result, err := h.svc.Parse(c.Request.Context(), req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, result)
}
