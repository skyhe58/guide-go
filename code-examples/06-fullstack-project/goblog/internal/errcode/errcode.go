// Package errcode 定义 GoBlog 的业务错误码体系
// 每个业务错误包含错误码、错误消息和对应的 HTTP 状态码
package errcode

import (
	"fmt"
	"net/http"
)

// AppError 业务错误类型，实现 error 接口
// Code 为业务错误码，Message 为用户可见的错误描述
// HTTPStatus 为对应的 HTTP 状态码，不会序列化到 JSON 响应中
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return fmt.Sprintf("错误码: %d, 消息: %s", e.Code, e.Message)
}

// GetCode 返回业务错误码
func (e *AppError) GetCode() int {
	return e.Code
}

// GetMessage 返回错误消息
func (e *AppError) GetMessage() string {
	return e.Message
}

// GetHTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) GetHTTPStatus() int {
	return e.HTTPStatus
}

// WithDetails 返回一个新的 AppError，附加详细错误信息
// 用于参数验证等需要返回具体错误详情的场景
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message + ": " + details,
		HTTPStatus: e.HTTPStatus,
	}
}

// NewAppError 创建自定义业务错误
func NewAppError(code int, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// httpStatusMap 定义错误码范围到 HTTP 状态码的默认映射
// 用于根据错误码自动推断 HTTP 状态码
var httpStatusMap = map[int]int{
	0: http.StatusOK, // 成功
}

// RegisterHTTPStatus 注册错误码到 HTTP 状态码的映射
func RegisterHTTPStatus(code int, httpStatus int) {
	httpStatusMap[code] = httpStatus
}

// LookupHTTPStatus 根据错误码查找对应的 HTTP 状态码
// 如果未注册，返回 500 Internal Server Error
func LookupHTTPStatus(code int) int {
	if status, ok := httpStatusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}
