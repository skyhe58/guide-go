// 二叉树遍历与 BST — 面试高频树算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：前中后序遍历、层序遍历、验证 BST、最近公共祖先
// 运行方式：go run main.go
package main

import "fmt"

// TreeNode 二叉树节点定义
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// inorderTraversal 中序遍历（递归）— LeetCode 94
// 顺序：左 → 根 → 右
// 时间复杂度：O(n)，空间复杂度：O(h)，h 为树高
func inorderTraversal(root *TreeNode) []int {
	var result []int
	var inorder func(node *TreeNode)
	inorder = func(node *TreeNode) {
		if node == nil {
			return
		}
		inorder(node.Left)                    // 遍历左子树
		result = append(result, node.Val)     // 访问根节点
		inorder(node.Right)                   // 遍历右子树
	}
	inorder(root)
	return result
}

// preorderTraversal 前序遍历（递归）— LeetCode 144
// 顺序：根 → 左 → 右
func preorderTraversal(root *TreeNode) []int {
	var result []int
	var preorder func(node *TreeNode)
	preorder = func(node *TreeNode) {
		if node == nil {
			return
		}
		result = append(result, node.Val) // 先访问根节点
		preorder(node.Left)
		preorder(node.Right)
	}
	preorder(root)
	return result
}

// postorderTraversal 后序遍历（递归）— LeetCode 145
// 顺序：左 → 右 → 根
func postorderTraversal(root *TreeNode) []int {
	var result []int
	var postorder func(node *TreeNode)
	postorder = func(node *TreeNode) {
		if node == nil {
			return
		}
		postorder(node.Left)
		postorder(node.Right)
		result = append(result, node.Val) // 最后访问根节点
	}
	postorder(root)
	return result
}

// levelOrder 层序遍历（BFS）— LeetCode 102
// 思路：使用队列，每次处理一层的所有节点
// 时间复杂度：O(n)，空间复杂度：O(n)
func levelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var result [][]int
	queue := []*TreeNode{root}
	for len(queue) > 0 {
		size := len(queue) // 当前层的节点数
		var level []int
		for i := 0; i < size; i++ {
			node := queue[0]
			queue = queue[1:]
			level = append(level, node.Val)
			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}
		result = append(result, level)
	}
	return result
}

// isValidBST 验证二叉搜索树（LeetCode 98）
// 思路：BST 的中序遍历结果是严格递增的
// 时间复杂度：O(n)，空间复杂度：O(h)
func isValidBST(root *TreeNode) bool {
	var prev *int
	var validate func(node *TreeNode) bool
	validate = func(node *TreeNode) bool {
		if node == nil {
			return true
		}
		// 先验证左子树
		if !validate(node.Left) {
			return false
		}
		// 检查当前节点是否大于前一个节点
		if prev != nil && node.Val <= *prev {
			return false
		}
		prev = &node.Val
		// 再验证右子树
		return validate(node.Right)
	}
	return validate(root)
}

// lowestCommonAncestor 最近公共祖先（LeetCode 236）
// 思路：递归查找，如果当前节点是 p 或 q，或者左右子树各找到一个，则当前节点是 LCA
// 时间复杂度：O(n)，空间复杂度：O(h)
func lowestCommonAncestor(root, p, q *TreeNode) *TreeNode {
	if root == nil || root == p || root == q {
		return root
	}
	left := lowestCommonAncestor(root.Left, p, q)
	right := lowestCommonAncestor(root.Right, p, q)
	if left != nil && right != nil {
		return root // p 和 q 分别在左右子树，当前节点就是 LCA
	}
	if left != nil {
		return left
	}
	return right
}

func main() {
	fmt.Println("=== 二叉树算法示例 ===")

	// 构建测试树：
	//       1
	//      / \
	//     2   3
	//    / \
	//   4   5
	root := &TreeNode{Val: 1}
	root.Left = &TreeNode{Val: 2}
	root.Right = &TreeNode{Val: 3}
	root.Left.Left = &TreeNode{Val: 4}
	root.Left.Right = &TreeNode{Val: 5}

	// 1. 三种遍历
	fmt.Println("\n--- 二叉树遍历 ---")
	fmt.Println("  前序遍历:", preorderTraversal(root))  // [1,2,4,5,3]
	fmt.Println("  中序遍历:", inorderTraversal(root))   // [4,2,5,1,3]
	fmt.Println("  后序遍历:", postorderTraversal(root)) // [4,5,2,3,1]

	// 2. 层序遍历
	fmt.Println("\n--- 层序遍历 ---")
	fmt.Println("  层序:", levelOrder(root)) // [[1],[2,3],[4,5]]

	// 3. 验证 BST
	fmt.Println("\n--- 验证 BST ---")
	bst := &TreeNode{Val: 2}
	bst.Left = &TreeNode{Val: 1}
	bst.Right = &TreeNode{Val: 3}
	fmt.Println("  [2,1,3] 是 BST:", isValidBST(bst)) // true

	notBST := &TreeNode{Val: 5}
	notBST.Left = &TreeNode{Val: 1}
	notBST.Right = &TreeNode{Val: 4}
	notBST.Right.Left = &TreeNode{Val: 3}
	notBST.Right.Right = &TreeNode{Val: 6}
	fmt.Println("  [5,1,4,nil,nil,3,6] 是 BST:", isValidBST(notBST)) // false

	// 4. 最近公共祖先
	fmt.Println("\n--- 最近公共祖先 ---")
	p := root.Left       // 节点 2
	q := root.Left.Right // 节点 5
	lca := lowestCommonAncestor(root, p, q)
	fmt.Printf("  节点 %d 和 %d 的 LCA: %d\n", p.Val, q.Val, lca.Val) // 2
}
