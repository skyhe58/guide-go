---
title: "日志分析"
module: "nginx"
difficulty: "intermediate"
interviewFrequency: "low"
tags:
  - Nginx
  - 日志
  - access_log
  - error_log
  - 日志分析
relatedEntries:
  - "/5-devops/5.2-linux/04-log-analysis"
  - "/5-devops/5.3-nginx/02-reverse-proxy"
prerequisites:
  - "/5-devops/5.3-nginx/02-reverse-proxy"
estimatedTime: "25min"
---

# 日志分析

## 概念说明

Nginx 日志是排查线上问题的重要数据源。Nginx 提供两种日志：`access_log`（访问日志，记录每个请求）和 `error_log`（错误日志，记录 Nginx 自身的错误）。通过自定义日志格式和分析工具，可以快速定位慢请求、错误请求、异常流量等问题。

## 核心原理

### 日志类型

| 日志类型 | 文件 | 内容 | 用途 |
|----------|------|------|------|
| `access_log` | `/var/log/nginx/access.log` | 每个 HTTP 请求的详细信息 | 流量分析、慢请求定位 |
| `error_log` | `/var/log/nginx/error.log` | Nginx 错误、上游连接失败等 | 故障排查 |

## 标准配置方案

### 自定义日志格式

```nginx
http {
    # JSON 格式日志（便于 ELK 等日志系统解析）
    log_format json_log escape=json
        '{'
            '"time":"$time_iso8601",'
            '"remote_addr":"$remote_addr",'
            '"request_method":"$request_method",'
            '"request_uri":"$request_uri",'
            '"status":$status,'
            '"body_bytes_sent":$body_bytes_sent,'
            '"request_time":$request_time,'
            '"upstream_response_time":"$upstream_response_time",'
            '"http_referer":"$http_referer",'
            '"http_user_agent":"$http_user_agent",'
            '"http_x_forwarded_for":"$http_x_forwarded_for"'
        '}';

    # 使用 JSON 格式
    access_log /var/log/nginx/access.log json_log;

    # 错误日志级别：debug | info | notice | warn | error | crit
    error_log /var/log/nginx/error.log warn;
}
```

### 常用日志变量

| 变量 | 含义 | 示例 |
|------|------|------|
| `$time_iso8601` | ISO 8601 时间 | `2024-01-01T10:00:00+08:00` |
| `$remote_addr` | 客户端 IP | `192.168.1.100` |
| `$request_method` | HTTP 方法 | `GET` |
| `$request_uri` | 请求 URI | `/api/users?page=1` |
| `$status` | HTTP 状态码 | `200` |
| `$request_time` | 请求处理总时间（秒） | `0.052` |
| `$upstream_response_time` | 后端响应时间（秒） | `0.050` |
| `$body_bytes_sent` | 响应体大小（字节） | `1234` |

### 日志分析命令

```bash
# 统计 HTTP 状态码分布
cat access.log | jq -r '.status' | sort | uniq -c | sort -rn

# 找出最慢的 10 个请求
cat access.log | jq -r '[.request_time, .request_method, .request_uri] | @tsv' | sort -rn | head -10

# 统计每分钟的请求量（QPS 趋势）
cat access.log | jq -r '.time[:16]' | sort | uniq -c

# 统计 5xx 错误的请求路径
cat access.log | jq -r 'select(.status >= 500) | .request_uri' | sort | uniq -c | sort -rn

# 统计访问量最大的 IP（排查恶意请求）
cat access.log | jq -r '.remote_addr' | sort | uniq -c | sort -rn | head -20

# 统计后端响应时间 > 1 秒的慢请求
cat access.log | jq 'select(.request_time > 1) | {time, request_uri, request_time}'
```

### 按条件分离日志

```nginx
# 健康检查请求不记录日志
location /health {
    access_log off;
    proxy_pass http://127.0.0.1:8080/health;
}

# 静态文件使用单独的日志文件
location /static/ {
    access_log /var/log/nginx/static.log;
    alias /var/www/static/;
}
```

## 常见面试题

### Q1: `$request_time` 和 `$upstream_response_time` 的区别？

**难度**：⭐⭐ | **频率**：🔥

**标准答案**：

- `$request_time`：从 Nginx 接收到客户端第一个字节到发送完最后一个字节的总时间，包含 Nginx 处理时间 + 后端响应时间 + 网络传输时间
- `$upstream_response_time`：从 Nginx 向后端发送请求到接收完后端响应的时间，仅包含后端处理时间

如果 `$request_time` 远大于 `$upstream_response_time`，说明瓶颈在网络传输或 Nginx 自身处理。

## 常见陷阱

1. **日志写满磁盘**：高流量服务的 access_log 增长很快，务必配置 logrotate
2. **日志缓冲**：高并发时频繁写日志影响性能，可使用 `access_log ... buffer=32k flush=5s` 缓冲写入
3. **敏感信息泄露**：日志中可能包含 Token、密码等敏感信息，注意脱敏

## 参考资料

- [Nginx log_format 文档](https://nginx.org/en/docs/http/ngx_http_log_module.html)
- [Nginx 变量列表](https://nginx.org/en/docs/varindex.html)
