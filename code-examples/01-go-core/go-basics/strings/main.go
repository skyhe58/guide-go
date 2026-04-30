// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 字符串处理示例：strings 包、strconv 包、rune vs byte、strings.Builder
package main

import (
	"fmt"
	"strings"
	"strconv"
	"unicode/utf8"
)

func main() {
	fmt.Println("========== Go 字符串处理示例 ==========")

	// ========== 1. 字符串基础 ==========
	fmt.Println("\n--- 1. 字符串基础 ---")

	s := "Hello, 世界！"
	fmt.Printf("字符串: %q\n", s)
	fmt.Printf("字节数 len(): %d\n", len(s))
	fmt.Printf("字符数 RuneCount: %d\n", utf8.RuneCountInString(s))

	// 字符串是不可变的
	// s[0] = 'h' // 编译错误！
	fmt.Println("字符串不可变，修改需转换为 []byte 或 []rune")

	// ========== 2. rune vs byte ==========
	fmt.Println("\n--- 2. rune vs byte ---")

	str := "Go语言"
	fmt.Printf("字符串: %q\n", str)

	// byte 遍历（按字节）
	fmt.Println("byte 遍历:")
	for i := 0; i < len(str); i++ {
		fmt.Printf("  [%d] 0x%02x\n", i, str[i])
	}

	// rune 遍历（按字符，推荐）
	fmt.Println("rune 遍历 (for-range):")
	for i, r := range str {
		fmt.Printf("  字节偏移=%d, 字符=%c, Unicode=U+%04X, 大小=%d字节\n",
			i, r, r, utf8.RuneLen(r))
	}

	// []byte vs []rune 转换
	fmt.Println("\n[]byte vs []rune:")
	bytes := []byte(str)
	runes := []rune(str)
	fmt.Printf("  []byte: %v (长度=%d)\n", bytes, len(bytes))
	fmt.Printf("  []rune: %v (长度=%d)\n", runes, len(runes))

	// 修改字符串中的字符
	runeSlice := []rune("Hello, 世界")
	runeSlice[7] = '中'
	runeSlice = append(runeSlice, '国')
	fmt.Printf("修改后: %s\n", string(runeSlice))

	// ========== 3. strings 包常用函数 ==========
	fmt.Println("\n--- 3. strings 包 ---")

	text := "  Hello, Go World!  "

	// 查找
	fmt.Printf("Contains(\"Go\"): %v\n", strings.Contains(text, "Go"))
	fmt.Printf("HasPrefix(\"  Hello\"): %v\n", strings.HasPrefix(text, "  Hello"))
	fmt.Printf("HasSuffix(\"!  \"): %v\n", strings.HasSuffix(text, "!  "))
	fmt.Printf("Index(\"Go\"): %d\n", strings.Index(text, "Go"))
	fmt.Printf("Count(\"l\"): %d\n", strings.Count(text, "l"))

	// 转换
	fmt.Printf("ToUpper: %q\n", strings.ToUpper(text))
	fmt.Printf("ToLower: %q\n", strings.ToLower(text))
	fmt.Printf("TrimSpace: %q\n", strings.TrimSpace(text))
	fmt.Printf("Trim(\"!\"): %q\n", strings.Trim("!!!hello!!!", "!"))

	// 分割与连接
	csv := "Go,Rust,Python,Java"
	parts := strings.Split(csv, ",")
	fmt.Printf("Split: %v\n", parts)
	fmt.Printf("Join: %q\n", strings.Join(parts, " | "))

	// 替换
	fmt.Printf("Replace: %q\n", strings.Replace("aabbcc", "b", "B", 1))
	fmt.Printf("ReplaceAll: %q\n", strings.ReplaceAll("aabbcc", "b", "B"))

	// 重复
	fmt.Printf("Repeat: %q\n", strings.Repeat("Go! ", 3))

	// ========== 4. strconv 包 ==========
	fmt.Println("\n--- 4. strconv 包（字符串与数字转换） ---")

	// int ↔ string
	numStr := strconv.Itoa(42)
	fmt.Printf("Itoa(42) = %q\n", numStr)

	num, err := strconv.Atoi("123")
	fmt.Printf("Atoi(\"123\") = %d, err=%v\n", num, err)

	_, err = strconv.Atoi("abc")
	fmt.Printf("Atoi(\"abc\") err=%v\n", err)

	// float ↔ string
	fStr := strconv.FormatFloat(3.14159, 'f', 2, 64)
	fmt.Printf("FormatFloat(3.14159, 2位) = %q\n", fStr)

	fVal, _ := strconv.ParseFloat("3.14", 64)
	fmt.Printf("ParseFloat(\"3.14\") = %v\n", fVal)

	// bool ↔ string
	bStr := strconv.FormatBool(true)
	fmt.Printf("FormatBool(true) = %q\n", bStr)

	bVal, _ := strconv.ParseBool("true")
	fmt.Printf("ParseBool(\"true\") = %v\n", bVal)

	// ========== 5. 字符串拼接性能对比 ==========
	fmt.Println("\n--- 5. 字符串拼接方式 ---")

	// 方式1: + 拼接（性能最差，每次分配新内存）
	result1 := "Hello" + ", " + "World"
	fmt.Printf("+ 拼接: %q\n", result1)

	// 方式2: fmt.Sprintf（有反射开销）
	result2 := fmt.Sprintf("%s, %s", "Hello", "World")
	fmt.Printf("Sprintf: %q\n", result2)

	// 方式3: strings.Join（一次分配）
	result3 := strings.Join([]string{"Hello", "World"}, ", ")
	fmt.Printf("Join: %q\n", result3)

	// 方式4: strings.Builder（最佳，推荐大量拼接时使用）
	var builder strings.Builder
	builder.Grow(100) // 预分配容量
	for i := 0; i < 5; i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("item%d", i))
	}
	fmt.Printf("Builder: %q\n", builder.String())

	// ========== 6. 多行字符串 ==========
	fmt.Println("\n--- 6. 多行字符串（反引号） ---")
	multiLine := `这是多行字符串
    保留原始格式
    不解释转义字符 \n \t
    适合 SQL、JSON、HTML 模板`
	fmt.Println(multiLine)

	// ========== 7. 字符串切片陷阱 ==========
	fmt.Println("\n--- 7. 字符串切片陷阱 ---")
	chinese := "你好世界"
	fmt.Printf("字符串: %q, 字节数: %d\n", chinese, len(chinese))

	// 安全切片：按 rune
	runes2 := []rune(chinese)
	fmt.Printf("前两个字符: %s\n", string(runes2[:2]))

	// 危险切片：按 byte 可能截断字符
	fmt.Printf("前3字节: %q (刚好一个中文字符)\n", chinese[:3])
	fmt.Printf("前4字节: %q (截断了！乱码)\n", chinese[:4])

	fmt.Println("\n========== 示例结束 ==========")
}
