package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/errors"
)

type CredentialHandler struct {
	svc *service.CredentialService
}

func NewCredentialHandler(svc *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{svc: svc}
}

// Create godoc
// @Summary 创建凭证
// @Description 创建 LLM 凭证，绑定到 API Key，用于外部用户自带模型
// @Tags 凭证
// @Accept json
// @Produce json
// @Param request body service.CredentialCreateRequest true "创建请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/credentials [post]
func (h *CredentialHandler) Create(c *gin.Context) {
	var req service.CredentialCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	cred, err := h.svc.Create(req, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, cred)
}

// List godoc
// @Summary 获取凭证列表
// @Description 获取凭证列表，可按 API Key 过滤
// @Tags 凭证
// @Produce json
// @Param api_key_id query int false "API Key ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/credentials [get]
func (h *CredentialHandler) List(c *gin.Context) {
	var apiKeyID uint
	if v := c.Query("api_key_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			apiKeyID = uint(n)
		}
	}
	list, err := h.svc.List(apiKeyID, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, list)
}

// Update godoc
// @Summary 更新凭证
// @Description 更新凭证信息
// @Tags 凭证
// @Accept json
// @Produce json
// @Param id path int true "凭证 ID"
// @Param request body service.CredentialUpdateRequest true "更新请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Security BearerAuth
// @Router /admin/credentials/{id} [put]
func (h *CredentialHandler) Update(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	var req service.CredentialUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}
	cred, err := h.svc.Update(id, req, getTenantID(c))
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, cred)
}

// Delete godoc
// @Summary 删除凭证
// @Description 删除指定凭证
// @Tags 凭证
// @Produce json
// @Param id path int true "凭证 ID"
// @Success 200 {object} resp.Response
// @Security BearerAuth
// @Router /admin/credentials/{id} [delete]
func (h *CredentialHandler) Delete(c *gin.Context) {
	id, err := resp.ParseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(id, getTenantID(c)); err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, err.Error()))
		return
	}
	resp.Success(c, nil)
}
