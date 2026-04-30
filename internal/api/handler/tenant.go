package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// Create godoc
// @Summary 创建租户
// @Description 创建新租户
// @Tags 租户
// @Accept json
// @Produce json
// @Param request body service.TenantCreateRequest true "创建请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/tenants [post]
func (h *TenantHandler) Create(c *gin.Context) {
	var req service.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	t, err := h.svc.Create(req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, t)
}

// List godoc
// @Summary 获取租户列表
// @Description 获取所有租户
// @Tags 租户
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/tenants [get]
func (h *TenantHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Update godoc
// @Summary 更新租户
// @Description 更新租户名称、状态、配额等
// @Tags 租户
// @Accept json
// @Produce json
// @Param id path int true "租户 ID"
// @Param request body service.TenantUpdateRequest true "更新请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/tenants/{id} [put]
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req service.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	t, err := h.svc.Update(id, req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, t)
}

// Delete godoc
// @Summary 删除租户
// @Description 删除指定租户
// @Tags 租户
// @Produce json
// @Param id path int true "租户 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/tenants/{id} [delete]
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}
