// Package errcode 定义所有业务错误码常量
// 错误码分段规则：
//   - 0:           成功
//   - 10000-10099: 通用错误
//   - 10100-10199: 用户模块错误
//   - 10200-10299: 文章模块错误
//   - 10300-10399: 标签模块错误
//   - 10400-10499: 评论模块错误
//   - 10500-10599: 管理模块错误
package errcode

import "net/http"

// ==================== 成功 ====================

// ErrSuccess 操作成功
var ErrSuccess = &AppError{Code: 0, Message: "success", HTTPStatus: http.StatusOK}

// ==================== 通用错误 10000-10099 ====================

// ErrInvalidParams 参数验证失败
var ErrInvalidParams = &AppError{Code: 10001, Message: "参数验证失败", HTTPStatus: http.StatusBadRequest}

// ErrUnauthorized 未授权（未登录或 Token 无效）
var ErrUnauthorized = &AppError{Code: 10002, Message: "未授权，请先登录", HTTPStatus: http.StatusUnauthorized}

// ErrForbidden 禁止访问（权限不足）
var ErrForbidden = &AppError{Code: 10003, Message: "权限不足，禁止访问", HTTPStatus: http.StatusForbidden}

// ErrNotFound 资源不存在
var ErrNotFound = &AppError{Code: 10004, Message: "资源不存在", HTTPStatus: http.StatusNotFound}

// ErrTooManyRequests 请求过于频繁（触发限流）
var ErrTooManyRequests = &AppError{Code: 10005, Message: "请求过于频繁，请稍后再试", HTTPStatus: http.StatusTooManyRequests}

// ErrInternal 服务器内部错误
var ErrInternal = &AppError{Code: 10006, Message: "服务器内部错误", HTTPStatus: http.StatusInternalServerError}

// ==================== 用户模块错误 10100-10199 ====================

// ErrUserNotFound 用户不存在
var ErrUserNotFound = &AppError{Code: 10101, Message: "用户不存在", HTTPStatus: http.StatusNotFound}

// ErrPasswordWrong 密码错误
var ErrPasswordWrong = &AppError{Code: 10102, Message: "密码错误", HTTPStatus: http.StatusUnauthorized}

// ErrUsernameExists 用户名已存在
var ErrUsernameExists = &AppError{Code: 10103, Message: "用户名已存在", HTTPStatus: http.StatusConflict}

// ErrEmailExists 邮箱已注册
var ErrEmailExists = &AppError{Code: 10104, Message: "邮箱已注册", HTTPStatus: http.StatusConflict}

// ErrTokenInvalid Token 无效
var ErrTokenInvalid = &AppError{Code: 10105, Message: "Token 无效", HTTPStatus: http.StatusUnauthorized}

// ErrTokenExpired Token 已过期
var ErrTokenExpired = &AppError{Code: 10106, Message: "Token 已过期", HTTPStatus: http.StatusUnauthorized}

// ErrRefreshTokenInvalid Refresh Token 无效
var ErrRefreshTokenInvalid = &AppError{Code: 10107, Message: "Refresh Token 无效", HTTPStatus: http.StatusUnauthorized}

// ==================== 文章模块错误 10200-10299 ====================

// ErrArticleNotFound 文章不存在
var ErrArticleNotFound = &AppError{Code: 10201, Message: "文章不存在", HTTPStatus: http.StatusNotFound}

// ErrArticleNoPermission 无权编辑此文章
var ErrArticleNoPermission = &AppError{Code: 10202, Message: "无权编辑此文章", HTTPStatus: http.StatusForbidden}

// ErrArticleTitleDuplicate 文章标题重复
var ErrArticleTitleDuplicate = &AppError{Code: 10203, Message: "文章标题重复", HTTPStatus: http.StatusConflict}

// ==================== 标签模块错误 10300-10399 ====================

// ErrTagNotFound 标签不存在
var ErrTagNotFound = &AppError{Code: 10301, Message: "标签不存在", HTTPStatus: http.StatusNotFound}

// ErrTagNameExists 标签名已存在
var ErrTagNameExists = &AppError{Code: 10302, Message: "标签名已存在", HTTPStatus: http.StatusConflict}

// ==================== 评论模块错误 10400-10499 ====================

// ErrCommentNotFound 评论不存在
var ErrCommentNotFound = &AppError{Code: 10401, Message: "评论不存在", HTTPStatus: http.StatusNotFound}

// ErrCommentNoPermission 无权删除此评论
var ErrCommentNoPermission = &AppError{Code: 10402, Message: "无权删除此评论", HTTPStatus: http.StatusForbidden}

// ==================== 管理模块错误 10500-10599 ====================

// ErrCannotChangeOwnRole 不能修改自己的角色
var ErrCannotChangeOwnRole = &AppError{Code: 10501, Message: "不能修改自己的角色", HTTPStatus: http.StatusBadRequest}

// ErrInvalidRole 无效的角色值
var ErrInvalidRole = &AppError{Code: 10502, Message: "无效的角色值", HTTPStatus: http.StatusBadRequest}
