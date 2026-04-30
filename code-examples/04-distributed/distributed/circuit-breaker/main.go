// 熔断器 — 完整状态机实现
// 演示：Closed → Open → Half-Open 状态转换，模拟服务故障与恢复
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./circuit-breaker/
//
// 本示例为纯 Go 实现，无需外部依赖

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 熔断器状态定义
// ============================================================

// State 熔断器状态
type State int

const (
	// StateClosed 关闭状态（正常）：请求正常通过，统计失败次数
	StateClosed State = iota
	// StateOpen 打开状态（熔断）：请求直接拒绝，返回降级结果
	StateOpen
	// StateHalfOpen 半开状态（探测）：允许少量请求通过，根据结果决定恢复或继续熔断
	StateHalfOpen
)

// String 状态的字符串表示
func (s State) String() string {
	switch s {
	case StateClosed:
		return "Closed（正常）"
	case StateOpen:
		return "Open（熔断）"
	case StateHalfOpen:
		return "Half-Open（探测）"
	default:
		return "Unknown"
	}
}

// ============================================================
// 熔断器核心实现
// ============================================================

// ErrCircuitOpen 熔断器打开时返回的错误
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Counts 请求计数器
type Counts struct {
	Requests             int // 总请求数
	TotalSuccesses       int // 总成功数
	TotalFailures        int // 总失败数
	ConsecutiveSuccesses int // 连续成功数
	ConsecutiveFailures  int // 连续失败数
}

// onSuccess 记录一次成功
func (c *Counts) onSuccess() {
	c.Requests++
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

// onFailure 记录一次失败
func (c *Counts) onFailure() {
	c.Requests++
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

// reset 重置计数器
func (c *Counts) reset() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// Settings 熔断器配置
type Settings struct {
	Name string // 熔断器名称（标识下游服务）

	// 触发熔断的条件
	FailureThreshold int // 连续失败多少次触发熔断（默认 5）

	// 恢复探测配置
	Timeout        time.Duration // 熔断后多久进入半开状态（默认 10s）
	MaxHalfOpenReq int           // 半开状态允许通过的最大请求数（默认 1）

	// 成功恢复条件
	SuccessThreshold int // 半开状态连续成功多少次恢复为关闭（默认 1）

	// 回调函数
	OnStateChange func(name string, from, to State) // 状态变更回调
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu sync.Mutex

	name             string
	state            State
	counts           Counts
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	maxHalfOpenReq   int
	halfOpenReq      int // 半开状态已放行的请求数
	lastStateChange  time.Time
	onStateChange    func(string, State, State)
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(settings Settings) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             settings.Name,
		state:            StateClosed,
		failureThreshold: settings.FailureThreshold,
		successThreshold: settings.SuccessThreshold,
		timeout:          settings.Timeout,
		maxHalfOpenReq:   settings.MaxHalfOpenReq,
		lastStateChange:  time.Now(),
		onStateChange:    settings.OnStateChange,
	}

	// 设置默认值
	if cb.failureThreshold <= 0 {
		cb.failureThreshold = 5
	}
	if cb.successThreshold <= 0 {
		cb.successThreshold = 1
	}
	if cb.timeout <= 0 {
		cb.timeout = 10 * time.Second
	}
	if cb.maxHalfOpenReq <= 0 {
		cb.maxHalfOpenReq = 1
	}

	return cb
}

// Execute 通过熔断器执行请求
// 如果熔断器打开，直接返回 ErrCircuitOpen
// 否则执行 fn，根据结果更新熔断器状态
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// 检查是否允许请求通过
	if err := cb.beforeRequest(); err != nil {
		cb.mu.Unlock()
		return err
	}
	cb.mu.Unlock()

	// 执行实际请求
	err := fn()

	// 根据结果更新状态
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return err
}

// beforeRequest 请求前检查，决定是否放行
func (cb *CircuitBreaker) beforeRequest() error {
	switch cb.state {
	case StateClosed:
		// 关闭状态：正常放行
		return nil

	case StateOpen:
		// 打开状态：检查是否超时，超时则转为半开
		if time.Since(cb.lastStateChange) >= cb.timeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenReq = 1 // 放行第一个探测请求
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// 半开状态：限制放行请求数
		if cb.halfOpenReq < cb.maxHalfOpenReq {
			cb.halfOpenReq++
			return nil
		}
		return ErrCircuitOpen

	default:
		return ErrCircuitOpen
	}
}

// onSuccess 请求成功时的处理
func (cb *CircuitBreaker) onSuccess() {
	cb.counts.onSuccess()

	switch cb.state {
	case StateClosed:
		// 关闭状态：成功不需要特殊处理
	case StateHalfOpen:
		// 半开状态：连续成功达到阈值，恢复为关闭
		if cb.counts.ConsecutiveSuccesses >= cb.successThreshold {
			cb.setState(StateClosed)
		}
	}
}

// onFailure 请求失败时的处理
func (cb *CircuitBreaker) onFailure() {
	cb.counts.onFailure()

	switch cb.state {
	case StateClosed:
		// 关闭状态：连续失败达到阈值，触发熔断
		if cb.counts.ConsecutiveFailures >= cb.failureThreshold {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		// 半开状态：探测失败，重新熔断
		cb.setState(StateOpen)
	}
}

// setState 切换状态
func (cb *CircuitBreaker) setState(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()
	cb.counts.reset()
	cb.halfOpenReq = 0

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, newState)
	}
}

// State 获取当前状态（线程安全）
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Counts 获取当前计数（线程安全）
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.counts
}

// ============================================================
// 模拟下游服务
// ============================================================

// DownstreamService 模拟下游服务
type DownstreamService struct {
	mu        sync.Mutex
	healthy   bool
	failRate  float64 // 故障率（0.0 ~ 1.0）
	callCount atomic.Int64
}

// NewDownstreamService 创建模拟服务
func NewDownstreamService(healthy bool) *DownstreamService {
	return &DownstreamService{
		healthy:  healthy,
		failRate: 0.0,
	}
}

// SetHealthy 设置服务健康状态
func (ds *DownstreamService) SetHealthy(healthy bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.healthy = healthy
	if healthy {
		ds.failRate = 0.0
	} else {
		ds.failRate = 1.0
	}
}

// SetFailRate 设置故障率
func (ds *DownstreamService) SetFailRate(rate float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.failRate = rate
}

// Call 调用服务
func (ds *DownstreamService) Call() error {
	ds.callCount.Add(1)
	ds.mu.Lock()
	failRate := ds.failRate
	ds.mu.Unlock()

	// 模拟网络延迟
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	if rand.Float64() < failRate {
		return errors.New("service unavailable")
	}
	return nil
}

// ============================================================
// 演示场景
// ============================================================

func main() {
	fmt.Println("=== 熔断器状态机演示 ===")
	fmt.Println()

	// 场景一：基本状态转换
	demoBasicStateTransition()

	// 场景二：并发请求下的熔断保护
	demoConcurrentProtection()

	// 场景三：服务恢复后自动恢复
	demoAutoRecovery()
}

// demoBasicStateTransition 演示基本状态转换流程
func demoBasicStateTransition() {
	fmt.Println("--- 场景一：基本状态转换 ---")
	fmt.Println("Closed → Open → Half-Open → Closed")
	fmt.Println()

	service := NewDownstreamService(true)

	cb := NewCircuitBreaker(Settings{
		Name:             "payment-service",
		FailureThreshold: 3,           // 连续失败 3 次触发熔断
		Timeout:          time.Second,  // 1 秒后进入半开状态（演示用，生产环境建议 10-60s）
		MaxHalfOpenReq:   1,           // 半开状态允许 1 个探测请求
		SuccessThreshold: 1,           // 探测成功 1 次恢复
		OnStateChange: func(name string, from, to State) {
			fmt.Printf("  ⚡ [%s] 状态变更: %s → %s\n", name, from, to)
		},
	})

	// 阶段 1：正常请求（Closed 状态）
	fmt.Println("  阶段 1：正常请求")
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error { return service.Call() })
		fmt.Printf("    请求 %d: err=%v, 状态=%s\n", i+1, err, cb.State())
	}
	fmt.Println()

	// 阶段 2：服务故障，触发熔断
	fmt.Println("  阶段 2：服务故障，触发熔断")
	service.SetHealthy(false)
	for i := 0; i < 5; i++ {
		err := cb.Execute(func() error { return service.Call() })
		fmt.Printf("    请求 %d: err=%v, 状态=%s\n", i+1, err, cb.State())
	}
	fmt.Println()

	// 阶段 3：等待超时，进入半开状态
	fmt.Println("  阶段 3：等待超时，服务恢复")
	time.Sleep(1100 * time.Millisecond)
	service.SetHealthy(true)

	err := cb.Execute(func() error { return service.Call() })
	fmt.Printf("    探测请求: err=%v, 状态=%s\n", err, cb.State())
	fmt.Println()
}

// demoConcurrentProtection 演示并发场景下的熔断保护
func demoConcurrentProtection() {
	fmt.Println("--- 场景二：并发请求下的熔断保护 ---")
	fmt.Println("50 个 goroutine 并发请求，服务故障后熔断器保护")
	fmt.Println()

	service := NewDownstreamService(true)

	cb := NewCircuitBreaker(Settings{
		Name:             "order-service",
		FailureThreshold: 5,
		Timeout:          2 * time.Second,
		MaxHalfOpenReq:   2,
		SuccessThreshold: 2,
		OnStateChange: func(name string, from, to State) {
			fmt.Printf("  ⚡ [%s] 状态变更: %s → %s\n", name, from, to)
		},
	})

	var (
		success     atomic.Int64
		failed      atomic.Int64
		circuitOpen atomic.Int64
	)

	// 启动并发请求
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				err := cb.Execute(func() error { return service.Call() })
				if err == nil {
					success.Add(1)
				} else if errors.Is(err, ErrCircuitOpen) {
					circuitOpen.Add(1)
				} else {
					failed.Add(1)
				}
				time.Sleep(20 * time.Millisecond)
			}
		}(i)

		// 在第 100ms 时让服务故障
		if i == 10 {
			go func() {
				time.Sleep(50 * time.Millisecond)
				fmt.Println("  💥 服务故障！")
				service.SetHealthy(false)
			}()
		}
	}

	wg.Wait()

	fmt.Printf("\n  结果统计:\n")
	fmt.Printf("    成功: %d\n", success.Load())
	fmt.Printf("    失败（服务错误）: %d\n", failed.Load())
	fmt.Printf("    拒绝（熔断保护）: %d\n", circuitOpen.Load())
	fmt.Printf("    实际调用下游次数: %d（熔断器减少了 %d 次无效调用）\n",
		service.callCount.Load(),
		success.Load()+failed.Load()+circuitOpen.Load()-service.callCount.Load())
	fmt.Println()
}

// demoAutoRecovery 演示服务恢复后熔断器自动恢复
func demoAutoRecovery() {
	fmt.Println("--- 场景三：服务恢复后自动恢复 ---")
	fmt.Println("服务故障 → 熔断 → 服务恢复 → 半开探测 → 自动恢复")
	fmt.Println()

	service := NewDownstreamService(true)

	cb := NewCircuitBreaker(Settings{
		Name:             "inventory-service",
		FailureThreshold: 3,
		Timeout:          500 * time.Millisecond, // 演示用短超时
		MaxHalfOpenReq:   2,
		SuccessThreshold: 2,
		OnStateChange: func(name string, from, to State) {
			fmt.Printf("  ⚡ [%s] %s → %s (时间: %s)\n",
				name, from, to, time.Now().Format("15:04:05.000"))
		},
	})

	// 模拟时间线
	timeline := []struct {
		delay   time.Duration
		action  string
		healthy bool
	}{
		{0, "服务正常", true},
		{200 * time.Millisecond, "服务故障", false},
		{800 * time.Millisecond, "服务恢复", true},
	}

	// 启动状态变更
	go func() {
		for _, event := range timeline {
			time.Sleep(event.delay)
			fmt.Printf("  📌 [%s] %s\n", time.Now().Format("15:04:05.000"), event.action)
			service.SetHealthy(event.healthy)
		}
	}()

	// 持续发送请求
	for i := 0; i < 40; i++ {
		err := cb.Execute(func() error { return service.Call() })
		status := "✅"
		if errors.Is(err, ErrCircuitOpen) {
			status = "🚫 熔断拒绝"
		} else if err != nil {
			status = "❌ " + err.Error()
		}

		if i%5 == 0 {
			fmt.Printf("    请求 %2d: %s | 状态: %s\n", i+1, status, cb.State())
		}
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Println()
}
