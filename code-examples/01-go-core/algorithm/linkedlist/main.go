// 链表操作 — 面试高频链表算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：反转链表、合并两个有序链表、环形链表检测
// 运行方式：go run main.go
package main

import "fmt"

// ListNode 单链表节点定义
type ListNode struct {
	Val  int
	Next *ListNode
}

// buildList 从切片构建链表（辅助函数）
func buildList(vals []int) *ListNode {
	dummy := &ListNode{}
	curr := dummy
	for _, v := range vals {
		curr.Next = &ListNode{Val: v}
		curr = curr.Next
	}
	return dummy.Next
}

// listToSlice 将链表转换为切片（辅助函数）
func listToSlice(head *ListNode) []int {
	var result []int
	for head != nil {
		result = append(result, head.Val)
		head = head.Next
	}
	return result
}

// reverseList 反转链表（LeetCode 206）
// 思路：用三个指针 prev、curr、next 逐个翻转指针方向
// 时间复杂度：O(n)，空间复杂度：O(1)
func reverseList(head *ListNode) *ListNode {
	var prev *ListNode // 前驱节点，初始为 nil
	curr := head       // 当前节点
	for curr != nil {
		next := curr.Next // 暂存下一个节点，防止链表断裂
		curr.Next = prev  // 翻转指针方向
		prev = curr       // prev 前进一步
		curr = next       // curr 前进一步
	}
	return prev // prev 就是新的头节点
}

// mergeTwoLists 合并两个有序链表（LeetCode 21）
// 思路：使用哨兵节点简化边界处理，逐个比较取较小值
// 时间复杂度：O(m+n)，空间复杂度：O(1)
func mergeTwoLists(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{} // 哨兵节点，简化头节点处理
	curr := dummy
	for l1 != nil && l2 != nil {
		if l1.Val <= l2.Val {
			curr.Next = l1
			l1 = l1.Next
		} else {
			curr.Next = l2
			l2 = l2.Next
		}
		curr = curr.Next
	}
	// 拼接剩余部分（只有一个链表还有剩余）
	if l1 != nil {
		curr.Next = l1
	} else {
		curr.Next = l2
	}
	return dummy.Next
}

// hasCycle 环形链表检测（LeetCode 141）
// 思路：快慢指针，快指针每次走两步，慢指针每次走一步
// 如果有环，快慢指针必定在环内相遇
// 时间复杂度：O(n)，空间复杂度：O(1)
func hasCycle(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next      // 慢指针走一步
		fast = fast.Next.Next // 快指针走两步
		if slow == fast {
			return true // 相遇说明有环
		}
	}
	return false // 快指针到达末尾，说明无环
}

func main() {
	fmt.Println("=== 链表算法示例 ===")

	// 1. 反转链表
	fmt.Println("\n--- 反转链表 ---")
	list1 := buildList([]int{1, 2, 3, 4, 5})
	fmt.Println("原链表:", listToSlice(list1))
	reversed := reverseList(list1)
	fmt.Println("反转后:", listToSlice(reversed))

	// 2. 合并两个有序链表
	fmt.Println("\n--- 合并两个有序链表 ---")
	l1 := buildList([]int{1, 3, 5})
	l2 := buildList([]int{2, 4, 6})
	fmt.Println("链表1:", listToSlice(l1))
	fmt.Println("链表2:", listToSlice(l2))
	merged := mergeTwoLists(l1, l2)
	fmt.Println("合并后:", listToSlice(merged))

	// 3. 环形链表检测
	fmt.Println("\n--- 环形链表检测 ---")
	noLoop := buildList([]int{1, 2, 3})
	fmt.Println("无环链表:", hasCycle(noLoop))

	// 构造有环链表：3 -> 4 -> 5 -> 3（环）
	loopHead := &ListNode{Val: 3}
	loopHead.Next = &ListNode{Val: 4}
	loopHead.Next.Next = &ListNode{Val: 5}
	loopHead.Next.Next.Next = loopHead // 形成环
	fmt.Println("有环链表:", hasCycle(loopHead))
}
