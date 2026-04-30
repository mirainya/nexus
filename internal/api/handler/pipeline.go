package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type PipelineHandler struct {
	svc *service.PipelineService
}

func NewPipelineHandler(svc *service.PipelineService) *PipelineHandler {
	return &PipelineHandler{svc: svc}
}

// Create godoc
// @Summary 创建 Pipeline
// @Description 创建新的处理流水线
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param request body model.Pipeline true "Pipeline 配置"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines [post]
func (h *PipelineHandler) Create(c *gin.Context) {
	var p model.Pipeline
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
// @Summary 获取 Pipeline 详情
// @Description 获取指定 Pipeline 及其步骤
// @Tags Pipeline
// @Produce json
// @Param id path int true "Pipeline ID"
// @Success 200 {object} resp.Response
// @Failure 404 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id} [get]
func (h *PipelineHandler) Get(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	p, err := h.svc.GetByID(id)
	if err != nil {
		resp.NotFound(c, errors.ErrPipelineNotFound)
		return
	}
	resp.Success(c, p)
}

// List godoc
// @Summary 获取 Pipeline 列表
// @Description 获取所有 Pipeline
// @Tags Pipeline
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines [get]
func (h *PipelineHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Update godoc
// @Summary 更新 Pipeline
// @Description 更新 Pipeline 配置
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path int true "Pipeline ID"
// @Param request body model.Pipeline true "Pipeline 配置"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id} [put]
func (h *PipelineHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var p model.Pipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	p.ID = id
	if err := h.svc.Update(&p); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, p)
}

// Delete godoc
// @Summary 删除 Pipeline
// @Description 删除指定 Pipeline
// @Tags Pipeline
// @Produce json
// @Param id path int true "Pipeline ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id} [delete]
func (h *PipelineHandler) Delete(c *gin.Context) {
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

// CreateStep godoc
// @Summary 创建 Pipeline 步骤
// @Description 为指定 Pipeline 添加处理步骤
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path int true "Pipeline ID"
// @Param request body model.PipelineStep true "步骤配置"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id}/steps [post]
func (h *PipelineHandler) CreateStep(c *gin.Context) {
	pipelineID, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var step model.PipelineStep
	if err := c.ShouldBindJSON(&step); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	step.PipelineID = pipelineID
	if err := h.svc.CreateStep(&step); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, step)
}

// UpdateStep godoc
// @Summary 更新 Pipeline 步骤
// @Description 更新指定步骤配置
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path int true "Pipeline ID"
// @Param step_id path int true "步骤 ID"
// @Param request body model.PipelineStep true "步骤配置"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id}/steps/{step_id} [put]
func (h *PipelineHandler) UpdateStep(c *gin.Context) {
	_, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	stepID, err := resp.ParseUintParam(c, "step_id")
	if err != nil {
		return
	}
	var step model.PipelineStep
	if err := c.ShouldBindJSON(&step); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	step.ID = stepID
	if err := h.svc.UpdateStep(&step); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, step)
}

// DeleteStep godoc
// @Summary 删除 Pipeline 步骤
// @Description 删除指定步骤
// @Tags Pipeline
// @Produce json
// @Param id path int true "Pipeline ID"
// @Param step_id path int true "步骤 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id}/steps/{step_id} [delete]
func (h *PipelineHandler) DeleteStep(c *gin.Context) {
	stepID, err := resp.ParseUintParam(c, "step_id")
	if err != nil {
		return
	}
	if err := h.svc.DeleteStep(stepID); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}

// ReorderSteps godoc
// @Summary 重排 Pipeline 步骤
// @Description 调整步骤执行顺序
// @Tags Pipeline
// @Accept json
// @Produce json
// @Param id path int true "Pipeline ID"
// @Param request body object true "步骤 ID 列表 {step_ids: [1,2,3]}"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/pipelines/{id}/steps/reorder [put]
func (h *PipelineHandler) ReorderSteps(c *gin.Context) {
	pipelineID, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req struct {
		StepIDs []uint `json:"step_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	if err := h.svc.ReorderSteps(pipelineID, req.StepIDs); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}
