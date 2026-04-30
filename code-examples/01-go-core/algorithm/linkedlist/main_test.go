// 链表算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import (
	"reflect"
	"testing"
)

// TestReverseList 反转链表测试
func TestReverseList(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"正常链表", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"两个元素", []int{1, 2}, []int{2, 1}},
		{"单个元素", []int{1}, []int{1}},
		{"空链表", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := buildList(tt.input)
			got := listToSlice(reverseList(head))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverseList(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestMergeTwoLists 合并两个有序链表测试
func TestMergeTwoLists(t *testing.T) {
	tests := []struct {
		name string
		l1   []int
		l2   []int
		want []int
	}{
		{"正常合并", []int{1, 3, 5}, []int{2, 4, 6}, []int{1, 2, 3, 4, 5, 6}},
		{"有重复元素", []int{1, 2, 4}, []int{1, 3, 4}, []int{1, 1, 2, 3, 4, 4}},
		{"l1 为空", nil, []int{1, 2}, []int{1, 2}},
		{"l2 为空", []int{1, 2}, nil, []int{1, 2}},
		{"都为空", nil, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l1 := buildList(tt.l1)
			l2 := buildList(tt.l2)
			got := listToSlice(mergeTwoLists(l1, l2))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeTwoLists(%v, %v) = %v, want %v", tt.l1, tt.l2, got, tt.want)
			}
		})
	}
}

// TestHasCycle 环形链表检测测试
func TestHasCycle(t *testing.T) {
	tests := []struct {
		name     string
		buildFn  func() *ListNode
		wantLoop bool
	}{
		{
			"无环链表",
			func() *ListNode { return buildList([]int{1, 2, 3, 4}) },
			false,
		},
		{
			"有环链表",
			func() *ListNode {
				head := &ListNode{Val: 1}
				head.Next = &ListNode{Val: 2}
				head.Next.Next = &ListNode{Val: 3}
				head.Next.Next.Next = head.Next // 环：3 -> 2
				return head
			},
			true,
		},
		{
			"单节点无环",
			func() *ListNode { return &ListNode{Val: 1} },
			false,
		},
		{
			"单节点自环",
			func() *ListNode {
				node := &ListNode{Val: 1}
				node.Next = node
				return node
			},
			true,
		},
		{
			"空链表",
			func() *ListNode { return nil },
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := tt.buildFn()
			got := hasCycle(head)
			if got != tt.wantLoop {
				t.Errorf("hasCycle() = %v, want %v", got, tt.wantLoop)
			}
		})
	}
}
