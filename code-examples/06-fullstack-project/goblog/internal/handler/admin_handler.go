package handler

import (
	"strconv"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/middleware"
	"guide-go/goblog/internal/service"
	"guide-go/goblog/pkg/pagination"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理后台 Handler
type AdminHandler struct {
	adminSvc service.AdminService
}

// NewAdminHandler 创建管理 Handler 实例
func NewAdminHandler(adminSvc service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

// ListUsers 获取用户列表
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	params := pagination.GetParams(c)

	users, total, appErr := h.adminSvc.ListUsers(params.Offset(), params.PageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, pagination.NewPagedResult(users, total, params))
}

// updateRoleRequest 修改角色请求参数
type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin author reader"`
}

// UpdateRole 修改用户角色
// PUT /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	operatorID := middleware.GetUserID(c)

	appErr := h.adminSvc.UpdateRole(uint(userID), operatorID, req.Role)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, nil)
}

// updateArticleStatusRequest 更新文章状态请求参数
type updateArticleStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft published archived"`
}

// UpdateArticleStatus 更新文章状态（审核上架/下架）
// PUT /api/v1/admin/articles/:id/status
func (h *AdminHandler) UpdateArticleStatus(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req updateArticleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	appErr := h.adminSvc.UpdateArticleStatus(uint(articleID), req.Status)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, nil)
}

// GetStats 获取系统统计数据
// GET /api/v1/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, appErr := h.adminSvc.GetStats()
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, stats)
}
