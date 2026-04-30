// 并发编程 — channel（各种用法/select）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 channel 的核心用法：
// 1. 无缓冲 channel —— 同步通信
// 2. 有缓冲 channel —— 异步通信
// 3. 方向限制 —— 只发送/只接收
// 4. select 多路复用
// 5. for-range 遍历 channel
// 6. 关闭语义和常见错误
//
// 运行方式：go run main.go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("========== channel 示例 ==========")

	// --- 1. 无缓冲 channel ---
	demoUnbuffered()

	// --- 2. 有缓冲 channel ---
	demoBuffered()

	// --- 3. 方向限制 ---
	demoDirectional()

	// --- 4. select 多路复用 ---
	demoSelect()

	// --- 5. for-range 遍历 ---
	demoForRange()

	// --- 6. 关闭语义 ---
	demoCloseSemantics()

	// --- 7. 常见错误演示 ---
	demoCommonMistakes()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. 无缓冲 channel —— 同步通信
// ============================================================

func demoUnbuffered() {
	fmt.Println("\n--- 1. 无缓冲 channel ---")

	ch := make(chan string) // 无缓冲：发送方阻塞直到接收方就绪

	go func() {
		// 发送方会阻塞，直到 main goroutine 准备接收
		ch <- "Hello from goroutine"
	}()

	msg := <-ch // 接收方阻塞，直到有数据可读
	fmt.Printf("  收到: %s\n", msg)

	// 无缓冲 channel 实现精确同步
	done := make(chan struct{})
	go func() {
		fmt.Println("  任务执行中...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  任务完成")
		close(done) // 用 close 发送信号比发送值更惯用
	}()
	<-done
	fmt.Println("  收到完成信号")
}

// ============================================================
// 2. 有缓冲 channel —— 异步通信
// ============================================================

func demoBuffered() {
	fmt.Println("\n--- 2. 有缓冲 channel ---")

	ch := make(chan int, 3) // 缓冲区大小为 3

	// 缓冲区未满时，发送不阻塞
	ch <- 1
	ch <- 2
	ch <- 3
	// ch <- 4 // ⚠️ 这里会阻塞，因为缓冲区已满

	fmt.Printf("  缓冲区长度: %d, 容量: %d\n", len(ch), cap(ch))

	// 接收数据
	fmt.Printf("  接收: %d, %d, %d\n", <-ch, <-ch, <-ch)
	fmt.Printf("  接收后长度: %d\n", len(ch))
}

// ============================================================
// 3. 方向限制 —— 只发送/只接收
// ============================================================

// producer 只能向 channel 发送数据
func producer(ch chan<- int, id int) {
	for i := 0; i < 3; i++ {
		val := id*100 + i
		ch <- val
		fmt.Printf("  生产者 %d 发送: %d\n", id, val)
	}
}

// consumer 只能从 channel 接收数据
func consumer(ch <-chan int, done chan<- struct{}) {
	for val := range ch {
		fmt.Printf("  消费者收到: %d\n", val)
	}
	done <- struct{}{}
}

func demoDirectional() {
	fmt.Println("\n--- 3. 方向限制 ---")

	ch := make(chan int, 10)
	done := make(chan struct{})

	// 双向 channel 自动转换为单向
	go producer(ch, 1)
	go producer(ch, 2)
	go consumer(ch, done)

	// 等待生产者完成后关闭 channel
	time.Sleep(100 * time.Millisecond)
	close(ch)
	<-done
}

// ============================================================
// 4. select 多路复用
// ============================================================

func demoSelect() {
	fmt.Println("\n--- 4. select 多路复用 ---")

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	go func() {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		ch1 <- "来自 channel 1"
	}()
	go func() {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		ch2 <- "来自 channel 2"
	}()

	// select 等待多个 channel，哪个先就绪就执行哪个
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("  select 收到: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("  select 收到: %s\n", msg)
		}
	}

	// select + timeout 模式
	fmt.Println("  select + timeout:")
	ch3 := make(chan string)
	go func() {
		time.Sleep(200 * time.Millisecond) // 模拟慢操作
		ch3 <- "慢操作结果"
	}()

	select {
	case msg := <-ch3:
		fmt.Printf("  收到: %s\n", msg)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  ⏰ 超时！")
	}

	// select + default 非阻塞模式
	fmt.Println("  select + default（非阻塞）:")
	ch4 := make(chan int, 1)
	select {
	case val := <-ch4:
		fmt.Printf("  收到: %d\n", val)
	default:
		fmt.Println("  channel 为空，执行 default")
	}
}

// ============================================================
// 5. for-range 遍历 channel
// ============================================================

func demoForRange() {
	fmt.Println("\n--- 5. for-range 遍历 channel ---")

	ch := make(chan int, 5)

	// 生产者发送数据后关闭 channel
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i * i
		}
		close(ch) // ✅ 关闭 channel，for-range 才能退出
	}()

	// for-range 持续接收直到 channel 关闭
	fmt.Print("  平方数: ")
	for val := range ch {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
}

// ============================================================
// 6. 关闭语义
// ============================================================

func demoCloseSemantics() {
	fmt.Println("\n--- 6. 关闭语义 ---")

	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	close(ch)

	// 从已关闭的 channel 接收：先读取缓冲区中的数据
	val1, ok1 := <-ch
	fmt.Printf("  第一次接收: val=%d, ok=%t\n", val1, ok1)

	val2, ok2 := <-ch
	fmt.Printf("  第二次接收: val=%d, ok=%t\n", val2, ok2)

	// 缓冲区为空后，返回零值和 false
	val3, ok3 := <-ch
	fmt.Printf("  第三次接收: val=%d, ok=%t（channel 已关闭且为空）\n", val3, ok3)

	// 使用 sync.Once 确保只关闭一次
	fmt.Println("  使用 sync.Once 安全关闭:")
	ch2 := make(chan int)
	var once sync.Once
	safeClose := func() {
		once.Do(func() {
			close(ch2)
			fmt.Println("  channel 已安全关闭")
		})
	}
	safeClose()
	safeClose() // 第二次调用不会 panic
}

// ============================================================
// 7. 常见错误演示
// ============================================================

func demoCommonMistakes() {
	fmt.Println("\n--- 7. 常见错误演示 ---")

	// ❌ 错误 1：向已关闭的 channel 发送数据会 panic
	fmt.Println("  ❌ 向已关闭的 channel 发送数据会 panic")
	fmt.Println("     示例: close(ch); ch <- 1 // panic: send on closed channel")

	// ❌ 错误 2：重复关闭 channel 会 panic
	fmt.Println("  ❌ 重复关闭 channel 会 panic")
	fmt.Println("     示例: close(ch); close(ch) // panic: close of closed channel")

	// ❌ 错误 3：nil channel 永久阻塞
	fmt.Println("  ❌ nil channel 的发送和接收都会永久阻塞")
	fmt.Println("     示例: var ch chan int; <-ch // 永久阻塞")

	// ✅ nil channel 在 select 中的妙用：禁用某个 case
	fmt.Println("  ✅ nil channel 在 select 中可以禁用 case:")
	ch1 := make(chan string, 1)
	ch1 <- "hello"
	var ch2 chan string // nil channel

	select {
	case msg := <-ch1:
		fmt.Printf("    从 ch1 收到: %s\n", msg)
	case msg := <-ch2:
		// 这个 case 永远不会被选中，因为 ch2 是 nil
		fmt.Printf("    从 ch2 收到: %s\n", msg)
	}
}
