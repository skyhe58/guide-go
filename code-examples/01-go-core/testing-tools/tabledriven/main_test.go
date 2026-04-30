// Go 1.22+ | 验证日期：2025-01-01
// 表驱动测试完整示例
// 演示 Go 社区最推崇的测试模式：表驱动测试（Table-Driven Tests）
// 包含：基本表驱动、子测试 t.Run、t.Parallel、t.Helper、错误场景测试
package tabledriven

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// ============================================================
// 被测函数
// ============================================================

// Add 两数相加
func Add(a, b int) int {
	return a + b
}

// Divide 除法运算，除数为零时返回错误
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零")
	}
	return a / b, nil
}

// Reverse 反转字符串（支持 Unicode）
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Capitalize 将字符串首字母大写
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToUpper(string(r)) + s[size:]
}

// IsPalindrome 判断字符串是否为回文
func IsPalindrome(s string) bool {
	s = strings.ToLower(s)
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

// ============================================================
// 辅助函数（演示 t.Helper）
// ============================================================

// assertEqual 通用断言辅助函数
// t.Helper() 使错误报告指向调用者而非此函数内部
func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// assertError 断言是否返回错误
func assertError(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}
}

// ============================================================
// 测试用例
// ============================================================

// TestAdd 基本表驱动测试示例
// 将测试用例组织为结构体切片，通过循环遍历执行
func TestAdd(t *testing.T) {
	tests := []struct {
		name string // 用例名称，失败时精确定位
		a, b int    // 输入参数
		want int    // 期望输出
	}{
		{"正数相加", 1, 2, 3},
		{"负数相加", -1, -2, -3},
		{"零值相加", 0, 0, 0},
		{"正负混合", 1, -1, 0},
		{"大数相加", 1000000, 2000000, 3000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			assertEqual(t, got, tt.want)
		})
	}
}

// TestDivide 带错误返回值的表驱动测试
// 演示如何测试返回 error 的函数
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool // 是否期望返回错误
	}{
		{"正常除法", 10, 2, 5, false},
		{"小数结果", 10, 3, 3.3333333333333335, false},
		{"除以零", 10, 0, 0, true},
		{"零除以数", 0, 5, 0, false},
		{"负数除法", -10, 2, -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			assertError(t, err, tt.wantErr)
			if !tt.wantErr {
				assertEqual(t, got, tt.want)
			}
		})
	}
}

// TestReverse 字符串反转测试
// 演示 Unicode 字符串的测试
func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"英文字符串", "hello", "olleh"},
		{"中文字符串", "你好世界", "界世好你"},
		{"空字符串", "", ""},
		{"单字符", "a", "a"},
		{"回文字符串", "aba", "aba"},
		{"混合字符", "Go语言", "言语oG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reverse(tt.input)
			assertEqual(t, got, tt.want)
		})
	}
}

// TestCapitalize 首字母大写测试
func TestCapitalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"普通单词", "hello", "Hello"},
		{"已大写", "Hello", "Hello"},
		{"空字符串", "", ""},
		{"单字符", "a", "A"},
		{"数字开头", "123abc", "123abc"},
		{"中文开头", "你好world", "你好world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Capitalize(tt.input)
			assertEqual(t, got, tt.want)
		})
	}
}

// TestIsPalindrome 回文判断测试
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"回文-奇数长度", "aba", true},
		{"回文-偶数长度", "abba", true},
		{"非回文", "abc", false},
		{"空字符串", "", true},
		{"单字符", "a", true},
		{"大小写混合回文", "Aba", true},
		{"中文回文", "上海自来水来自海上", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPalindrome(tt.input)
			assertEqual(t, got, tt.want)
		})
	}
}

// TestAddParallel 并行测试示例
// 演示 t.Parallel() 加速测试执行
func TestAddParallel(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"case1", 1, 2, 3},
		{"case2", 10, 20, 30},
		{"case3", -5, 5, 0},
		{"case4", 100, 200, 300},
	}

	for _, tt := range tests {
		// Go 1.22+ for-range 变量语义已修复
		// 不再需要 tt := tt 的 shadow 技巧
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 标记为可并行执行
			got := Add(tt.a, tt.b)
			assertEqual(t, got, tt.want)
		})
	}
}
