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

func (h *PromptHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

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
