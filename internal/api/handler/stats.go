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

func (h *StatsHandler) Dashboard(c *gin.Context) {
	stats, err := h.svc.GetDashboardStats(getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, stats)
}

func (h *StatsHandler) PipelinePerformance(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetPipelinePerformance(days)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}

func (h *StatsHandler) LLMPerformance(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetLLMPerformance(days)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, data)
}

func (h *StatsHandler) ErrorAnalysis(c *gin.Context) {
	days := parseDays(c)
	data, err := h.svc.GetErrorAnalysis(days)
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
