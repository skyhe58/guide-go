// Package pagination 提供分页参数解析与分页响应构建功能
// 支持从 HTTP 查询参数中解析 page 和 page_size，并构建统一的分页响应
package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 分页参数默认值与限制
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Params 分页请求参数
type Params struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Offset 计算数据库查询的偏移量
func (p *Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PagedResult 分页响应结构体
type PagedResult struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// GetParams 从 Gin 上下文中解析分页参数
// 自动处理默认值和边界校验
func GetParams(c *gin.Context) Params {
	page := parseIntParam(c.Query("page"), DefaultPage)
	pageSize := parseIntParam(c.Query("page_size"), DefaultPageSize)

	// 边界校验：page 最小为 1
	if page < 1 {
		page = DefaultPage
	}

	// 边界校验：page_size 范围 [1, MaxPageSize]
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// NewPagedResult 构建分页响应
func NewPagedResult(items interface{}, total int64, params Params) *PagedResult {
	return &PagedResult{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
}

// parseIntParam 解析整数查询参数，解析失败时返回默认值
func parseIntParam(value string, defaultVal int) int {
	if value == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return n
}
