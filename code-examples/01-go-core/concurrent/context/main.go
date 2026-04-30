// 并发编程 — context 包（传播/超时控制）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 context 包的核心用法：
// 1. WithCancel —— 手动取消
// 2. WithTimeout —— 超时自动取消
// 3. WithDeadline —— 截止时间取消
// 4. WithValue —— 传递请求范围值
// 5. context 树形传播
// 6. 常见错误演示
//
// 运行方式：go run main.go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("========== context 包示例 ==========")

	// --- 1. WithCancel ---
	demoWithCancel()

	// --- 2. WithTimeout ---
	demoWithTimeout()

	// --- 3. WithDeadline ---
	demoWithDeadline()

	// --- 4. WithValue ---
	demoWithValue()

	// --- 5. 树形传播 ---
	demoTreePropagation()

	// --- 6. 常见错误 ---
	demoCommonMistakes()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. WithCancel —— 手动取消
// ============================================================

func demoWithCancel() {
	fmt.Println("\n--- 1. WithCancel 手动取消 ---")

	ctx, cancel := context.WithCancel(context.Background())

	// 启动一个持续工作的 goroutine
	go func(ctx context.Context) {
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				fmt.Printf("  Worker 收到取消信号: %v\n", ctx.Err())
				return
			default:
				fmt.Printf("  Worker 正在工作... 第 %d 次\n", i)
				time.Sleep(50 * time.Millisecond)
			}
		}
	}(ctx)

	// 让 worker 工作一段时间后取消
	time.Sleep(160 * time.Millisecond)
	cancel() // 发送取消信号
	time.Sleep(50 * time.Millisecond)
	fmt.Println("  ✅ Worker 已停止")
}

// ============================================================
// 2. WithTimeout —— 超时自动取消
// ============================================================

// simulateSlowAPI 模拟一个可能超时的 API 调用
func simulateSlowAPI(ctx context.Context, delay time.Duration) (string, error) {
	select {
	case <-time.After(delay):
		return "API 响应数据", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func demoWithTimeout() {
	fmt.Println("\n--- 2. WithTimeout 超时控制 ---")

	// 场景 1：API 在超时前完成
	ctx1, cancel1 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel1() // ✅ 始终调用 cancel 释放资源

	result, err := simulateSlowAPI(ctx1, 50*time.Millisecond)
	if err != nil {
		fmt.Printf("  场景 1 失败: %v\n", err)
	} else {
		fmt.Printf("  场景 1 成功: %s\n", result)
	}

	// 场景 2：API 超时
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	result, err = simulateSlowAPI(ctx2, 200*time.Millisecond)
	if err != nil {
		fmt.Printf("  场景 2 超时: %v\n", err)
	} else {
		fmt.Printf("  场景 2 成功: %s\n", result)
	}
}

// ============================================================
// 3. WithDeadline —— 截止时间取消
// ============================================================

func demoWithDeadline() {
	fmt.Println("\n--- 3. WithDeadline 截止时间 ---")

	deadline := time.Now().Add(150 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// 检查截止时间
	if d, ok := ctx.Deadline(); ok {
		fmt.Printf("  截止时间: %v（距现在 %v）\n", d.Format("15:04:05.000"), time.Until(d).Round(time.Millisecond))
	}

	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Println("  操作完成")
	case <-ctx.Done():
		fmt.Printf("  到达截止时间: %v\n", ctx.Err())
	}
}

// ============================================================
// 4. WithValue —— 传递请求范围值
// ============================================================

// 使用自定义类型作为 key，避免包间冲突
type contextKey string

const (
	traceIDKey contextKey = "traceID"
	userIDKey  contextKey = "userID"
)

// processRequest 模拟请求处理链路
func processRequest(ctx context.Context) {
	traceID := ctx.Value(traceIDKey)
	userID := ctx.Value(userIDKey)
	fmt.Printf("  处理请求: traceID=%v, userID=%v\n", traceID, userID)

	// 调用下游服务（context 自动传递）
	callDownstream(ctx)
}

func callDownstream(ctx context.Context) {
	traceID := ctx.Value(traceIDKey)
	fmt.Printf("  下游服务: traceID=%v（从 context 中获取）\n", traceID)
}

func demoWithValue() {
	fmt.Println("\n--- 4. WithValue 传递请求范围值 ---")

	// 模拟 HTTP 请求处理：在中间件中注入 traceID 和 userID
	ctx := context.Background()
	ctx = context.WithValue(ctx, traceIDKey, "trace-abc-123")
	ctx = context.WithValue(ctx, userIDKey, "user-42")

	processRequest(ctx)

	// ⚠️ 注意：WithValue 的 key 应使用自定义类型
	fmt.Println("  ⚠️ key 应使用自定义类型（如 contextKey），避免字符串冲突")
}

// ============================================================
// 5. context 树形传播
// ============================================================

func demoTreePropagation() {
	fmt.Println("\n--- 5. context 树形传播 ---")

	// 创建 context 树：
	// Background → parentCtx(cancel) → childCtx1(timeout) + childCtx2(value)
	parentCtx, parentCancel := context.WithCancel(context.Background())

	childCtx1, childCancel1 := context.WithTimeout(parentCtx, 500*time.Millisecond)
	defer childCancel1()

	childCtx2 := context.WithValue(parentCtx, traceIDKey, "trace-tree-demo")

	// 启动使用不同 context 的 goroutine
	done := make(chan string, 2)

	go func() {
		select {
		case <-childCtx1.Done():
			done <- fmt.Sprintf("child1 取消: %v", childCtx1.Err())
		}
	}()

	go func() {
		select {
		case <-childCtx2.Done():
			done <- fmt.Sprintf("child2 取消: %v (traceID=%v)", childCtx2.Err(), childCtx2.Value(traceIDKey))
		}
	}()

	// 取消父 context → 所有子 context 自动取消
	fmt.Println("  取消父 context...")
	parentCancel()

	// 收集结果
	for i := 0; i < 2; i++ {
		msg := <-done
		fmt.Printf("  %s\n", msg)
	}
	fmt.Println("  ✅ 父 context 取消后，所有子 context 都被取消")
}

// ============================================================
// 6. 常见错误演示
// ============================================================

func demoCommonMistakes() {
	fmt.Println("\n--- 6. 常见错误 ---")

	// ❌ 错误 1：忘记调用 cancel
	fmt.Println("  ❌ 忘记调用 cancel 导致资源泄漏")
	fmt.Println("     始终使用 defer cancel()")

	// ❌ 错误 2：WithValue 传递业务参数
	fmt.Println("  ❌ 不要用 WithValue 传递业务参数")
	fmt.Println("     只传递请求范围的元数据（traceID、userID）")

	// ❌ 错误 3：使用 string 类型作为 key
	fmt.Println("  ❌ 不要使用 string 作为 WithValue 的 key")
	fmt.Println("     应使用自定义类型避免包间冲突")

	// ✅ 正确模式
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() // ✅ 始终 defer cancel
	_ = ctx
	fmt.Println("  ✅ 正确: ctx, cancel := ...; defer cancel()")
}
