package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type APIKeyHandler struct {
	svc *service.APIKeyService
}

func NewAPIKeyHandler(svc *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

// Create godoc
// @Summary 创建 API Key
// @Description 创建新的 API Key，用于外部接口认证
// @Tags API Key
// @Accept json
// @Produce json
// @Param request body service.APIKeyCreateRequest true "创建请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req service.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	req.TenantID = getTenantID(c)
	key, err := h.svc.Create(req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, key)
}

// List godoc
// @Summary 获取 API Key 列表
// @Description 获取所有 API Key
// @Tags API Key
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	list, err := h.svc.List(getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Update godoc
// @Summary 更新 API Key
// @Description 更新 API Key 的名称、状态等信息
// @Tags API Key
// @Accept json
// @Produce json
// @Param id path int true "API Key ID"
// @Param request body service.APIKeyUpdateRequest true "更新请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/api-keys/{id} [put]
func (h *APIKeyHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req service.APIKeyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	key, err := h.svc.Update(id, req, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, key)
}

// Delete godoc
// @Summary 删除 API Key
// @Description 删除指定 API Key
// @Tags API Key
// @Produce json
// @Param id path int true "API Key ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/api-keys/{id} [delete]
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(id, getTenantID(c)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

// Usage godoc
// @Summary 查询 API Key 用量
// @Description 查询指定 API Key 的使用统计
// @Tags API Key
// @Produce json
// @Param id path int true "API Key ID"
// @Param days query int false "统计天数" default(30)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/api-keys/{id}/usage [get]
func (h *APIKeyHandler) Usage(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	days := 30
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	usage, err := h.svc.GetUsage(id, days, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, usage)
}
