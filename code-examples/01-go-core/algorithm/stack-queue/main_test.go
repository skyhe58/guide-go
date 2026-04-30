// 栈与队列算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import "testing"

// TestIsValid 有效括号测试
func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"简单匹配", "()", true},
		{"多种括号", "()[]{}", true},
		{"不匹配", "(]", false},
		{"交叉不匹配", "([)]", false},
		{"嵌套匹配", "{[]}", true},
		{"空字符串", "", true},
		{"单个左括号", "(", false},
		{"单个右括号", ")", false},
		{"深层嵌套", "((([{}])))", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValid(tt.input)
			if got != tt.want {
				t.Errorf("isValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestMinStack 最小栈测试
func TestMinStack(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		s := NewMinStack()
		s.Push(-2)
		s.Push(0)
		s.Push(-3)

		if got := s.GetMin(); got != -3 {
			t.Errorf("GetMin() = %d, want -3", got)
		}
		s.Pop()
		if got := s.Top(); got != 0 {
			t.Errorf("Top() = %d, want 0", got)
		}
		if got := s.GetMin(); got != -2 {
			t.Errorf("GetMin() = %d, want -2", got)
		}
	})

	t.Run("重复最小值", func(t *testing.T) {
		s := NewMinStack()
		s.Push(1)
		s.Push(1)
		s.Push(1)

		if got := s.GetMin(); got != 1 {
			t.Errorf("GetMin() = %d, want 1", got)
		}
		s.Pop()
		if got := s.GetMin(); got != 1 {
			t.Errorf("GetMin() after pop = %d, want 1", got)
		}
	})
}

// TestMyQueue 用栈实现队列测试
func TestMyQueue(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		q := NewMyQueue()
		q.Push(1)
		q.Push(2)

		if got := q.Peek(); got != 1 {
			t.Errorf("Peek() = %d, want 1", got)
		}
		if got := q.Pop(); got != 1 {
			t.Errorf("Pop() = %d, want 1", got)
		}
		if got := q.Empty(); got != false {
			t.Errorf("Empty() = %v, want false", got)
		}
	})

	t.Run("交替入队出队", func(t *testing.T) {
		q := NewMyQueue()
		q.Push(1)
		q.Push(2)
		if got := q.Pop(); got != 1 {
			t.Errorf("Pop() = %d, want 1", got)
		}
		q.Push(3)
		if got := q.Pop(); got != 2 {
			t.Errorf("Pop() = %d, want 2", got)
		}
		if got := q.Pop(); got != 3 {
			t.Errorf("Pop() = %d, want 3", got)
		}
		if got := q.Empty(); got != true {
			t.Errorf("Empty() = %v, want true", got)
		}
	})
}
