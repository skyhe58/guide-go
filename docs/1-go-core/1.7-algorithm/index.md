---
title: "数据结构与算法"
module: "algorithm"
difficulty: "intermediate"
tags:
  - 数据结构
  - 算法
  - 面试高频
  - LeetCode
---

# 数据结构与算法

> 📌 **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)。本模块优先覆盖 LeetCode Hot 100 高频题，所有算法均提供 Go 实现。

## 模块概述

本模块系统讲解面试高频数据结构与算法，所有题目均使用 Go 语言实现，配合详细中文注释和表驱动测试。内容按面试频率排序，优先覆盖 LeetCode Hot 100 中的经典题目。

Go 语言实现算法的优势：
- 内置切片（slice）天然适合动态数组操作
- `container/heap`、`container/list` 等标准库提供基础数据结构
- `sort` 包提供高效排序，源码值得深入学习
- 简洁的语法让算法逻辑更清晰

## 知识点索引（按面试频率排序）

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [链表](./01-linked-list.md) | ⭐⭐ | 🔥🔥🔥 | 45min |
| 02 | [栈与队列](./02-stack-queue.md) | ⭐⭐ | 🔥🔥🔥 | 40min |
| 03 | [哈希表](./03-hash-table.md) | ⭐⭐ | 🔥🔥🔥 | 45min |
| 04 | [树与二叉树](./04-tree.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 05 | [堆](./05-heap.md) | ⭐⭐⭐ | 🔥🔥 | 40min |
| 06 | [图](./06-graph.md) | ⭐⭐⭐ | 🔥🔥 | 40min |
| 07 | [排序算法](./07-sorting.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 08 | [二分查找](./08-binary-search.md) | ⭐⭐ | 🔥🔥🔥 | 40min |
| 09 | [双指针与滑动窗口](./09-two-pointers.md) | ⭐⭐ | 🔥🔥🔥 | 45min |
| 10 | [动态规划](./10-dp.md) | ⭐⭐⭐ | 🔥🔥🔥 | 90min |
| 11 | [回溯算法](./11-backtracking.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 12 | [贪心算法](./12-greedy.md) | ⭐⭐ | 🔥🔥 | 40min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/algorithm/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/algorithm/)

| 示例目录 | 对应知识点 | 运行方式 |
|---------|-----------|---------|
| `linkedlist/` | 反转链表/合并有序链表/环形链表 | `go run main.go` |
| `stack-queue/` | 有效括号/最小栈/用栈实现队列 | `go run main.go` |
| `hashtable/` | 两数之和/字母异位词/LRU 缓存 | `go run main.go` |
| `tree/` | 二叉树遍历/BST/最近公共祖先 | `go run main.go` |
| `sorting/` | 快排/归并/堆排序 | `go run main.go` |
| `dp/` | 爬楼梯/LIS/背包/编辑距离 | `go run main.go` |

## 学习建议

1. **先掌握基础数据结构**：链表、栈、队列、哈希表、树
2. **再学习算法思想**：双指针、二分查找、排序
3. **最后攻克高级算法**：动态规划、回溯、贪心
4. **每道题先手写再看答案**，理解思路比背代码更重要
5. **刷题顺序建议**：按本模块的面试频率排序，从高到低

## 参考资料

- [LeetCode Hot 100](https://leetcode.cn/studyplan/top-100-liked/)
- [Go 标准库 sort 包](https://pkg.go.dev/sort)
- [Go 标准库 container 包](https://pkg.go.dev/container)
