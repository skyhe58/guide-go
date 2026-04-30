// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 切片表驱动测试示例
package main

import "testing"

// TestSum 表驱动测试：切片求和
func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{
			name:     "空切片",
			input:    []int{},
			expected: 0,
		},
		{
			name:     "单个元素",
			input:    []int{42},
			expected: 42,
		},
		{
			name:     "多个正数",
			input:    []int{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "包含负数",
			input:    []int{-1, 0, 1},
			expected: 0,
		},
		{
			name:     "全部负数",
			input:    []int{-1, -2, -3},
			expected: -6,
		},
		{
			name:     "nil 切片",
			input:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sum(tt.input)
			if got != tt.expected {
				t.Errorf("Sum(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestGrowSlice 表驱动测试：切片扩容
func TestGrowSlice(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{name: "0个元素", n: 0},
		{name: "1个元素", n: 1},
		{name: "10个元素", n: 10},
		{name: "100个元素", n: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := GrowSlice(tt.n)
			// 验证容量序列是递增的
			for i := 1; i < len(caps); i++ {
				if caps[i] <= caps[i-1] {
					t.Errorf("容量应递增: caps[%d]=%d <= caps[%d]=%d", i, caps[i], i-1, caps[i-1])
				}
			}
			// 验证最终容量 >= n
			if tt.n > 0 {
				finalCap := caps[len(caps)-1]
				if finalCap < tt.n {
					t.Errorf("最终容量 %d < 元素数 %d", finalCap, tt.n)
				}
			}
		})
	}
}

// TestSliceSharing 表驱动测试：底层数组共享
func TestSliceSharing(t *testing.T) {
	tests := []struct {
		name     string
		original []int
		sliceFrom int
		sliceTo   int
		modifyIdx int
		modifyVal int
	}{
		{
			name:      "修改子切片影响原切片",
			original:  []int{1, 2, 3, 4, 5},
			sliceFrom: 1,
			sliceTo:   3,
			modifyIdx: 0,
			modifyVal: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建原始切片的副本用于比较
			orig := make([]int, len(tt.original))
			copy(orig, tt.original)

			sub := orig[tt.sliceFrom:tt.sliceTo]
			sub[tt.modifyIdx] = tt.modifyVal

			// 验证原切片被修改
			expectedIdx := tt.sliceFrom + tt.modifyIdx
			if orig[expectedIdx] != tt.modifyVal {
				t.Errorf("原切片 orig[%d] = %d, want %d (底层数组应共享)",
					expectedIdx, orig[expectedIdx], tt.modifyVal)
			}
		})
	}
}
