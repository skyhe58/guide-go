// OpenTelemetry 链路追踪示例 — Tracer/Span/属性/事件/上下文传播/父子 Span
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 OpenTelemetry Go SDK 的链路追踪功能：
// - Tracer Provider 初始化（stdout exporter）
// - Span 创建、属性设置、事件记录
// - Context 传播实现父子 Span 关系
// - 模拟完整的请求处理链路（HTTP → Service → DB）
// - 错误 Span 标记
//
// 运行方式：go run ./otel-tracing/
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// ============================================================
// OpenTelemetry 初始化
// ============================================================

// initTracer 初始化 OTel Tracer Provider
// 使用 stdout exporter 将 Span 输出到控制台（生产环境替换为 Jaeger/OTLP exporter）
func initTracer() (*sdktrace.TracerProvider, error) {
	// 创建 stdout exporter（输出到 stderr 避免与业务输出混淆）
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stderr),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 stdout exporter 失败: %w", err)
	}

	// 定义服务资源信息
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("order-service"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 resource 失败: %w", err)
	}

	// 创建 Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// 开发环境采样所有请求
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 设置全局 Tracer Provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// ============================================================
// 业务逻辑：模拟订单处理链路
// 通过 context 传播实现父子 Span 关系
// ============================================================

// 获取全局 Tracer
var tracer = otel.Tracer("order-service")

// handleCreateOrder 处理创建订单请求（顶层 Span）
func handleCreateOrder(ctx context.Context, orderID string, userID string, amount int) error {
	// 创建顶层 Span
	ctx, span := tracer.Start(ctx, "HTTP POST /api/orders",
		trace.WithAttributes(
			attribute.String("http.method", "POST"),
			attribute.String("http.route", "/api/orders"),
			attribute.String("order.id", orderID),
			attribute.String("user.id", userID),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// 添加事件：请求开始
	span.AddEvent("请求参数验证通过", trace.WithAttributes(
		attribute.Int("order.amount", amount),
	))

	fmt.Printf("[Trace] 开始处理订单: orderID=%s, userID=%s, amount=%d\n", orderID, userID, amount)

	// 调用业务服务层（子 Span）
	if err := createOrder(ctx, orderID, userID, amount); err != nil {
		// 标记 Span 为错误状态
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		span.SetAttributes(attribute.Int("http.status_code", 500))
		return err
	}

	span.SetAttributes(attribute.Int("http.status_code", 201))
	span.AddEvent("订单创建成功")
	return nil
}

// createOrder 订单服务层（子 Span）
func createOrder(ctx context.Context, orderID string, userID string, amount int) error {
	ctx, span := tracer.Start(ctx, "OrderService.CreateOrder",
		trace.WithAttributes(
			attribute.String("order.id", orderID),
		),
	)
	defer span.End()

	// 步骤 1：验证用户
	if err := validateUser(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "用户验证失败")
		return err
	}

	// 步骤 2：检查库存
	if err := checkInventory(ctx, orderID, amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "库存检查失败")
		return err
	}

	// 步骤 3：写入数据库
	if err := saveOrder(ctx, orderID, userID, amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "数据库写入失败")
		return err
	}

	span.AddEvent("订单处理完成")
	return nil
}

// validateUser 验证用户（子 Span）
func validateUser(ctx context.Context, userID string) error {
	_, span := tracer.Start(ctx, "UserService.ValidateUser",
		trace.WithAttributes(
			attribute.String("user.id", userID),
		),
	)
	defer span.End()

	// 模拟 RPC 调用延迟
	time.Sleep(time.Duration(5+rand.Intn(15)) * time.Millisecond)

	span.AddEvent("用户验证通过", trace.WithAttributes(
		attribute.String("user.role", "premium"),
	))

	fmt.Printf("[Trace]   ├── 用户验证通过: userID=%s\n", userID)
	return nil
}

// checkInventory 检查库存（子 Span）
func checkInventory(ctx context.Context, orderID string, amount int) error {
	_, span := tracer.Start(ctx, "InventoryService.CheckStock",
		trace.WithAttributes(
			attribute.String("order.id", orderID),
			attribute.Int("requested.quantity", 1),
		),
	)
	defer span.End()

	// 模拟库存查询延迟
	time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)

	span.AddEvent("库存充足", trace.WithAttributes(
		attribute.Int("available.quantity", 100),
	))

	fmt.Printf("[Trace]   ├── 库存检查通过: orderID=%s\n", orderID)
	return nil
}

// saveOrder 保存订单到数据库（子 Span）
func saveOrder(ctx context.Context, orderID string, userID string, amount int) error {
	_, span := tracer.Start(ctx, "DB.InsertOrder",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.sql.table", "orders"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// 模拟数据库写入延迟
	time.Sleep(time.Duration(20+rand.Intn(50)) * time.Millisecond)

	span.AddEvent("SQL 执行完成", trace.WithAttributes(
		attribute.String("db.statement", "INSERT INTO orders (id, user_id, amount) VALUES ($1, $2, $3)"),
		attribute.Int("db.rows_affected", 1),
	))

	fmt.Printf("[Trace]   └── 数据库写入完成: orderID=%s\n", orderID)
	return nil
}

// ============================================================
// 演示错误链路
// ============================================================

// handleFailedOrder 模拟失败的订单（演示错误 Span）
func handleFailedOrder(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "HTTP POST /api/orders",
		trace.WithAttributes(
			attribute.String("http.method", "POST"),
			attribute.String("http.route", "/api/orders"),
			attribute.String("order.id", "ORD-FAIL"),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	// 模拟数据库连接失败
	_, dbSpan := tracer.Start(ctx, "DB.InsertOrder",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	err := fmt.Errorf("数据库连接超时: dial tcp localhost:5432: i/o timeout")
	dbSpan.RecordError(err)
	dbSpan.SetStatus(codes.Error, err.Error())
	dbSpan.End()

	span.SetStatus(codes.Error, "订单创建失败")
	span.SetAttributes(attribute.Int("http.status_code", 500))

	return err
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println("========== OpenTelemetry 链路追踪演示 ==========")
	fmt.Println("Span 数据输出到 stderr（JSON 格式），业务日志输出到 stdout")
	fmt.Println()

	// 初始化 Tracer
	tp, err := initTracer()
	if err != nil {
		fmt.Printf("初始化 Tracer 失败: %v\n", err)
		return
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("关闭 Tracer Provider 失败: %v\n", err)
		}
	}()

	ctx := context.Background()

	// 演示 1：成功的订单链路
	fmt.Println("--- 1. 成功的订单链路 ---")
	fmt.Println("链路结构：HTTP → OrderService → UserService + InventoryService + DB")
	if err := handleCreateOrder(ctx, "ORD-1001", "user-42", 9900); err != nil {
		fmt.Printf("订单创建失败: %v\n", err)
	} else {
		fmt.Println("订单创建成功！")
	}
	fmt.Println()

	// 演示 2：另一个成功的订单
	fmt.Println("--- 2. 另一个订单链路 ---")
	if err := handleCreateOrder(ctx, "ORD-1002", "user-88", 19900); err != nil {
		fmt.Printf("订单创建失败: %v\n", err)
	} else {
		fmt.Println("订单创建成功！")
	}
	fmt.Println()

	// 演示 3：失败的订单链路（错误 Span）
	fmt.Println("--- 3. 失败的订单链路（错误 Span）---")
	if err := handleFailedOrder(ctx); err != nil {
		fmt.Printf("订单创建失败（预期）: %v\n", err)
	}
	fmt.Println()

	// 等待 exporter 刷新
	time.Sleep(100 * time.Millisecond)

	fmt.Println("========== 演示完成 ==========")
	fmt.Println("每个 Span 包含：")
	fmt.Println("  - TraceID: 全局唯一，串联整个请求链路")
	fmt.Println("  - SpanID: 当前操作的唯一标识")
	fmt.Println("  - ParentSpanID: 父操作标识（实现父子关系）")
	fmt.Println("  - Attributes: 键值对属性（如 order.id, http.method）")
	fmt.Println("  - Events: 时间点事件（如 '库存检查通过'）")
	fmt.Println("  - Status: 操作状态（OK/Error）")
	fmt.Println()
	fmt.Println("生产环境中，将 stdout exporter 替换为 OTLP exporter，")
	fmt.Println("发送到 Jaeger/Tempo 后端，即可在 UI 中查看完整链路图。")
}
