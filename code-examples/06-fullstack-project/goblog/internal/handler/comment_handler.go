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

// CommentHandler 评论模块 Handler
type CommentHandler struct {
	commentSvc service.CommentService
}

// NewCommentHandler 创建评论 Handler 实例
func NewCommentHandler(commentSvc service.CommentService) *CommentHandler {
	return &CommentHandler{commentSvc: commentSvc}
}

// createCommentRequest 创建评论请求参数
type createCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// Create 发表评论
// POST /api/v1/articles/:id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	userID := middleware.GetUserID(c)

	comment, appErr := h.commentSvc.Create(uint(articleID), userID, req.Content)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, comment)
}

// List 获取文章评论列表
// GET /api/v1/articles/:id/comments
func (h *CommentHandler) List(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	params := pagination.GetParams(c)

	comments, total, appErr := h.commentSvc.ListByArticleID(uint(articleID), params.Offset(), params.PageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, pagination.NewPagedResult(comments, total, params))
}

// Delete 删除评论
// DELETE /api/v1/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetRole(c)

	appErr := h.commentSvc.Delete(uint(id), operatorID, operatorRole)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, nil)
}
