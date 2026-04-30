// 排序算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import (
	"reflect"
	"sort"
	"testing"
)

// TestQuickSort 快速排序测试
func TestQuickSort(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"正常数组", []int{3, 6, 8, 10, 1, 2, 1}, []int{1, 1, 2, 3, 6, 8, 10}},
		{"已排序", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{"逆序", []int{5, 4, 3, 2, 1}, []int{1, 2, 3, 4, 5}},
		{"有重复", []int{3, 3, 3, 1, 1}, []int{1, 1, 3, 3, 3}},
		{"单个元素", []int{1}, []int{1}},
		{"两个元素", []int{2, 1}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := make([]int, len(tt.input))
			copy(nums, tt.input)
			quickSort(nums, 0, len(nums)-1)
			if !reflect.DeepEqual(nums, tt.want) {
				t.Errorf("quickSort(%v) = %v, want %v", tt.input, nums, tt.want)
			}
		})
	}
}

// TestMergeSort 归并排序测试
func TestMergeSort(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"正常数组", []int{38, 27, 43, 3, 9, 82, 10}, []int{3, 9, 10, 27, 38, 43, 82}},
		{"已排序", []int{1, 2, 3}, []int{1, 2, 3}},
		{"逆序", []int{3, 2, 1}, []int{1, 2, 3}},
		{"单个元素", []int{1}, []int{1}},
		{"空数组", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeSort(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSort(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestHeapSort 堆排序测试
func TestHeapSort(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"正常数组", []int{12, 11, 13, 5, 6, 7}, []int{5, 6, 7, 11, 12, 13}},
		{"已排序", []int{1, 2, 3, 4}, []int{1, 2, 3, 4}},
		{"逆序", []int{4, 3, 2, 1}, []int{1, 2, 3, 4}},
		{"有重复", []int{5, 3, 5, 1, 3}, []int{1, 3, 3, 5, 5}},
		{"单个元素", []int{1}, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := make([]int, len(tt.input))
			copy(nums, tt.input)
			heapSort(nums)
			if !reflect.DeepEqual(nums, tt.want) {
				t.Errorf("heapSort(%v) = %v, want %v", tt.input, nums, tt.want)
			}
		})
	}
}

// TestSortingConsistency 验证三种排序算法结果一致性
func TestSortingConsistency(t *testing.T) {
	input := []int{9, 3, 7, 1, 5, 8, 2, 6, 4, 10}

	// 标准库排序作为参考
	expected := make([]int, len(input))
	copy(expected, input)
	sort.Ints(expected)

	// 快速排序
	arr1 := make([]int, len(input))
	copy(arr1, input)
	quickSort(arr1, 0, len(arr1)-1)
	if !reflect.DeepEqual(arr1, expected) {
		t.Errorf("quickSort 结果不一致: got %v, want %v", arr1, expected)
	}

	// 归并排序
	arr2 := mergeSort(input)
	if !reflect.DeepEqual(arr2, expected) {
		t.Errorf("mergeSort 结果不一致: got %v, want %v", arr2, expected)
	}

	// 堆排序
	arr3 := make([]int, len(input))
	copy(arr3, input)
	heapSort(arr3)
	if !reflect.DeepEqual(arr3, expected) {
		t.Errorf("heapSort 结果不一致: got %v, want %v", arr3, expected)
	}
}
