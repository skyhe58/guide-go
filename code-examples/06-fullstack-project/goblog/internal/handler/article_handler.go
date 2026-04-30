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

// ArticleHandler 文章模块 Handler
type ArticleHandler struct {
	articleSvc service.ArticleService
}

// NewArticleHandler 创建文章 Handler 实例
func NewArticleHandler(articleSvc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleSvc: articleSvc}
}

// createArticleRequest 创建文章请求参数
type createArticleRequest struct {
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
	Slug    string `json:"slug" binding:"required,max=200"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published"`
	TagIDs  []uint `json:"tag_ids"`
}

// Create 创建文章
// POST /api/v1/articles
func (h *ArticleHandler) Create(c *gin.Context) {
	var req createArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	authorID := middleware.GetUserID(c)
	status := req.Status
	if status == "" {
		status = "draft"
	}

	article, appErr := h.articleSvc.Create(authorID, req.Title, req.Content, req.Slug, status, req.TagIDs)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, article)
}

// updateArticleRequest 更新文章请求参数
type updateArticleRequest struct {
	Title   string `json:"title" binding:"omitempty,max=200"`
	Content string `json:"content"`
	Slug    string `json:"slug" binding:"omitempty,max=200"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published archived"`
	TagIDs  []uint `json:"tag_ids"`
}

// Update 更新文章
// PUT /api/v1/articles/:id
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req updateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParams.WithDetails(err.Error()))
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetRole(c)

	article, appErr := h.articleSvc.Update(uint(id), operatorID, operatorRole, req.Title, req.Content, req.Slug, req.Status, req.TagIDs)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, article)
}

// Delete 删除文章（软删除）
// DELETE /api/v1/articles/:id
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetRole(c)

	appErr := h.articleSvc.Delete(uint(id), operatorID, operatorRole)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, nil)
}

// GetByID 获取文章详情
// GET /api/v1/articles/:id
func (h *ArticleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	article, appErr := h.articleSvc.GetByID(uint(id))
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, article)
}

// List 文章列表（分页）
// GET /api/v1/articles?page=1&page_size=20&status=published&tag_id=1
func (h *ArticleHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	status := c.Query("status")

	var tagID uint
	if tagIDStr := c.Query("tag_id"); tagIDStr != "" {
		if id, err := strconv.ParseUint(tagIDStr, 10, 64); err == nil {
			tagID = uint(id)
		}
	}

	articles, total, appErr := h.articleSvc.List(params.Offset(), params.PageSize, status, tagID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, pagination.NewPagedResult(articles, total, params))
}

// Search 文章搜索
// GET /api/v1/articles/search?q=keyword&page=1&page_size=20
func (h *ArticleHandler) Search(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		response.Error(c, errcode.ErrInvalidParams.WithDetails("搜索关键词不能为空"))
		return
	}

	params := pagination.GetParams(c)

	articles, total, appErr := h.articleSvc.Search(keyword, params.Offset(), params.PageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.Success(c, pagination.NewPagedResult(articles, total, params))
}
