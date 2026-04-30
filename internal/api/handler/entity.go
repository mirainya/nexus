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

func NewEntityHandler(svc *service.EntityService) *EntityHandler {
	return &EntityHandler{svc: svc}
}

// List godoc
// @Summary 获取实体列表
// @Description 分页获取实体，支持按类型和关键词筛选
// @Tags 实体
// @Produce json
// @Param type query string false "实体类型"
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/entities [get]
func (h *EntityHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	entityType := c.Query("type")
	keyword := c.Query("keyword")
	list, total, err := h.svc.List(entityType, keyword, page, pageSize, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": list, "total": total})
}

// Get godoc
// @Summary 获取实体详情
// @Description 获取指定实体
// @Tags 实体
// @Produce json
// @Param id path int true "实体 ID"
// @Success 200 {object} resp.Response
// @Failure 404 {object} resp.Response
// @Security BearerAuth
// @Router /admin/entities/{id} [get]
func (h *EntityHandler) Get(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	e, err := h.svc.GetByID(id, getTenantID(c))
	if err != nil {
		resp.NotFound(c, errors.ErrNotFound)
		return
	}
	resp.Success(c, e)
}

// GetRelations godoc
// @Summary 获取实体关系
// @Description 获取指定实体的所有关系
// @Tags 实体
// @Produce json
// @Param id path int true "实体 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/entities/{id}/relations [get]
func (h *EntityHandler) GetRelations(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	list, err := h.svc.GetRelations(id, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}
