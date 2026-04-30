// 哈希表算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import (
	"reflect"
	"sort"
	"testing"
)

// TestTwoSum 两数之和测试
func TestTwoSum(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   []int
	}{
		{"基本用例", []int{2, 7, 11, 15}, 9, []int{0, 1}},
		{"中间元素", []int{3, 2, 4}, 6, []int{1, 2}},
		{"相同元素", []int{3, 3}, 6, []int{0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := twoSum(tt.nums, tt.target)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("twoSum(%v, %d) = %v, want %v", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

// TestGroupAnagrams 字母异位词分组测试
func TestGroupAnagrams(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int // 期望的分组数量
	}{
		{"基本用例", []string{"eat", "tea", "tan", "ate", "nat", "bat"}, 3},
		{"空字符串", []string{""}, 1},
		{"单个字符", []string{"a"}, 1},
		{"无异位词", []string{"abc", "def", "ghi"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupAnagrams(tt.input)
			if len(got) != tt.want {
				t.Errorf("groupAnagrams(%v) 分组数 = %d, want %d", tt.input, len(got), tt.want)
			}
			// 验证每组内的元素确实是异位词
			for _, group := range got {
				if len(group) > 1 {
					sorted0 := sortString(group[0])
					for _, s := range group[1:] {
						if sortString(s) != sorted0 {
							t.Errorf("分组内 %q 和 %q 不是异位词", group[0], s)
						}
					}
				}
			}
		})
	}
}

// sortString 辅助函数：对字符串排序
func sortString(s string) string {
	bs := []byte(s)
	sort.Slice(bs, func(i, j int) bool { return bs[i] < bs[j] })
	return string(bs)
}

// TestLRUCache LRU 缓存测试
func TestLRUCache(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		lru := NewLRUCache(2)
		lru.Put(1, 1)
		lru.Put(2, 2)

		if got := lru.Get(1); got != 1 {
			t.Errorf("Get(1) = %d, want 1", got)
		}

		lru.Put(3, 3) // 淘汰 key=2

		if got := lru.Get(2); got != -1 {
			t.Errorf("Get(2) = %d, want -1 (已淘汰)", got)
		}

		lru.Put(4, 4) // 淘汰 key=1

		if got := lru.Get(1); got != -1 {
			t.Errorf("Get(1) = %d, want -1 (已淘汰)", got)
		}
		if got := lru.Get(3); got != 3 {
			t.Errorf("Get(3) = %d, want 3", got)
		}
		if got := lru.Get(4); got != 4 {
			t.Errorf("Get(4) = %d, want 4", got)
		}
	})

	t.Run("更新已有 key", func(t *testing.T) {
		lru := NewLRUCache(2)
		lru.Put(1, 1)
		lru.Put(2, 2)
		lru.Put(1, 10) // 更新 key=1

		if got := lru.Get(1); got != 10 {
			t.Errorf("Get(1) = %d, want 10", got)
		}
	})

	t.Run("容量为 1", func(t *testing.T) {
		lru := NewLRUCache(1)
		lru.Put(1, 1)
		lru.Put(2, 2) // 淘汰 key=1

		if got := lru.Get(1); got != -1 {
			t.Errorf("Get(1) = %d, want -1", got)
		}
		if got := lru.Get(2); got != 2 {
			t.Errorf("Get(2) = %d, want 2", got)
		}
	})
}
