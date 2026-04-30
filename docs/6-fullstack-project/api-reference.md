---
title: "API 接口参考文档"
module: "fullstack-project"
difficulty: "intermediate"
tags:
  - API
  - REST
  - Swagger
codeExample: "06-fullstack-project/goblog/"
estimatedTime: "30min"
---

# API 接口参考文档

## 概念说明

GoBlog 提供 RESTful API 接口，所有接口遵循统一的请求/响应格式。本文档列出所有 API 端点及其参数说明。

## 统一响应格式

### 成功响应

```json
{
    "code": 0,
    "message": "success",
    "data": { ... }
}
```

### 错误响应

```json
{
    "code": 10001,
    "message": "参数验证失败",
    "data": null
}
```

### 分页响应

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [ ... ],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}
```

## 认证方式

需要认证的接口在 Header 中携带 JWT Token：

```
Authorization: Bearer <access_token>
```

## API 端点

### 用户模块

#### POST /api/v1/auth/register

用户注册。

**请求体：**

```json
{
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "password": "123456"
}
```

**响应：** 201 Created

#### POST /api/v1/auth/login

用户登录，返回 JWT 双令牌。

**请求体：**

```json
{
    "username": "zhangsan",
    "password": "123456"
}
```

**响应：**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 900
    }
}
```

#### POST /api/v1/auth/refresh

刷新 Access Token。

**请求体：**

```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### POST /api/v1/auth/logout

用户登出（需认证），将 Token 加入黑名单。

#### GET /api/v1/users/me

获取当前用户信息（需认证）。

#### PUT /api/v1/users/me

更新用户资料（需认证）。

**请求体：**

```json
{
    "avatar": "https://example.com/avatar.jpg",
    "bio": "Go 语言爱好者"
}
```

### 文章模块

#### POST /api/v1/articles

创建文章（需 author/admin 角色）。

**请求体：**

```json
{
    "title": "Go Slice 扩容机制",
    "content": "# Slice 扩容\n\n...",
    "slug": "go-slice-grow",
    "status": "published",
    "tag_ids": [1, 2]
}
```

#### PUT /api/v1/articles/:id

更新文章（需文章作者或 admin）。

#### DELETE /api/v1/articles/:id

删除文章（软删除，需文章作者或 admin）。

#### GET /api/v1/articles/:id

获取文章详情（公开，含缓存）。

#### GET /api/v1/articles

文章列表（公开，支持分页）。

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 20 | 每页数量（最大 100） |
| tag_id | int | — | 按标签筛选 |
| status | string | — | 按状态筛选 |

#### GET /api/v1/articles/search

文章搜索（公开）。

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| q | string | 搜索关键词（标题和内容模糊匹配） |

### 标签模块

#### POST /api/v1/tags

创建标签（需 author/admin 角色）。

**请求体：**

```json
{
    "name": "Go 基础",
    "slug": "go-basics"
}
```

#### GET /api/v1/tags

获取标签列表（公开）。

#### GET /api/v1/tags/:id/articles

获取标签下的文章列表（公开，支持分页）。

### 评论模块

#### POST /api/v1/articles/:id/comments

发表评论（需登录）。

**请求体：**

```json
{
    "content": "写得很好，学到了！"
}
```

#### GET /api/v1/articles/:id/comments

获取文章评论列表（公开，支持分页）。

#### DELETE /api/v1/comments/:id

删除评论（需评论作者或 admin）。

### 管理模块

#### GET /api/v1/admin/users

用户列表（仅 admin）。

#### PUT /api/v1/admin/users/:id/role

修改用户角色（仅 admin）。

**请求体：**

```json
{
    "role": "author"
}
```

#### PUT /api/v1/admin/articles/:id/status

文章审核（仅 admin）。

**请求体：**

```json
{
    "status": "published"
}
```

#### GET /api/v1/admin/stats

系统统计（仅 admin）。

### 监控端点

#### GET /metrics

Prometheus 指标端点。

#### GET /healthz

健康检查端点。

## 错误码表

| 错误码 | 说明 | HTTP 状态码 |
|--------|------|------------|
| 0 | 成功 | 200 |
| 10001 | 参数验证失败 | 400 |
| 10002 | 未授权 | 401 |
| 10003 | 禁止访问 | 403 |
| 10004 | 资源不存在 | 404 |
| 10005 | 请求过于频繁 | 429 |
| 10101 | 用户不存在 | 404 |
| 10102 | 密码错误 | 401 |
| 10103 | 用户名已存在 | 409 |
| 10104 | 邮箱已注册 | 409 |
| 10105 | Token 无效 | 401 |
| 10106 | Token 已过期 | 401 |
| 10107 | Refresh Token 无效 | 401 |
| 10201 | 文章不存在 | 404 |
| 10202 | 无权编辑此文章 | 403 |
| 10301 | 标签不存在 | 404 |
| 10302 | 标签名已存在 | 409 |
| 10401 | 评论不存在 | 404 |
| 10402 | 无权删除此评论 | 403 |

## 参考资料

- [RESTful API 设计指南](https://restfulapi.net/)
- [HTTP 状态码](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Status)
