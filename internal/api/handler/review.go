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

// List godoc
// @Summary 获取审核列表
// @Description 分页获取审核记录，支持按状态筛选
// @Tags 审核
// @Produce json
// @Param status query string false "状态筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/reviews [get]
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

// Approve godoc
// @Summary 通过审核
// @Description 通过指定审核记录
// @Tags 审核
// @Produce json
// @Param id path int true "审核 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/reviews/{id}/approve [put]
func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	reviewer, _ := c.Get("username")
	if err := h.svc.Approve(id, reviewer.(string), getTenantID(c)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

// Reject godoc
// @Summary 拒绝审核
// @Description 拒绝指定审核记录
// @Tags 审核
// @Produce json
// @Param id path int true "审核 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/reviews/{id}/reject [put]
func (h *ReviewHandler) Reject(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	reviewer, _ := c.Get("username")
	if err := h.svc.Reject(id, reviewer.(string), getTenantID(c)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

// Modify godoc
// @Summary 修改审核内容
// @Description 修改审核记录的提取结果
// @Tags 审核
// @Accept json
// @Produce json
// @Param id path int true "审核 ID"
// @Param request body object true "修改内容"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/reviews/{id}/modify [put]
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
	if err := h.svc.Modify(id, reviewer.(string), req, getTenantID(c)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}
