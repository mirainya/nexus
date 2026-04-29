package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type EntityHandler struct {
	svc *service.EntityService
}

func NewEntityHandler() *EntityHandler {
	return &EntityHandler{svc: service.NewEntityService()}
}

func (h *EntityHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	entityType := c.Query("type")
	keyword := c.Query("keyword")
	list, total, err := h.svc.List(entityType, keyword, page, pageSize)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": list, "total": total})
}

func (h *EntityHandler) Get(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	e, err := h.svc.GetByID(id)
	if err != nil {
		resp.NotFound(c, errors.ErrNotFound)
		return
	}
	resp.Success(c, e)
}

func (h *EntityHandler) GetRelations(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	list, err := h.svc.GetRelations(id)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}
