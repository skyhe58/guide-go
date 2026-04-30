// Prometheus 指标暴露 — Counter/Gauge/Histogram + HTTP 中间件 + /metrics 端点
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Prometheus client_golang 的完整用法：
// - 自定义 Counter（请求计数）
// - 自定义 Gauge（活跃连接数）
// - 自定义 Histogram（请求延迟分布）
// - HTTP Handler 指标采集中间件
// - /metrics 端点暴露（promhttp.Handler）
// - 使用 httptest 演示指标采集和查询
//
// 运行方式：go run ./prometheus/
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ============================================================
// 自定义 Prometheus 指标定义
// ============================================================

var (
	// Counter：HTTP 请求总数（按 method、path、status 分标签）
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "myapp",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "HTTP 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// Histogram：HTTP 请求延迟分布（秒）
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "myapp",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP 请求延迟分布（秒）",
			// 自定义桶：10ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"method", "path"},
	)

	// Gauge：当前活跃连接数
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "myapp",
			Subsystem: "http",
			Name:      "active_connections",
			Help:      "当前活跃 HTTP 连接数",
		},
	)

	// Gauge：当前 goroutine 数量（业务自定义）
	businessGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "business_goroutines",
			Help:      "业务 goroutine 数量",
		},
	)

	// Counter：业务错误计数
	businessErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "myapp",
			Name:      "business_errors_total",
			Help:      "业务错误总数",
		},
		[]string{"module", "error_type"},
	)
)

// registerMetrics 注册所有自定义指标
func registerMetrics() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(businessGoroutines)
	prometheus.MustRegister(businessErrors)
}

// ============================================================
// HTTP 指标采集中间件
// ============================================================

// metricsMiddleware 自动采集 HTTP 请求指标
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 活跃连接数 +1
		activeConnections.Inc()
		defer activeConnections.Dec()

		start := time.Now()

		// 包装 ResponseWriter 以获取状态码
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		// 记录请求计数和延迟
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ============================================================
// 业务 Handler
// ============================================================

// handleGetUsers 获取用户列表
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	// 模拟处理延迟（10-200ms）
	delay := time.Duration(10+rand.Intn(190)) * time.Millisecond
	time.Sleep(delay)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"users":["alice","bob"],"count":2}`)
}

// handleCreateOrder 创建订单
func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	// 模拟处理延迟（50-500ms）
	delay := time.Duration(50+rand.Intn(450)) * time.Millisecond
	time.Sleep(delay)

	// 模拟 30% 概率失败
	if rand.Float64() < 0.3 {
		businessErrors.WithLabelValues("order", "stock_insufficient").Inc()
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"库存不足"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"order_id":"ORD-1001","status":"created"}`)
}

// handleNotFound 404 处理
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error":"not found"}`)
}

// ============================================================
// 路由配置
// ============================================================

func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// 业务路由
	mux.HandleFunc("GET /api/users", handleGetUsers)
	mux.HandleFunc("POST /api/orders", handleCreateOrder)

	// Prometheus /metrics 端点
	mux.Handle("GET /metrics", promhttp.Handler())

	// 默认 404
	mux.HandleFunc("/", handleNotFound)

	// 包装中间件
	return metricsMiddleware(mux)
}

// ============================================================
// 主函数：使用 httptest 演示指标采集
// ============================================================

func main() {
	fmt.Println("========== Prometheus 指标暴露演示 ==========")
	fmt.Println()

	// 注册指标
	registerMetrics()

	// 设置初始 Gauge 值
	businessGoroutines.Set(5)

	router := setupRouter()

	// 模拟多次请求，产生指标数据
	fmt.Println("--- 模拟 HTTP 请求，产生指标数据 ---")

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/users", nil)
		router.ServeHTTP(w, req)
		fmt.Printf("  GET /api/users → %d\n", w.Code)
	}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/orders", nil)
		router.ServeHTTP(w, req)
		fmt.Printf("  POST /api/orders → %d\n", w.Code)
	}

	// 模拟 goroutine 数量变化
	businessGoroutines.Set(12)

	fmt.Println()
	fmt.Println("--- 查询 /metrics 端点 ---")

	// 查询 /metrics 端点
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	// 过滤输出自定义指标（myapp_ 前缀）
	fmt.Println("自定义指标（myapp_ 前缀）：")
	fmt.Println()
	lines := strings.Split(w.Body.String(), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "myapp_") {
			fmt.Println("  " + line)
		}
	}

	fmt.Println()
	fmt.Println("========== 演示完成 ==========")
	fmt.Println("生产环境中，Prometheus Server 会定期 Pull /metrics 端点采集指标")
	fmt.Println("然后在 Grafana 中通过 PromQL 查询和可视化")
	fmt.Println()
	fmt.Println("常用 PromQL 示例：")
	fmt.Println("  请求速率: rate(myapp_http_requests_total[5m])")
	fmt.Println("  P99 延迟: histogram_quantile(0.99, rate(myapp_http_request_duration_seconds_bucket[5m]))")
	fmt.Println("  错误率:   rate(myapp_http_requests_total{status=~\"5..\"}[5m]) / rate(myapp_http_requests_total[5m])")
}
