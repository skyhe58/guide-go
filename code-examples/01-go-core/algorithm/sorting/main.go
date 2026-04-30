// 排序算法 — 面试高频排序算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：快速排序、归并排序、堆排序
// 运行方式：go run main.go
package main

import "fmt"

// quickSort 快速排序
// 思路：选择 pivot，将数组分为小于和大于 pivot 的两部分，递归排序
// 时间复杂度：平均 O(n log n)，最坏 O(n²)
// 空间复杂度：O(log n)（递归栈）
func quickSort(nums []int, left, right int) {
	if left >= right {
		return
	}
	pivotIdx := partition(nums, left, right)
	quickSort(nums, left, pivotIdx-1)  // 排序左半部分
	quickSort(nums, pivotIdx+1, right) // 排序右半部分
}

// partition 分区函数：选最后一个元素作为 pivot
// 将小于 pivot 的元素放左边，大于的放右边
func partition(nums []int, left, right int) int {
	pivot := nums[right]
	i := left // i 指向小于 pivot 区域的右边界
	for j := left; j < right; j++ {
		if nums[j] < pivot {
			nums[i], nums[j] = nums[j], nums[i]
			i++
		}
	}
	nums[i], nums[right] = nums[right], nums[i] // pivot 归位
	return i
}

// mergeSort 归并排序
// 思路：分治法，将数组一分为二，分别排序后合并
// 时间复杂度：O(n log n)，空间复杂度：O(n)
func mergeSort(nums []int) []int {
	if len(nums) <= 1 {
		return nums
	}
	mid := len(nums) / 2
	left := mergeSort(append([]int{}, nums[:mid]...))   // 拷贝左半部分
	right := mergeSort(append([]int{}, nums[mid:]...))   // 拷贝右半部分
	return mergeSorted(left, right)
}

// mergeSorted 合并两个有序数组
func mergeSorted(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)  // 拼接左边剩余
	result = append(result, right[j:]...) // 拼接右边剩余
	return result
}

// heapSort 堆排序
// 思路：建立最大堆，依次将堆顶（最大值）与末尾交换，缩小堆范围
// 时间复杂度：O(n log n)，空间复杂度：O(1)
func heapSort(nums []int) {
	n := len(nums)
	// 建堆：从最后一个非叶子节点开始下沉
	for i := n/2 - 1; i >= 0; i-- {
		siftDown(nums, i, n)
	}
	// 排序：依次将堆顶（最大值）与末尾交换
	for i := n - 1; i > 0; i-- {
		nums[0], nums[i] = nums[i], nums[0] // 堆顶与末尾交换
		siftDown(nums, 0, i)                 // 对剩余元素重新调整堆
	}
}

// siftDown 下沉操作：维护最大堆性质
func siftDown(nums []int, i, n int) {
	for {
		largest := i
		left, right := 2*i+1, 2*i+2
		if left < n && nums[left] > nums[largest] {
			largest = left
		}
		if right < n && nums[right] > nums[largest] {
			largest = right
		}
		if largest == i {
			break // 已满足堆性质
		}
		nums[i], nums[largest] = nums[largest], nums[i]
		i = largest // 继续下沉
	}
}

func main() {
	fmt.Println("=== 排序算法示例 ===")

	// 1. 快速排序
	fmt.Println("\n--- 快速排序 ---")
	arr1 := []int{3, 6, 8, 10, 1, 2, 1}
	fmt.Println("  排序前:", arr1)
	quickSort(arr1, 0, len(arr1)-1)
	fmt.Println("  排序后:", arr1)

	// 2. 归并排序
	fmt.Println("\n--- 归并排序 ---")
	arr2 := []int{38, 27, 43, 3, 9, 82, 10}
	fmt.Println("  排序前:", arr2)
	sorted := mergeSort(arr2)
	fmt.Println("  排序后:", sorted)

	// 3. 堆排序
	fmt.Println("\n--- 堆排序 ---")
	arr3 := []int{12, 11, 13, 5, 6, 7}
	fmt.Println("  排序前:", arr3)
	heapSort(arr3)
	fmt.Println("  排序后:", arr3)
}
