// Package handler 提供 GoBlog 的 HTTP 请求处理层
// 每个 Handler 负责参数绑定与验证、调用 Service、返回统一 JSON 响应
package handler

import (
	"time"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/middleware"
	"guide-go/goblog/internal/service"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户模块 Handler
type UserHandler struct {
	userSvc service.UserService
}

// NewUserHandler 创建用户 Handler 实例
func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// registerRequest 注册请求参数
type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	user, appErr := h.userSvc.Register(req.Username, req.Email, req.Password)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, user)
}

// loginRequest 登录请求参数
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	tokenPair, appErr := h.userSvc.Login(req.Username, req.Password)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, tokenPair)
}

// refreshTokenRequest 刷新 Token 请求参数
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新 Access Token
// POST /api/v1/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	tokenPair, appErr := h.userSvc.RefreshToken(req.RefreshToken)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, tokenPair)
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *UserHandler) Logout(c *gin.Context) {
	jti, exists := c.Get(middleware.ContextJTIKey)
	if !exists {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	// Token 黑名单有效期设为 Access Token 剩余有效期（简化为 15 分钟）
	appErr := h.userSvc.Logout(jti.(string), 15*time.Minute)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, nil)
}

// GetProfile 获取当前用户资料
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	user, appErr := h.userSvc.GetProfile(userID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, user)
}

// updateProfileRequest 更新资料请求参数
type updateProfileRequest struct {
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
}

// UpdateProfile 更新用户资料
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, errcode.ErrUnauthorized)
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	user, appErr := h.userSvc.UpdateProfile(userID, req.Avatar, req.Bio)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, user)
}
