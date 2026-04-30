package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type StatsHandler struct {
	svc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// Dashboard godoc
// @Summary 仪表盘统计
// @Description 获取总览统计数据
// @Tags 统计
// @Produce json
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/stats [get]
func (h *StatsHandler) Dashboard(c *gin.Context) {
	stats, err := h.svc.GetDashboardStats(getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, stats)
}

// PipelinePerformance godoc
// @Summary Pipeline 性能统计
// @Description 获取 Pipeline 处理性能数据
// @Tags 统计
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/stats/pipeline-performance [get]
func (h *StatsHandler) PipelinePerformance(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetPipelinePerformance(days, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}

// LLMPerformance godoc
// @Summary LLM 性能统计
// @Description 获取 LLM 调用性能数据
// @Tags 统计
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/stats/llm-performance [get]
func (h *StatsHandler) LLMPerformance(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetLLMPerformance(days, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}

// ErrorAnalysis godoc
// @Summary 错误分析
// @Description 获取错误统计和分析数据
// @Tags 统计
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/stats/errors [get]
func (h *StatsHandler) ErrorAnalysis(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetErrorAnalysis(days, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}

func parseDays(c *gin.Context) int {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 7
	}
	return days
}
