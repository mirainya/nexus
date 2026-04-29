package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type JobHandler struct {
	svc *service.JobService
}

func NewJobHandler(client *asynq.Client) *JobHandler {
	return &JobHandler{svc: service.NewJobService(client)}
}

func (h *JobHandler) Submit(c *gin.Context) {
	var req service.JobSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	if req.Type == "image" && req.SourceURL == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "source_url is required for image type"))
		return
	}
	if req.Type != "image" && req.Content == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "content is required for text type"))
		return
	}
	if apiKeyID, exists := c.Get("api_key_id"); exists {
		if id, ok := apiKeyID.(uint); ok {
			req.APIKeyID = &id
		}
	}
	job, err := h.svc.Submit(req)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	result := gin.H{"job_id": job.UUID, "status": job.Status}
	if job.Status == "completed" && job.Result != nil {
		result["cached"] = true
		result["result"] = job.Result
	}
	resp.Success(c, result)
}

func (h *JobHandler) GetStatus(c *gin.Context) {
	uuid := c.Param("id")
	job, err := h.svc.GetByUUID(uuid)
	if err != nil {
		resp.NotFound(c, errors.ErrJobNotFound)
		return
	}
	resp.Success(c, job)
}

func (h *JobHandler) Recommend(c *gin.Context) {
	scene := c.Query("scene")
	if scene == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "scene is required"))
		return
	}
	items, err := h.svc.RecommendByScene(scene, 20)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, items)
}

func (h *JobHandler) Retry(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	job, err := h.svc.Retry(id)
	if err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"job_id": job.UUID, "status": job.Status})
}

func (h *JobHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	jobs, total, err := h.svc.List(page, pageSize, status)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": jobs, "total": total})
}
