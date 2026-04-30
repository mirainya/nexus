package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type PromptHandler struct {
	svc *service.PromptService
}

func NewPromptHandler(svc *service.PromptService) *PromptHandler {
	return &PromptHandler{svc: svc}
}

// Create godoc
// @Summary 创建提示词模板
// @Description 创建新的提示词模板
// @Tags 提示词
// @Accept json
// @Produce json
// @Param request body model.PromptTemplate true "模板内容"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/prompts [post]
func (h *PromptHandler) Create(c *gin.Context) {
	var p model.PromptTemplate
	if err := c.ShouldBindJSON(&p); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	if err := h.svc.Create(&p); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, p)
}

// Get godoc
// @Summary 获取提示词模板详情
// @Description 获取指定提示词模板
// @Tags 提示词
// @Produce json
// @Param id path int true "模板 ID"
// @Success 200 {object} resp.Response
// @Failure 404 {object} resp.Response
// @Security BearerAuth
// @Router /admin/prompts/{id} [get]
func (h *PromptHandler) Get(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	p, err := h.svc.GetByID(id)
	if err != nil {
		resp.NotFound(c, errors.ErrPromptNotFound)
		return
	}
	resp.Success(c, p)
}

// List godoc
// @Summary 获取提示词模板列表
// @Description 获取所有提示词模板
// @Tags 提示词
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/prompts [get]
func (h *PromptHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Update godoc
// @Summary 更新提示词模板
// @Description 更新指定提示词模板
// @Tags 提示词
// @Accept json
// @Produce json
// @Param id path int true "模板 ID"
// @Param request body model.PromptTemplate true "模板内容"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/prompts/{id} [put]
func (h *PromptHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	p, err := h.svc.GetByID(id)
	if err != nil {
		resp.NotFound(c, errors.ErrPromptNotFound)
		return
	}
	if err := c.ShouldBindJSON(p); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	p.ID = id
	if err := h.svc.Update(p); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, p)
}

// Delete godoc
// @Summary 删除提示词模板
// @Description 删除指定提示词模板
// @Tags 提示词
// @Produce json
// @Param id path int true "模板 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/prompts/{id} [delete]
func (h *PromptHandler) Delete(c *gin.Context) {
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
