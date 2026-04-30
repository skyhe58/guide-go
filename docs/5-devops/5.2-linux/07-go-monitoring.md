---
title: "Go 服务监控"
module: "linux"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - Prometheus
  - Grafana
  - 监控
  - metrics
  - Go 服务
relatedEntries:
  - "/5-devops/5.2-linux/06-go-troubleshooting"
  - "/2-web-data/2.7-observability/"
prerequisites:
  - "/1-go-core/1.1-go-basics/"
  - "/5-devops/5.2-linux/05-go-deploy"
estimatedTime: "40min"
---

# Go 服务监控

## 概念说明

监控是保障线上服务稳定性的基石。Go 服务监控的核心方案是 **Prometheus + Grafana**：Prometheus 负责采集和存储指标数据，Grafana 负责可视化展示和告警。Go 标准库和 Prometheus 客户端库提供了开箱即用的指标暴露能力。

## 核心原理

### 监控架构

```mermaid
graph LR
    subgraph "Go 服务"
        APP[业务逻辑] --> METRICS[Prometheus Client<br/>指标采集]
        METRICS --> ENDPOINT[/metrics 端点]
    end
    
    subgraph "监控平台"
        PROM[Prometheus<br/>指标存储] -->|拉取| ENDPOINT
        PROM --> GRAFANA[Grafana<br/>可视化看板]
        PROM --> ALERT[AlertManager<br/>告警通知]
    end
    
    ALERT --> SLACK[Slack/钉钉]
    ALERT --> EMAIL[邮件]
```

### Prometheus 指标类型

| 类型 | 用途 | Go 服务示例 |
|------|------|-------------|
| **Counter** | 只增不减的计数器 | HTTP 请求总数、错误总数 |
| **Gauge** | 可增可减的仪表盘 | 当前 goroutine 数、连接池活跃连接数 |
| **Histogram** | 分布统计（分桶） | 请求延迟分布、响应大小分布 |
| **Summary** | 分位数统计 | P50/P95/P99 延迟 |

## 标准库方案

### Go 服务暴露 Prometheus 指标

```go
package main

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义指标
var (
    // Counter：HTTP 请求总数
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "HTTP 请求总数",
        },
        []string{"method", "path", "status"},
    )

    // Histogram：请求延迟分布
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP 请求延迟（秒）",
            Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
        },
        []string{"method", "path"},
    )

    // Gauge：当前活跃连接数
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "当前活跃连接数",
        },
    )
)

func main() {
    // 暴露 /metrics 端点
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### Gin 中间件指标采集

```go
// Gin 中间件：自动采集每个请求的指标
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

### Go 运行时指标

Prometheus Go 客户端库默认暴露以下运行时指标：

| 指标 | 含义 | 告警建议 |
|------|------|----------|
| `go_goroutines` | 当前 goroutine 数量 | 持续增长告警 |
| `go_memstats_alloc_bytes` | 当前堆内存分配 | 超过阈值告警 |
| `go_memstats_sys_bytes` | 从 OS 获取的总内存 | 监控趋势 |
| `go_gc_duration_seconds` | GC 暂停时间 | P99 > 10ms 告警 |
| `go_threads` | OS 线程数 | 异常增长告警 |
| `process_cpu_seconds_total` | 进程 CPU 使用时间 | 计算 CPU 使用率 |
| `process_resident_memory_bytes` | 进程物理内存 | 超过阈值告警 |

### Prometheus 配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s      # 采集间隔
  evaluation_interval: 15s  # 规则评估间隔

scrape_configs:
  - job_name: 'go-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### Grafana 看板常用面板

```
# Go 服务 Grafana 看板推荐面板

1. 请求速率（QPS）
   PromQL: rate(http_requests_total[5m])

2. 请求延迟 P95
   PromQL: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

3. 错误率
   PromQL: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

4. Goroutine 数量
   PromQL: go_goroutines

5. 堆内存使用
   PromQL: go_memstats_alloc_bytes

6. GC 暂停时间
   PromQL: go_gc_duration_seconds{quantile="1"}

7. CPU 使用率
   PromQL: rate(process_cpu_seconds_total[5m])
```

## 常见面试题

### Q1: Go 服务如何暴露 Prometheus 指标？

**难度**：⭐⭐ | **频率**：🔥🔥

**标准答案**：

1. 引入 `prometheus/client_golang` 库
2. 定义业务指标（Counter/Gauge/Histogram）
3. 在 HTTP Handler 或 Gin 中间件中记录指标
4. 暴露 `/metrics` 端点：`http.Handle("/metrics", promhttp.Handler())`
5. Prometheus 定期拉取 `/metrics` 端点的数据

### Q2: Prometheus 的四种指标类型分别适用于什么场景？

**难度**：⭐⭐ | **频率**：🔥🔥

**标准答案**：

- **Counter**：只增不减，适合计数场景（请求总数、错误总数、处理的消息数）
- **Gauge**：可增可减，适合当前状态（goroutine 数、队列长度、温度）
- **Histogram**：分桶统计分布，适合延迟和大小（请求延迟、响应大小），可计算分位数
- **Summary**：客户端计算分位数，适合精确的 P50/P95/P99，但不支持聚合

## 常见陷阱

1. **Label 基数爆炸**：不要用高基数值（如用户 ID、请求 ID）作为 Label，会导致 Prometheus 内存暴涨
2. **Histogram Bucket 设置**：Bucket 边界要根据实际延迟分布设置，默认值可能不适合你的服务
3. **指标命名规范**：遵循 `<namespace>_<subsystem>_<name>_<unit>` 格式，如 `http_request_duration_seconds`
4. **采集间隔**：不要设置过短的采集间隔（如 1 秒），会增加 Prometheus 和服务的负担

## 参考资料

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Prometheus Go 客户端](https://github.com/prometheus/client_golang)
- [Grafana Go 服务看板模板](https://grafana.com/grafana/dashboards/)
