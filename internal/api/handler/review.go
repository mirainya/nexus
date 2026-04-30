package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	list, total, err := h.svc.List(status, page, pageSize, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": list, "total": total})
}

func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	reviewer, _ := c.Get("username")
	if err := h.svc.Approve(id, reviewer.(string)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

func (h *ReviewHandler) Reject(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	reviewer, _ := c.Get("username")
	if err := h.svc.Reject(id, reviewer.(string)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

func (h *ReviewHandler) Modify(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	reviewer, _ := c.Get("username")
	if err := h.svc.Modify(id, reviewer.(string), req); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}
