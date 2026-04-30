package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type JobHandler struct {
	svc          *service.JobService
	recommendSvc *service.RecommendService
}

func NewJobHandler(svc *service.JobService, recommendSvc *service.RecommendService) *JobHandler {
	return &JobHandler{svc: svc, recommendSvc: recommendSvc}
}

// Submit godoc
// @Summary 提交处理任务
// @Description 提交文档到 Pipeline 进行异步处理，返回任务 ID
// @Tags 任务
// @Accept json
// @Produce json
// @Param request body service.JobSubmitRequest true "任务提交请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Failure 500 {object} resp.Response
// @Security ApiKeyAuth
// @Router /v1/jobs [post]
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
	req.TenantID = getTenantID(c)
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

// GetStatus godoc
// @Summary 查询任务状态
// @Description 根据任务 UUID 查询处理状态和结果
// @Tags 任务
// @Produce json
// @Param id path string true "任务 UUID"
// @Success 200 {object} resp.Response
// @Failure 404 {object} resp.Response
// @Security ApiKeyAuth
// @Router /v1/jobs/{id} [get]
func (h *JobHandler) GetStatus(c *gin.Context) {
	uuid := c.Param("id")
	job, err := h.svc.GetByUUID(uuid, getTenantID(c))
	if err != nil {
		resp.NotFound(c, errors.ErrJobNotFound)
		return
	}
	resp.Success(c, job)
}

// Recommend godoc
// @Summary 推荐任务
// @Description 根据场景推荐相关任务
// @Tags 任务
// @Produce json
// @Param scene query string true "场景名称"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/jobs/recommend [get]
func (h *JobHandler) Recommend(c *gin.Context) {
	scene := c.Query("scene")
	if scene == "" {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, "scene is required"))
		return
	}
	items, err := h.recommendSvc.ByScene(scene, 20)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, items)
}

// Retry godoc
// @Summary 重试任务
// @Description 重试失败的任务
// @Tags 任务
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/jobs/{id}/retry [post]
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

// List godoc
// @Summary 获取任务列表
// @Description 分页获取任务列表，支持按状态筛选
// @Tags 任务
// @Produce json
// @Param status query string false "状态筛选"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/jobs [get]
func (h *JobHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	jobs, total, err := h.svc.List(page, pageSize, status, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, gin.H{"list": jobs, "total": total})
}
