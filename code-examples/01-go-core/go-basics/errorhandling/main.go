// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 错误处理示例：error 接口、errors.New、fmt.Errorf、%w 包装、errors.Is、errors.As、panic 与 recover
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// ========== 自定义错误类型 ==========

// NotFoundError 资源未找到错误
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s(id=%d) 不存在", e.Resource, e.ID)
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %q 验证失败: %s", e.Field, e.Message)
}

// ========== 哨兵错误 ==========
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal error")
)

func main() {
	fmt.Println("========== Go 错误处理示例 ==========")

	// ========== 1. 基本错误处理 ==========
	fmt.Println("\n--- 1. 基本错误处理 ---")

	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("10 / 3 = %.2f\n", result)
	}

	_, err = divide(10, 0)
	if err != nil {
		fmt.Println("错误:", err)
	}

	// ========== 2. 创建错误的方式 ==========
	fmt.Println("\n--- 2. 创建错误的方式 ---")

	// errors.New
	err1 := errors.New("简单错误")
	fmt.Println("errors.New:", err1)

	// fmt.Errorf
	userID := 42
	err2 := fmt.Errorf("用户 %d 不存在", userID)
	fmt.Println("fmt.Errorf:", err2)

	// 自定义错误类型
	err3 := &NotFoundError{Resource: "用户", ID: 42}
	fmt.Println("自定义错误:", err3)

	// ========== 3. 错误包装与链 ==========
	fmt.Println("\n--- 3. 错误包装 (%w) ---")

	// 模拟多层调用的错误包装
	err = getUser(999)
	fmt.Println("包装后的错误:", err)

	// errors.Is — 沿错误链查找特定错误值
	fmt.Println("\n--- 4. errors.Is ---")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("✅ errors.Is: 确认是 NotFound 错误")
	}

	// errors.As — 沿错误链查找特定错误类型
	fmt.Println("\n--- 5. errors.As ---")
	var notFound *NotFoundError
	if errors.As(err, &notFound) {
		fmt.Printf("✅ errors.As: 资源=%s, ID=%d\n", notFound.Resource, notFound.ID)
	}

	// ========== 4. 错误处理最佳实践 ==========
	fmt.Println("\n--- 6. 错误处理最佳实践 ---")

	// 多错误类型判断
	testCases := []string{"42", "abc", "-1"}
	for _, tc := range testCases {
		val, err := parsePositiveInt(tc)
		if err != nil {
			var valErr *ValidationError
			if errors.As(err, &valErr) {
				fmt.Printf("  验证错误 [%s]: %s\n", tc, valErr.Message)
			} else {
				fmt.Printf("  解析错误 [%s]: %v\n", tc, err)
			}
		} else {
			fmt.Printf("  成功 [%s]: %d\n", tc, val)
		}
	}

	// ========== 5. panic 与 recover ==========
	fmt.Println("\n--- 7. panic 与 recover ---")

	// recover 捕获 panic
	fmt.Println("调用 safeDivide(10, 0):")
	result2, err := safeDivide(10, 0)
	if err != nil {
		fmt.Println("  捕获到 panic:", err)
	} else {
		fmt.Println("  结果:", result2)
	}

	fmt.Println("调用 safeDivide(10, 3):")
	result2, err = safeDivide(10, 3)
	if err != nil {
		fmt.Println("  错误:", err)
	} else {
		fmt.Printf("  结果: %d\n", result2)
	}

	// 演示 panic 后 defer 仍然执行
	fmt.Println("\npanic 后 defer 仍然执行:")
	panicWithDefer()

	fmt.Println("\n========== 示例结束 ==========")
}

// 基本错误返回
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil
}

// 错误包装链
func findUserByID(id int) (*NotFoundError, error) {
	if id <= 0 || id > 100 {
		return nil, &NotFoundError{Resource: "用户", ID: id}
	}
	return nil, nil
}

func getUser(id int) error {
	_, err := findUserByID(id)
	if err != nil {
		// 使用 %w 包装错误，保留错误链
		return fmt.Errorf("获取用户失败: %w", err)
	}
	return nil
}

// 解析正整数（演示多种错误类型）
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("解析 %q 失败: %w", s, err)
	}
	if n <= 0 {
		return 0, &ValidationError{
			Field:   "number",
			Message: fmt.Sprintf("必须为正整数，实际值: %d", n),
		}
	}
	return n, nil
}

// panic + recover
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil // b=0 时会 panic
}

// 演示 panic 后 defer 执行
func panicWithDefer() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  recover 捕获: %v\n", r)
		}
	}()
	defer fmt.Println("  defer 1: 我会执行")
	defer fmt.Println("  defer 2: 我也会执行")

	fmt.Println("  即将 panic...")
	panic("出了大问题！")
}
