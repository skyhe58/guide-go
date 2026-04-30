// Package response 提供 GoBlog 统一的 JSON 响应格式
// 所有 API 接口返回统一的 {code, message, data} 结构
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 返回成功响应，HTTP 状态码 200
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应，根据 AppError 中的 HTTP 状态码返回
// err 需实现 AppErrorInterface 接口
func Error(c *gin.Context, err AppErrorInterface) {
	c.JSON(err.GetHTTPStatus(), Response{
		Code:    err.GetCode(),
		Message: err.GetMessage(),
		Data:    nil,
	})
}

// ErrorWithData 返回带附加数据的错误响应
// 适用于参数验证失败等需要返回详细错误信息的场景
func ErrorWithData(c *gin.Context, err AppErrorInterface, data interface{}) {
	c.JSON(err.GetHTTPStatus(), Response{
		Code:    err.GetCode(),
		Message: err.GetMessage(),
		Data:    data,
	})
}

// AppErrorInterface 定义错误响应接口
// 解耦 response 包与 errcode 包的直接依赖
type AppErrorInterface interface {
	GetCode() int
	GetMessage() string
	GetHTTPStatus() int
}
