// 哈希表 — 面试高频哈希表算法 Go 实现
// Go 1.22+ | 验证日期：2025-01-01
//
// 包含：两数之和、字母异位词分组、LRU 缓存
// 运行方式：go run main.go
package main

import (
	"fmt"
	"sort"
)

// twoSum 两数之和（LeetCode 1）
// 思路：用哈希表存储已遍历的数及其索引，查找 target-nums[i] 是否存在
// 时间复杂度：O(n)，空间复杂度：O(n)
func twoSum(nums []int, target int) []int {
	m := make(map[int]int) // 值 -> 索引
	for i, num := range nums {
		if j, ok := m[target-num]; ok {
			return []int{j, i} // 找到配对，返回两个索引
		}
		m[num] = i // 记录当前值和索引
	}
	return nil // 无解
}

// groupAnagrams 字母异位词分组（LeetCode 49）
// 思路：将每个单词排序后作为 key，相同 key 的单词归为一组
// 时间复杂度：O(n * k * log(k))，k 为单词最大长度
func groupAnagrams(strs []string) [][]string {
	groups := make(map[string][]string)
	for _, s := range strs {
		// 将字符串的字符排序后作为分组 key
		bs := []byte(s)
		sort.Slice(bs, func(i, j int) bool { return bs[i] < bs[j] })
		key := string(bs)
		groups[key] = append(groups[key], s)
	}
	result := make([][]string, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}
	return result
}

// LRUNode LRU 缓存的双向链表节点
type LRUNode struct {
	key, value int
	prev, next *LRUNode
}

// LRUCache LRU 缓存实现（LeetCode 146）
// 思路：哈希表 + 双向链表
// 哈希表实现 O(1) 查找，双向链表维护访问顺序
type LRUCache struct {
	capacity   int
	cache      map[int]*LRUNode
	head, tail *LRUNode // 哨兵节点
}

// NewLRUCache 创建 LRU 缓存
func NewLRUCache(capacity int) *LRUCache {
	head := &LRUNode{}
	tail := &LRUNode{}
	head.next = tail
	tail.prev = head
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*LRUNode),
		head:     head,
		tail:     tail,
	}
}

// Get 获取缓存值，访问后移到头部（最近使用）
func (c *LRUCache) Get(key int) int {
	if node, ok := c.cache[key]; ok {
		c.moveToHead(node)
		return node.value
	}
	return -1 // 未找到
}

// Put 写入缓存，超容量时淘汰最久未使用的
func (c *LRUCache) Put(key, value int) {
	if node, ok := c.cache[key]; ok {
		node.value = value
		c.moveToHead(node)
		return
	}
	// 新建节点
	node := &LRUNode{key: key, value: value}
	c.cache[key] = node
	c.addToHead(node)
	// 超容量，移除尾部（最久未使用）
	if len(c.cache) > c.capacity {
		removed := c.removeTail()
		delete(c.cache, removed.key)
	}
}

// addToHead 将节点添加到头部（最近使用）
func (c *LRUCache) addToHead(node *LRUNode) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

// removeNode 从链表中移除节点
func (c *LRUCache) removeNode(node *LRUNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// moveToHead 将节点移到头部
func (c *LRUCache) moveToHead(node *LRUNode) {
	c.removeNode(node)
	c.addToHead(node)
}

// removeTail 移除尾部节点（最久未使用）
func (c *LRUCache) removeTail() *LRUNode {
	node := c.tail.prev
	c.removeNode(node)
	return node
}

func main() {
	fmt.Println("=== 哈希表算法示例 ===")

	// 1. 两数之和
	fmt.Println("\n--- 两数之和 ---")
	nums := []int{2, 7, 11, 15}
	target := 9
	fmt.Printf("  nums=%v, target=%d -> %v\n", nums, target, twoSum(nums, target))

	// 2. 字母异位词分组
	fmt.Println("\n--- 字母异位词分组 ---")
	strs := []string{"eat", "tea", "tan", "ate", "nat", "bat"}
	groups := groupAnagrams(strs)
	fmt.Printf("  输入: %v\n", strs)
	for _, g := range groups {
		fmt.Printf("  分组: %v\n", g)
	}

	// 3. LRU 缓存
	fmt.Println("\n--- LRU 缓存 ---")
	lru := NewLRUCache(2)
	lru.Put(1, 1)
	lru.Put(2, 2)
	fmt.Println("  Get(1):", lru.Get(1)) // 1
	lru.Put(3, 3)                         // 淘汰 key=2
	fmt.Println("  Get(2):", lru.Get(2)) // -1（已淘汰）
	lru.Put(4, 4)                         // 淘汰 key=1
	fmt.Println("  Get(1):", lru.Get(1)) // -1（已淘汰）
	fmt.Println("  Get(3):", lru.Get(3)) // 3
	fmt.Println("  Get(4):", lru.Get(4)) // 4
}
