package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/sse"
)

type SSEHandler struct {
	hub *sse.Hub
}

func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) Stream(c *gin.Context) {
	jobUUID := c.Param("id")
	if jobUUID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.hub.Subscribe(jobUUID)
	defer h.hub.Unsubscribe(jobUUID, ch)

	ctx := c.Request.Context()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case evt, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent("message", string(evt.JSON()))
			if evt.Type == "completed" || evt.Type == "failed" {
				return false
			}
			return true
		}
	})

	fmt.Fprint(c.Writer, "event: close\ndata: {}\n\n")
	c.Writer.Flush()
}
