package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

// List godoc
// @Summary 获取 Webhook 日志
// @Description 分页获取 Webhook 投递日志
// @Tags Webhook
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/webhooks [get]
func (h *WebhookHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	logs, total, err := h.svc.List(page, pageSize)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": logs, "total": total})
}
