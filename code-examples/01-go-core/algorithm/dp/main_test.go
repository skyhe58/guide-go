// 动态规划算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import "testing"

// TestClimbStairs 爬楼梯测试
func TestClimbStairs(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"1 阶", 1, 1},
		{"2 阶", 2, 2},
		{"3 阶", 3, 3},
		{"4 阶", 4, 5},
		{"5 阶", 5, 8},
		{"10 阶", 10, 89},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := climbStairs(tt.n)
			if got != tt.want {
				t.Errorf("climbStairs(%d) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}

// TestLengthOfLIS 最长递增子序列测试
func TestLengthOfLIS(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"基本用例", []int{10, 9, 2, 5, 3, 7, 101, 18}, 4},
		{"全递增", []int{1, 2, 3, 4, 5}, 5},
		{"全递减", []int{5, 4, 3, 2, 1}, 1},
		{"单个元素", []int{7}, 1},
		{"有重复", []int{1, 3, 6, 7, 9, 4, 10, 5, 6}, 6},
		{"空数组", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lengthOfLIS(tt.nums)
			if got != tt.want {
				t.Errorf("lengthOfLIS(%v) = %d, want %d", tt.nums, got, tt.want)
			}
		})
	}
}

// TestKnapsack 0-1 背包测试
func TestKnapsack(t *testing.T) {
	tests := []struct {
		name     string
		weights  []int
		values   []int
		capacity int
		want     int
	}{
		{"基本用例", []int{2, 3, 4, 5}, []int{3, 4, 5, 6}, 8, 10},
		{"容量为 0", []int{1, 2}, []int{3, 4}, 0, 0},
		{"单个物品可放", []int{3}, []int{5}, 5, 5},
		{"单个物品放不下", []int{6}, []int{5}, 5, 0},
		{"全部可放", []int{1, 1, 1}, []int{1, 2, 3}, 3, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := knapsack(tt.weights, tt.values, tt.capacity)
			if got != tt.want {
				t.Errorf("knapsack(%v, %v, %d) = %d, want %d",
					tt.weights, tt.values, tt.capacity, got, tt.want)
			}
		})
	}
}

// TestMinDistance 编辑距离测试
func TestMinDistance(t *testing.T) {
	tests := []struct {
		name  string
		word1 string
		word2 string
		want  int
	}{
		{"horse->ros", "horse", "ros", 3},
		{"intention->execution", "intention", "execution", 5},
		{"空串->abc", "", "abc", 3},
		{"abc->空串", "abc", "", 3},
		{"相同字符串", "abc", "abc", 0},
		{"都为空", "", "", 0},
		{"单字符不同", "a", "b", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minDistance(tt.word1, tt.word2)
			if got != tt.want {
				t.Errorf("minDistance(%q, %q) = %d, want %d",
					tt.word1, tt.word2, got, tt.want)
			}
		})
	}
}
