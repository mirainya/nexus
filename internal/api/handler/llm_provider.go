package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type LLMProviderHandler struct {
	svc *service.LLMProviderService
}

func NewLLMProviderHandler(svc *service.LLMProviderService) *LLMProviderHandler {
	return &LLMProviderHandler{svc: svc}
}

// List godoc
// @Summary 获取 LLM Provider 列表
// @Description 获取所有已配置的 LLM 提供商
// @Tags LLM Provider
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers [get]
func (h *LLMProviderHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Create godoc
// @Summary 创建 LLM Provider
// @Description 添加新的 LLM 提供商配置
// @Tags LLM Provider
// @Accept json
// @Produce json
// @Param request body service.LLMProviderCreateRequest true "创建请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers [post]
func (h *LLMProviderHandler) Create(c *gin.Context) {
	var req service.LLMProviderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	result, err := h.svc.Create(req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, result)
}

// Update godoc
// @Summary 更新 LLM Provider
// @Description 更新 LLM 提供商配置
// @Tags LLM Provider
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Param request body service.LLMProviderUpdateRequest true "更新请求"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers/{id} [put]
func (h *LLMProviderHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req service.LLMProviderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	result, err := h.svc.Update(id, req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, result)
}

// Delete godoc
// @Summary 删除 LLM Provider
// @Description 删除指定 LLM 提供商
// @Tags LLM Provider
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers/{id} [delete]
func (h *LLMProviderHandler) Delete(c *gin.Context) {
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

// SetDefault godoc
// @Summary 设置默认 Provider
// @Description 将指定 Provider 设为默认
// @Tags LLM Provider
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers/{id}/default [put]
func (h *LLMProviderHandler) SetDefault(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.SetDefault(id); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

// ListModels godoc
// @Summary 获取 Provider 模型列表
// @Description 获取指定 Provider 支持的模型
// @Tags LLM Provider
// @Produce json
// @Param name path string true "Provider 名称"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/llm-providers/{name}/models [get]
func (h *LLMProviderHandler) ListModels(c *gin.Context) {
	name := c.Param("name")
	models, err := h.svc.ListModels(c.Request.Context(), name)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, models)
}
