// 并发编程 — errgroup（并发错误处理）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 errgroup 的核心用法：
// 1. 基本用法 —— 并发执行多个任务并收集错误
// 2. WithContext —— 任一任务失败时取消其余任务
// 3. SetLimit —— 限制并发 goroutine 数量
// 4. 实际场景 —— 并发获取多个 API 数据
// 5. 常见错误演示
//
// 依赖：golang.org/x/sync/errgroup
// 安装：go get golang.org/x/sync
//
// 运行方式：go run main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	fmt.Println("========== errgroup 示例 ==========")

	// --- 1. 基本用法 ---
	demoBasic()

	// --- 2. WithContext ---
	demoWithContext()

	// --- 3. SetLimit ---
	demoSetLimit()

	// --- 4. 实际场景 ---
	demoRealWorld()

	// --- 5. 常见错误 ---
	demoCommonMistakes()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. 基本用法 —— 并发执行并收集错误
// ============================================================

func demoBasic() {
	fmt.Println("\n--- 1. errgroup 基本用法 ---")

	g := new(errgroup.Group)

	// 并发执行 3 个任务
	g.Go(func() error {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  任务 1 完成")
		return nil
	})

	g.Go(func() error {
		time.Sleep(30 * time.Millisecond)
		fmt.Println("  任务 2 完成")
		return nil
	})

	g.Go(func() error {
		time.Sleep(40 * time.Millisecond)
		fmt.Println("  任务 3 完成")
		return nil
	})

	// Wait 等待所有任务完成，返回第一个错误
	if err := g.Wait(); err != nil {
		fmt.Printf("  ❌ 错误: %v\n", err)
	} else {
		fmt.Println("  ✅ 所有任务成功完成")
	}
}

// ============================================================
// 2. WithContext —— 任一失败时取消其余任务
// ============================================================

func demoWithContext() {
	fmt.Println("\n--- 2. WithContext 错误取消 ---")

	g, ctx := errgroup.WithContext(context.Background())

	// 任务 1：正常完成
	g.Go(func() error {
		select {
		case <-time.After(100 * time.Millisecond):
			fmt.Println("  任务 1 完成")
			return nil
		case <-ctx.Done():
			fmt.Println("  任务 1 被取消")
			return ctx.Err()
		}
	})

	// 任务 2：快速失败
	g.Go(func() error {
		time.Sleep(30 * time.Millisecond)
		fmt.Println("  任务 2 失败！")
		return errors.New("任务 2 发生错误")
	})

	// 任务 3：检查 context 取消
	g.Go(func() error {
		select {
		case <-time.After(200 * time.Millisecond):
			fmt.Println("  任务 3 完成")
			return nil
		case <-ctx.Done():
			fmt.Println("  任务 3 被取消（因为任务 2 失败）")
			return ctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("  ❌ 第一个错误: %v\n", err)
	}
	fmt.Println("  ✅ WithContext: 任一任务失败，其余任务自动取消")
}

// ============================================================
// 3. SetLimit —— 限制并发数量
// ============================================================

func demoSetLimit() {
	fmt.Println("\n--- 3. SetLimit 并发限制 ---")

	g := new(errgroup.Group)
	g.SetLimit(3) // 最多 3 个 goroutine 并发

	var running atomic.Int32

	for i := 1; i <= 10; i++ {
		i := i
		g.Go(func() error {
			current := running.Add(1)
			fmt.Printf("  任务 %d 开始（当前并发: %d）\n", i, current)
			time.Sleep(50 * time.Millisecond)
			running.Add(-1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("  ❌ 错误: %v\n", err)
	} else {
		fmt.Println("  ✅ 10 个任务完成，最大并发数限制为 3")
	}
}

// ============================================================
// 4. 实际场景 —— 并发获取多个 API 数据
// ============================================================

// UserProfile 用户资料
type UserProfile struct {
	Name  string
	Email string
}

// UserOrders 用户订单
type UserOrders struct {
	Count int
}

// UserStats 用户统计
type UserStats struct {
	LoginCount int
}

// fetchProfile 模拟获取用户资料
func fetchProfile(ctx context.Context) (*UserProfile, error) {
	select {
	case <-time.After(time.Duration(rand.Intn(50)) * time.Millisecond):
		return &UserProfile{Name: "张三", Email: "zhangsan@example.com"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// fetchOrders 模拟获取用户订单
func fetchOrders(ctx context.Context) (*UserOrders, error) {
	select {
	case <-time.After(time.Duration(rand.Intn(50)) * time.Millisecond):
		return &UserOrders{Count: 42}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// fetchStats 模拟获取用户统计
func fetchStats(ctx context.Context) (*UserStats, error) {
	select {
	case <-time.After(time.Duration(rand.Intn(50)) * time.Millisecond):
		return &UserStats{LoginCount: 128}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func demoRealWorld() {
	fmt.Println("\n--- 4. 实际场景：并发获取用户数据 ---")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	var profile *UserProfile
	var orders *UserOrders
	var stats *UserStats

	// 并发获取三种数据
	g.Go(func() error {
		var err error
		profile, err = fetchProfile(ctx)
		return err
	})

	g.Go(func() error {
		var err error
		orders, err = fetchOrders(ctx)
		return err
	})

	g.Go(func() error {
		var err error
		stats, err = fetchStats(ctx)
		return err
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("  ❌ 获取数据失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户: %s (%s)\n", profile.Name, profile.Email)
	fmt.Printf("  ✅ 订单数: %d\n", orders.Count)
	fmt.Printf("  ✅ 登录次数: %d\n", stats.LoginCount)
}

// ============================================================
// 5. 常见错误演示
// ============================================================

func demoCommonMistakes() {
	fmt.Println("\n--- 5. 常见错误 ---")

	// ❌ 错误 1：忽略 context 取消信号
	fmt.Println("  ❌ 使用 WithContext 时，goroutine 内部应检查 ctx.Done()")
	fmt.Println("     否则取消信号无法生效，goroutine 会继续运行")

	// ❌ 错误 2：期望收集所有错误
	fmt.Println("  ❌ errgroup.Wait() 只返回第一个非 nil 错误")
	fmt.Println("     如需收集所有错误，使用 channel 或 go-multierror 库")

	// ✅ 正确模式
	fmt.Println("  ✅ 正确模式:")
	fmt.Println("     g, ctx := errgroup.WithContext(ctx)")
	fmt.Println("     g.Go(func() error { select { case <-ctx.Done(): ... } })")
	fmt.Println("     if err := g.Wait(); err != nil { ... }")
}
