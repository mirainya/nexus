package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type GraphHandler struct {
	svc *service.GraphService
}

func NewGraphHandler(svc *service.GraphService) *GraphHandler {
	return &GraphHandler{svc: svc}
}

// GetGraph godoc
// @Summary 获取知识图谱
// @Description 获取实体关系图谱数据
// @Tags 图谱
// @Produce json
// @Param limit query int false "节点数量限制" default(200)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/graph [get]
func (h *GraphHandler) GetGraph(c *gin.Context) {
	limit := 200
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	data, err := h.svc.GetGraphData(limit, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}
