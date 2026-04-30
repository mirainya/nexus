package handler

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirainya/nexus/internal/api/middleware"
	"github.com/mirainya/nexus/internal/api/resp"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db       *gorm.DB
	attempts sync.Map
}

type loginAttempt struct {
	count    int
	lastTry  time.Time
}

const (
	maxLoginAttempts = 5
	loginLockout     = 15 * time.Minute
)

func NewAuthHandler(db *gorm.DB) *AuthHandler { return &AuthHandler{db: db} }

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary 管理员登录
// @Description 使用用户名密码登录，返回 JWT Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} resp.Response
// @Failure 400 {object} resp.Response
// @Failure 401 {object} resp.Response
// @Failure 429 {object} resp.Response
// @Router /admin/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, errors.WithMessage(errors.ErrInvalidParams, err.Error()))
		return
	}

	ip := c.ClientIP()
	key := ip + ":" + req.Username
	if val, ok := h.attempts.Load(key); ok {
		a := val.(*loginAttempt)
		if a.count >= maxLoginAttempts && time.Since(a.lastTry) < loginLockout {
			resp.Error(c, 429, errors.New(42900, "too many attempts, try again later"))
			return
		}
		if time.Since(a.lastTry) >= loginLockout {
			h.attempts.Delete(key)
		}
	}

	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		h.recordAttempt(key)
		resp.Unauthorized(c, errors.WithMessage(errors.ErrUnauthorized, "invalid credentials"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.recordAttempt(key)
		resp.Unauthorized(c, errors.WithMessage(errors.ErrUnauthorized, "invalid credentials"))
		return
	}

	h.attempts.Delete(key)

	token, err := middleware.GenerateToken(user.ID, user.Username, user.TenantID)
	if err != nil {
		resp.InternalError(c, errors.WithMessage(errors.ErrInternal, "failed to generate token"))
		return
	}

	result := gin.H{"token": token, "username": user.Username}
	if user.TenantID != nil {
		result["tenant_id"] = *user.TenantID
	}
	resp.Success(c, result)
}

func (h *AuthHandler) recordAttempt(key string) {
	val, _ := h.attempts.LoadOrStore(key, &loginAttempt{})
	a := val.(*loginAttempt)
	a.count++
	a.lastTry = time.Now()
}
