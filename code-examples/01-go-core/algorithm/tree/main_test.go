// 二叉树算法 — 表驱动测试
// Go 1.22+ | 验证日期：2025-01-01
package main

import (
	"reflect"
	"testing"
)

// buildTree 辅助函数：从层序数组构建二叉树（-1 表示 nil）
func buildTree(vals []int) *TreeNode {
	if len(vals) == 0 || vals[0] == -1 {
		return nil
	}
	root := &TreeNode{Val: vals[0]}
	queue := []*TreeNode{root}
	i := 1
	for len(queue) > 0 && i < len(vals) {
		node := queue[0]
		queue = queue[1:]
		if i < len(vals) && vals[i] != -1 {
			node.Left = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Left)
		}
		i++
		if i < len(vals) && vals[i] != -1 {
			node.Right = &TreeNode{Val: vals[i]}
			queue = append(queue, node.Right)
		}
		i++
	}
	return root
}

// TestInorderTraversal 中序遍历测试
func TestInorderTraversal(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want []int
	}{
		{"正常树", []int{1, 2, 3, 4, 5}, []int{4, 2, 5, 1, 3}},
		{"只有根", []int{1}, []int{1}},
		{"空树", nil, nil},
		{"左偏树", []int{3, 2, -1, 1}, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildTree(tt.vals)
			got := inorderTraversal(root)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inorderTraversal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPreorderTraversal 前序遍历测试
func TestPreorderTraversal(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want []int
	}{
		{"正常树", []int{1, 2, 3, 4, 5}, []int{1, 2, 4, 5, 3}},
		{"只有根", []int{1}, []int{1}},
		{"空树", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildTree(tt.vals)
			got := preorderTraversal(root)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("preorderTraversal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLevelOrder 层序遍历测试
func TestLevelOrder(t *testing.T) {
	tests := []struct {
		name string
		vals []int
		want [][]int
	}{
		{"正常树", []int{1, 2, 3, 4, 5}, [][]int{{1}, {2, 3}, {4, 5}}},
		{"只有根", []int{1}, [][]int{{1}}},
		{"空树", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildTree(tt.vals)
			got := levelOrder(root)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("levelOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsValidBST 验证 BST 测试
func TestIsValidBST(t *testing.T) {
	tests := []struct {
		name    string
		buildFn func() *TreeNode
		want    bool
	}{
		{
			"有效 BST",
			func() *TreeNode {
				root := &TreeNode{Val: 2}
				root.Left = &TreeNode{Val: 1}
				root.Right = &TreeNode{Val: 3}
				return root
			},
			true,
		},
		{
			"无效 BST",
			func() *TreeNode {
				root := &TreeNode{Val: 5}
				root.Left = &TreeNode{Val: 1}
				root.Right = &TreeNode{Val: 4}
				root.Right.Left = &TreeNode{Val: 3}
				root.Right.Right = &TreeNode{Val: 6}
				return root
			},
			false,
		},
		{
			"单节点",
			func() *TreeNode { return &TreeNode{Val: 1} },
			true,
		},
		{
			"空树",
			func() *TreeNode { return nil },
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.buildFn()
			got := isValidBST(root)
			if got != tt.want {
				t.Errorf("isValidBST() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLowestCommonAncestor 最近公共祖先测试
func TestLowestCommonAncestor(t *testing.T) {
	// 构建树：
	//       3
	//      / \
	//     5   1
	//    / \
	//   6   2
	root := &TreeNode{Val: 3}
	root.Left = &TreeNode{Val: 5}
	root.Right = &TreeNode{Val: 1}
	root.Left.Left = &TreeNode{Val: 6}
	root.Left.Right = &TreeNode{Val: 2}

	tests := []struct {
		name string
		p, q *TreeNode
		want int
	}{
		{"同一子树", root.Left, root.Left.Right, 5},
		{"不同子树", root.Left, root.Right, 3},
		{"p 是 q 的祖先", root.Left, root.Left.Left, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lowestCommonAncestor(root, tt.p, tt.q)
			if got == nil || got.Val != tt.want {
				gotVal := -1
				if got != nil {
					gotVal = got.Val
				}
				t.Errorf("lowestCommonAncestor() = %d, want %d", gotVal, tt.want)
			}
		})
	}
}
