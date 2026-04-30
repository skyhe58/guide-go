// slog 自定义 Handler 示例 — 带颜色的控制台 Handler + 上下文传播 + 日志分组
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 标准库 slog 的高级用法：
// - 实现自定义 slog.Handler（带颜色的控制台输出）
// - slog.With 上下文传播
// - slog.Group 日志分组
// - 日志级别控制
//
// 运行方式：go run ./slog-custom/
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
)

// ============================================================
// 自定义 Handler：带颜色的控制台输出
// 实现 slog.Handler 接口的四个方法
// ============================================================

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// ColorHandler 带颜色输出的控制台 Handler
type ColorHandler struct {
	opts      slog.HandlerOptions
	w         io.Writer
	mu        *sync.Mutex
	attrs     []slog.Attr  // 通过 WithAttrs 添加的属性
	groupName string       // 通过 WithGroup 添加的分组名
	groups    []string     // 嵌套分组链
}

// NewColorHandler 创建带颜色的控制台 Handler
func NewColorHandler(w io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColorHandler{
		opts: *opts,
		w:    w,
		mu:   &sync.Mutex{},
	}
}

// Enabled 判断是否需要处理该级别的日志
func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle 处理日志记录，输出带颜色的格式化文本
func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	// 时间戳
	timeStr := r.Time.Format("15:04:05.000")

	// 根据级别选择颜色
	levelColor := colorGreen
	levelText := r.Level.String()
	switch {
	case r.Level >= slog.LevelError:
		levelColor = colorRed
	case r.Level >= slog.LevelWarn:
		levelColor = colorYellow
	case r.Level >= slog.LevelInfo:
		levelColor = colorGreen
	default:
		levelColor = colorBlue
	}

	// 构建输出
	h.mu.Lock()
	defer h.mu.Unlock()

	// 基本格式：时间 [级别] 消息
	fmt.Fprintf(h.w, "%s%s%s %s%-5s%s %s",
		colorGray, timeStr, colorReset,
		levelColor, levelText, colorReset,
		r.Message,
	)

	// 输出 WithAttrs 添加的固定属性
	for _, attr := range h.attrs {
		h.writeAttr(attr)
	}

	// 输出本次日志的属性
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(a)
		return true
	})

	// 输出调用者信息（如果启用）
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		fmt.Fprintf(h.w, " %s(%s:%d)%s", colorGray, f.File, f.Line, colorReset)
	}

	fmt.Fprintln(h.w)
	return nil
}

// writeAttr 输出单个属性（支持分组）
func (h *ColorHandler) writeAttr(a slog.Attr) {
	// 处理分组属性
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		for _, ga := range attrs {
			prefix := ""
			if a.Key != "" {
				prefix = a.Key + "."
			}
			fmt.Fprintf(h.w, " %s%s%s%s=%v%s",
				colorCyan, prefix, ga.Key, colorReset,
				ga.Value, colorReset,
			)
		}
		return
	}

	// 普通属性
	key := a.Key
	if h.groupName != "" {
		key = h.groupName + "." + key
	}
	for i := len(h.groups) - 1; i >= 0; i-- {
		key = h.groups[i] + "." + key
	}
	fmt.Fprintf(h.w, " %s%s%s=%v", colorCyan, key, colorReset, a.Value)
}

// WithAttrs 返回带额外属性的新 Handler
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColorHandler{
		opts:      h.opts,
		w:         h.w,
		mu:        h.mu,
		attrs:     newAttrs,
		groupName: h.groupName,
		groups:    h.groups,
	}
}

// WithGroup 返回带分组名的新 Handler
func (h *ColorHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &ColorHandler{
		opts:   h.opts,
		w:      h.w,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// ============================================================
// 演示函数
// ============================================================

// demoBasicUsage 演示基本日志级别
func demoBasicUsage(logger *slog.Logger) {
	fmt.Println("\n--- 1. 基本日志级别 ---")
	logger.Debug("这条 DEBUG 日志不会显示（级别低于 INFO）")
	logger.Info("服务启动成功", "port", 8080)
	logger.Warn("配置项缺失，使用默认值", "key", "cache_ttl", "default", "5m")
	logger.Error("数据库连接失败", "host", "localhost:5432", "error", "connection refused")
}

// demoContextPropagation 演示 slog.With 上下文传播
func demoContextPropagation(logger *slog.Logger) {
	fmt.Println("\n--- 2. 上下文传播（slog.With）---")

	// 创建带固定字段的子 Logger，模拟请求处理链
	requestLogger := logger.With(
		"request_id", "req-abc-123",
		"user_id", 42,
	)

	// 后续所有日志自动携带 request_id 和 user_id
	requestLogger.Info("开始处理请求", "method", "POST", "path", "/api/orders")
	requestLogger.Info("参数验证通过", "order_amount", 9900)
	requestLogger.Info("订单创建成功", "order_id", "ORD-1001")
}

// demoLogGroup 演示日志分组
func demoLogGroup(logger *slog.Logger) {
	fmt.Println("\n--- 3. 日志分组（slog.Group）---")

	// 使用 slog.Group 将相关字段分组
	logger.Info("HTTP 请求完成",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.String("path", "/api/users/42"),
			slog.String("client_ip", "192.168.1.100"),
		),
		slog.Group("response",
			slog.Int("status", 200),
			slog.Duration("latency", 42*time.Millisecond),
			slog.Int("body_size", 256),
		),
	)
}

// demoWithGroup 演示 WithGroup 创建分组 Logger
func demoWithGroup(logger *slog.Logger) {
	fmt.Println("\n--- 4. WithGroup 分组 Logger ---")

	// 创建带分组前缀的 Logger
	dbLogger := logger.WithGroup("database")
	dbLogger.Info("执行查询",
		"table", "users",
		"duration_ms", 15,
	)
	dbLogger.Warn("慢查询检测",
		"sql", "SELECT * FROM orders WHERE ...",
		"duration_ms", 2500,
	)
}

// demoJSONHandler 演示标准 JSON Handler 对比
func demoJSONHandler() {
	fmt.Println("\n--- 5. 标准 JSONHandler 对比 ---")

	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	jsonLogger.Info("JSON 格式日志",
		"service", "user-api",
		"version", "1.2.0",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.Int("status", 200),
		),
	)
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println("========== slog 自定义 Handler 演示 ==========")

	// 创建自定义 ColorHandler
	handler := NewColorHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // 设置为 DEBUG 级别以演示过滤
		AddSource: false,
	})

	// 创建带固定字段的 Logger
	logger := slog.New(handler).With("service", "demo-api")

	// 设置为全局默认 Logger
	slog.SetDefault(logger)

	// 演示各种用法
	demoBasicUsage(logger)
	demoContextPropagation(logger)
	demoLogGroup(logger)
	demoWithGroup(logger)
	demoJSONHandler()

	fmt.Println("\n========== 演示完成 ==========")
	fmt.Println("自定义 ColorHandler 实现了 slog.Handler 接口的四个方法：")
	fmt.Println("  - Enabled: 日志级别过滤")
	fmt.Println("  - Handle: 格式化输出（带颜色）")
	fmt.Println("  - WithAttrs: 添加固定属性")
	fmt.Println("  - WithGroup: 添加分组前缀")
}
