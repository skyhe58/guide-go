// 栈与队列 — 面试高频栈和队列算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：有效括号、最小栈、用栈实现队列
// 运行方式：go run main.go
package main

import "fmt"

// isValid 有效括号匹配（LeetCode 20）
// 思路：遇到左括号入栈，遇到右括号检查栈顶是否匹配
// 时间复杂度：O(n)，空间复杂度：O(n)
func isValid(s string) bool {
	stack := []byte{}
	pairs := map[byte]byte{')': '(', ']': '[', '}': '{'}
	for i := 0; i < len(s); i++ {
		if s[i] == '(' || s[i] == '[' || s[i] == '{' {
			stack = append(stack, s[i]) // 左括号入栈
		} else {
			// 栈为空或栈顶不匹配
			if len(stack) == 0 || stack[len(stack)-1] != pairs[s[i]] {
				return false
			}
			stack = stack[:len(stack)-1] // 匹配成功，弹出栈顶
		}
	}
	return len(stack) == 0 // 栈为空说明全部匹配
}

// MinStack 最小栈（LeetCode 155）
// 在常数时间内获取栈中最小元素
// 思路：使用辅助栈同步记录当前最小值
type MinStack struct {
	stack    []int // 数据栈
	minStack []int // 辅助栈，栈顶始终是当前最小值
}

// NewMinStack 创建最小栈
func NewMinStack() *MinStack {
	return &MinStack{}
}

// Push 入栈
func (s *MinStack) Push(val int) {
	s.stack = append(s.stack, val)
	// 如果辅助栈为空或新值 <= 当前最小值，则入辅助栈
	if len(s.minStack) == 0 || val <= s.minStack[len(s.minStack)-1] {
		s.minStack = append(s.minStack, val)
	}
}

// Pop 出栈
func (s *MinStack) Pop() {
	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	// 如果弹出的是当前最小值，辅助栈也弹出
	if top == s.minStack[len(s.minStack)-1] {
		s.minStack = s.minStack[:len(s.minStack)-1]
	}
}

// Top 获取栈顶元素
func (s *MinStack) Top() int {
	return s.stack[len(s.stack)-1]
}

// GetMin 获取栈中最小元素，O(1) 时间
func (s *MinStack) GetMin() int {
	return s.minStack[len(s.minStack)-1]
}

// MyQueue 用两个栈实现队列（LeetCode 232）
// 思路：inStack 负责入队，outStack 负责出队
// 当 outStack 为空时，将 inStack 全部倒入 outStack
type MyQueue struct {
	inStack  []int // 入队栈
	outStack []int // 出队栈
}

// NewMyQueue 创建队列
func NewMyQueue() *MyQueue {
	return &MyQueue{}
}

// Push 入队
func (q *MyQueue) Push(x int) {
	q.inStack = append(q.inStack, x)
}

// transfer 将入队栈的元素转移到出队栈
func (q *MyQueue) transfer() {
	if len(q.outStack) == 0 {
		for len(q.inStack) > 0 {
			top := q.inStack[len(q.inStack)-1]
			q.inStack = q.inStack[:len(q.inStack)-1]
			q.outStack = append(q.outStack, top)
		}
	}
}

// Pop 出队
func (q *MyQueue) Pop() int {
	q.transfer()
	val := q.outStack[len(q.outStack)-1]
	q.outStack = q.outStack[:len(q.outStack)-1]
	return val
}

// Peek 查看队首元素
func (q *MyQueue) Peek() int {
	q.transfer()
	return q.outStack[len(q.outStack)-1]
}

// Empty 判断队列是否为空
func (q *MyQueue) Empty() bool {
	return len(q.inStack) == 0 && len(q.outStack) == 0
}

func main() {
	fmt.Println("=== 栈与队列算法示例 ===")

	// 1. 有效括号
	fmt.Println("\n--- 有效括号 ---")
	testCases := []string{"()", "()[]{}", "(]", "([)]", "{[]}"}
	for _, s := range testCases {
		fmt.Printf("  \"%s\" -> %v\n", s, isValid(s))
	}

	// 2. 最小栈
	fmt.Println("\n--- 最小栈 ---")
	ms := NewMinStack()
	ms.Push(-2)
	ms.Push(0)
	ms.Push(-3)
	fmt.Println("  最小值:", ms.GetMin()) // -3
	ms.Pop()
	fmt.Println("  弹出 -3 后栈顶:", ms.Top()) // 0
	fmt.Println("  最小值:", ms.GetMin())       // -2

	// 3. 用栈实现队列
	fmt.Println("\n--- 用栈实现队列 ---")
	q := NewMyQueue()
	q.Push(1)
	q.Push(2)
	fmt.Println("  Peek:", q.Peek()) // 1
	fmt.Println("  Pop:", q.Pop())   // 1
	fmt.Println("  Empty:", q.Empty()) // false
}
