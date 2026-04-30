// 动态规划 — 面试高频 DP 算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：爬楼梯、最长递增子序列、0-1 背包、编辑距离
// 运行方式：go run main.go
package main

import "fmt"

// climbStairs 爬楼梯（LeetCode 70）
// 状态：dp[i] = 到达第 i 阶的方法数
// 转移：dp[i] = dp[i-1] + dp[i-2]
// 时间复杂度：O(n)，空间复杂度：O(1)（滚动变量优化）
func climbStairs(n int) int {
	if n <= 2 {
		return n
	}
	prev, curr := 1, 2 // dp[1]=1, dp[2]=2
	for i := 3; i <= n; i++ {
		prev, curr = curr, prev+curr // 滚动更新
	}
	return curr
}

// lengthOfLIS 最长递增子序列（LeetCode 300）
// 状态：dp[i] = 以 nums[i] 结尾的最长递增子序列长度
// 转移：dp[i] = max(dp[j]+1)，其中 j < i 且 nums[j] < nums[i]
// 时间复杂度：O(n²)，空间复杂度：O(n)
func lengthOfLIS(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	dp := make([]int, n)
	for i := range dp {
		dp[i] = 1 // 每个元素自身构成长度为 1 的子序列
	}
	maxLen := 1
	for i := 1; i < n; i++ {
		for j := 0; j < i; j++ {
			if nums[j] < nums[i] && dp[j]+1 > dp[i] {
				dp[i] = dp[j] + 1
			}
		}
		if dp[i] > maxLen {
			maxLen = dp[i]
		}
	}
	return maxLen
}

// knapsack 0-1 背包问题
// 状态：dp[w] = 容量为 w 时的最大价值
// 转移：dp[w] = max(dp[w], dp[w-weight[i]] + value[i])
// 空间优化：一维数组，逆序遍历容量（确保每个物品只选一次）
// 时间复杂度：O(n*W)，空间复杂度：O(W)
func knapsack(weights, values []int, capacity int) int {
	dp := make([]int, capacity+1)
	for i := 0; i < len(weights); i++ {
		// 逆序遍历容量，确保每个物品只被选择一次
		for w := capacity; w >= weights[i]; w-- {
			val := dp[w-weights[i]] + values[i]
			if val > dp[w] {
				dp[w] = val
			}
		}
	}
	return dp[capacity]
}

// minDistance 编辑距离（LeetCode 72）
// 状态：dp[i][j] = word1[:i] 转换为 word2[:j] 的最少操作数
// 转移：
//
//	word1[i-1] == word2[j-1]: dp[i][j] = dp[i-1][j-1]（字符相同，无需操作）
//	否则: dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+1)
//	                      删除           插入           替换
//
// 时间复杂度：O(m*n)，空间复杂度：O(m*n)
func minDistance(word1, word2 string) int {
	m, n := len(word1), len(word2)
	// 初始化 DP 表
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		dp[i][0] = i // word1[:i] 转换为空串需要 i 次删除
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j // 空串转换为 word2[:j] 需要 j 次插入
	}
	// 填表
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if word1[i-1] == word2[j-1] {
				dp[i][j] = dp[i-1][j-1] // 字符相同，无需操作
			} else {
				dp[i][j] = minThree(
					dp[i-1][j]+1,   // 删除 word1[i-1]
					dp[i][j-1]+1,   // 插入 word2[j-1]
					dp[i-1][j-1]+1, // 替换 word1[i-1] 为 word2[j-1]
				)
			}
		}
	}
	return dp[m][n]
}

// minThree 返回三个数中的最小值
func minThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func main() {
	fmt.Println("=== 动态规划算法示例 ===")

	// 1. 爬楼梯
	fmt.Println("\n--- 爬楼梯 ---")
	for _, n := range []int{2, 3, 5, 10} {
		fmt.Printf("  %d 阶楼梯有 %d 种走法\n", n, climbStairs(n))
	}

	// 2. 最长递增子序列
	fmt.Println("\n--- 最长递增子序列 ---")
	nums := []int{10, 9, 2, 5, 3, 7, 101, 18}
	fmt.Printf("  nums=%v -> LIS 长度=%d\n", nums, lengthOfLIS(nums))

	// 3. 0-1 背包
	fmt.Println("\n--- 0-1 背包 ---")
	weights := []int{2, 3, 4, 5}
	values := []int{3, 4, 5, 6}
	capacity := 8
	fmt.Printf("  weights=%v, values=%v, capacity=%d\n", weights, values, capacity)
	fmt.Printf("  最大价值=%d\n", knapsack(weights, values, capacity))

	// 4. 编辑距离
	fmt.Println("\n--- 编辑距离 ---")
	pairs := [][2]string{
		{"horse", "ros"},
		{"intention", "execution"},
	}
	for _, p := range pairs {
		fmt.Printf("  \"%s\" -> \"%s\" 编辑距离=%d\n", p[0], p[1], minDistance(p[0], p[1]))
	}
}
