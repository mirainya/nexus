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

func (h *PipelineHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

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
