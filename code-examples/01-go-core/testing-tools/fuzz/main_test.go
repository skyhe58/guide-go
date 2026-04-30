// Go 1.22+ | 验证日期：2025-01-01
// Fuzz Testing 示例
// 演示 Go 1.18+ 内置的模糊测试功能
// 通过自动生成随机输入来发现代码中的边界 bug
package fuzz

import (
	"testing"
	"unicode/utf8"
)

// ============================================================
// 被测函数
// ============================================================

// Reverse 反转字符串（支持 Unicode）
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Abs 返回整数的绝对值
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ============================================================
// Fuzz 测试
// ============================================================

// FuzzReverse 模糊测试字符串反转函数
// 验证属性：
//  1. 两次反转等于原始值
//  2. 反转后的 rune 数量不变
//  3. 反转后仍然是合法的 UTF-8 字符串
func FuzzReverse(f *testing.F) {
	// 添加种子语料（seed corpus）
	// 种子语料是 fuzz 引擎的起点，引擎会基于这些输入进行变异
	f.Add("hello")
	f.Add("世界")
	f.Add("")
	f.Add("a")
	f.Add("Go语言测试")
	f.Add("12345")

	// 注册模糊测试目标函数
	f.Fuzz(func(t *testing.T, s string) {
		// 跳过无效的 UTF-8 字符串
		if !utf8.ValidString(s) {
			t.Skip("跳过无效 UTF-8 字符串")
		}

		rev := Reverse(s)
		doubleRev := Reverse(rev)

		// 属性 1：两次反转应该等于原始值
		if s != doubleRev {
			t.Errorf("两次反转不等于原始值: %q -> %q -> %q", s, rev, doubleRev)
		}

		// 属性 2：反转后 rune 数量不变
		if utf8.RuneCountInString(rev) != utf8.RuneCountInString(s) {
			t.Errorf("rune 数量不匹配: 原始 %d, 反转后 %d",
				utf8.RuneCountInString(s), utf8.RuneCountInString(rev))
		}

		// 属性 3：反转后仍然是合法的 UTF-8 字符串
		if !utf8.ValidString(rev) {
			t.Errorf("反转后不是合法的 UTF-8: %q", rev)
		}
	})
}

// FuzzAbs 模糊测试绝对值函数
// 验证属性：
//  1. 结果始终 >= 0（注意 math.MinInt 的特殊情况）
//  2. Abs(n) == Abs(-n)
func FuzzAbs(f *testing.F) {
	// 添加种子语料
	f.Add(0)
	f.Add(1)
	f.Add(-1)
	f.Add(42)
	f.Add(-100)

	f.Fuzz(func(t *testing.T, n int) {
		result := Abs(n)

		// 属性 1：结果始终 >= 0
		// 注意：math.MinInt64 的绝对值会溢出，这里跳过
		if n == -1<<63 {
			t.Skip("跳过 MinInt64 溢出情况")
		}

		if result < 0 {
			t.Errorf("Abs(%d) = %d, 期望 >= 0", n, result)
		}

		// 属性 2：Abs(n) == Abs(-n)
		if result != Abs(-n) {
			t.Errorf("Abs(%d) = %d, Abs(%d) = %d, 期望相等", n, result, -n, Abs(-n))
		}
	})
}
