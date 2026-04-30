package handler

import (
	"strconv"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/service"
	"guide-go/goblog/pkg/pagination"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签模块 Handler
type TagHandler struct {
	tagSvc service.TagService
}

// NewTagHandler 创建标签 Handler 实例
func NewTagHandler(tagSvc service.TagService) *TagHandler {
	return &TagHandler{tagSvc: tagSvc}
}

// createTagRequest 创建标签请求参数
type createTagRequest struct {
	Name string `json:"name" binding:"required,max=50"`
	Slug string `json:"slug" binding:"required,max=50"`
}

// Create 创建标签
// POST /api/v1/tags
func (h *TagHandler) Create(c *gin.Context) {
	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	tag, appErr := h.tagSvc.Create(req.Name, req.Slug)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, tag)
}

// List 获取标签列表
// GET /api/v1/tags
func (h *TagHandler) List(c *gin.Context) {
	tags, appErr := h.tagSvc.List()
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, tags)
}

// GetArticles 获取标签下的文章列表
// GET /api/v1/tags/:id/articles
func (h *TagHandler) GetArticles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	params := pagination.GetParams(c)

	articles, total, appErr := h.tagSvc.GetArticlesByTagID(uint(id), params.Offset(), params.PageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, pagination.NewPagedResult(articles, total, params))
}
