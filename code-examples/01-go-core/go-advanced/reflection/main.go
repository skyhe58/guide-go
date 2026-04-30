// Go 进阶特性 — 反射（Reflection）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 反射的核心操作：
// 1. reflect.Type —— 获取类型信息
// 2. reflect.Value —— 读写值
// 3. 结构体标签解析 —— JSON/自定义标签
// 4. 动态方法调用
// 5. 反射性能对比
//
// 适用场景：
//   - 框架/库开发：JSON 序列化、ORM、依赖注入、配置解析
//   - 通用工具：结构体标签处理、动态代理
//   - 调试/日志：打印任意类型的详细信息
//
// 最佳实践：
//   - 缓存 reflect.Type 信息，避免重复获取
//   - 优先使用接口和泛型，反射是最后手段
//   - 框架中使用反射，业务代码中避免使用
//
// 常见陷阱：
//   - 对不可设置的 Value 调用 Set 方法会 panic
//   - 对 nil 接口值调用 reflect.ValueOf 返回零值 Value
//   - 未导出字段无法通过反射设置
//   - 类型不匹配的 Set 调用会 panic（如对 int32 调用 SetInt）
package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ============================================================
// 1. reflect.Type —— 获取类型信息
// ============================================================

// User 结构体，带有多种标签
type User struct {
	Name  string `json:"name" validate:"required" doc:"用户姓名"`
	Age   int    `json:"age" validate:"min=0,max=150" doc:"用户年龄"`
	Email string `json:"email,omitempty" validate:"email" doc:"邮箱地址"`
}

// inspectType 打印类型的详细信息
func inspectType(v any) {
	t := reflect.TypeOf(v)
	fmt.Printf("  类型名: %s\n", t.Name())
	fmt.Printf("  种类(Kind): %s\n", t.Kind())
	fmt.Printf("  大小: %d 字节\n", t.Size())

	// 如果是结构体，打印字段信息
	if t.Kind() == reflect.Struct {
		fmt.Printf("  字段数: %d\n", t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Printf("    [%d] %s (%s) — 标签: %s\n",
				i, f.Name, f.Type, f.Tag)
		}
	}
}

// ============================================================
// 2. reflect.Value —— 读写值
// ============================================================

// modifyValue 通过反射修改值（必须传入指针）
func modifyValue() {
	x := 42
	fmt.Printf("  修改前: x = %d\n", x)

	// 反射三定律第三条：要修改值，必须传入指针
	v := reflect.ValueOf(&x).Elem() // Elem() 获取指针指向的值
	if v.CanSet() {
		v.SetInt(100)
	}
	fmt.Printf("  修改后: x = %d\n", x)

	// 修改结构体字段
	user := User{Name: "张三", Age: 25, Email: "zhangsan@example.com"}
	fmt.Printf("  修改前: %+v\n", user)

	rv := reflect.ValueOf(&user).Elem()
	nameField := rv.FieldByName("Name")
	if nameField.CanSet() {
		nameField.SetString("李四")
	}
	ageField := rv.FieldByName("Age")
	if ageField.CanSet() {
		ageField.SetInt(30)
	}
	fmt.Printf("  修改后: %+v\n", user)
}

// ============================================================
// 3. 结构体标签解析
// ============================================================

// FieldInfo 存储解析后的字段信息
type FieldInfo struct {
	Name     string
	JSONName string
	Required bool
	Doc      string
}

// parseStructTags 解析结构体的标签信息
func parseStructTags(v any) []FieldInfo {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var fields []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 解析 json 标签
		jsonTag := field.Tag.Get("json")
		jsonName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}

		// 解析 validate 标签
		validateTag := field.Tag.Get("validate")
		required := strings.Contains(validateTag, "required")

		// 解析 doc 标签
		doc := field.Tag.Get("doc")

		fields = append(fields, FieldInfo{
			Name:     field.Name,
			JSONName: jsonName,
			Required: required,
			Doc:      doc,
		})
	}
	return fields
}

// ============================================================
// 4. 动态方法调用
// ============================================================

// Calculator 计算器，用于演示动态方法调用
type Calculator struct{}

// Add 加法
func (c Calculator) Add(a, b int) int { return a + b }

// Multiply 乘法
func (c Calculator) Multiply(a, b int) int { return a * b }

// Greet 问候（不同参数类型）
func (c Calculator) Greet(name string) string {
	return fmt.Sprintf("你好, %s!", name)
}

// callMethod 通过反射动态调用方法
func callMethod(obj any, methodName string, args ...any) ([]any, error) {
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("方法 %s 不存在", methodName)
	}

	// 构建参数
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 调用方法
	results := method.Call(in)

	// 提取返回值
	out := make([]any, len(results))
	for i, r := range results {
		out[i] = r.Interface()
	}
	return out, nil
}

// ============================================================
// 5. 反射性能对比
// ============================================================

// benchmarkComparison 对比直接访问和反射访问的性能
func benchmarkComparison() {
	user := User{Name: "张三", Age: 25, Email: "test@example.com"}
	iterations := 1_000_000

	// 直接访问
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = user.Name
	}
	directDuration := time.Since(start)

	// 反射访问
	start = time.Now()
	v := reflect.ValueOf(user)
	for i := 0; i < iterations; i++ {
		_ = v.FieldByName("Name").String()
	}
	reflectDuration := time.Since(start)

	// 缓存字段索引的反射访问
	start = time.Now()
	t := reflect.TypeOf(user)
	nameIdx := -1
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Name" {
			nameIdx = i
			break
		}
	}
	for i := 0; i < iterations; i++ {
		_ = v.Field(nameIdx).String()
	}
	cachedDuration := time.Since(start)

	fmt.Printf("  直接访问:       %v\n", directDuration)
	fmt.Printf("  反射(ByName):   %v\n", reflectDuration)
	fmt.Printf("  反射(缓存索引): %v\n", cachedDuration)
	fmt.Printf("  反射/直接 比率: %.1fx\n",
		float64(reflectDuration)/float64(directDuration))
}

// ============================================================
// 通用工具：深度打印任意值
// ============================================================

// deepPrint 使用反射打印任意值的详细信息
func deepPrint(v any, indent string) {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	switch rv.Kind() {
	case reflect.Struct:
		fmt.Printf("%s%s {\n", indent, rt.Name())
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			value := rv.Field(i)
			fmt.Printf("%s  %s: %v\n", indent, field.Name, value.Interface())
		}
		fmt.Printf("%s}\n", indent)
	case reflect.Slice:
		fmt.Printf("%s[]%s (len=%d) [\n", indent, rt.Elem().Name(), rv.Len())
		for i := 0; i < rv.Len(); i++ {
			fmt.Printf("%s  [%d]: %v\n", indent, i, rv.Index(i).Interface())
		}
		fmt.Printf("%s]\n", indent)
	case reflect.Map:
		fmt.Printf("%smap[%s]%s (len=%d) {\n", indent,
			rt.Key().Name(), rt.Elem().Name(), rv.Len())
		for _, key := range rv.MapKeys() {
			fmt.Printf("%s  %v: %v\n", indent, key.Interface(), rv.MapIndex(key).Interface())
		}
		fmt.Printf("%s}\n", indent)
	default:
		fmt.Printf("%s%v (%s)\n", indent, rv.Interface(), rt.Name())
	}
}

func main() {
	fmt.Println("========== Go 反射示例 ==========")

	// --- 1. 类型信息 ---
	fmt.Println("\n--- 1. reflect.Type 类型信息 ---")
	inspectType(User{})

	// --- 2. 读写值 ---
	fmt.Println("\n--- 2. reflect.Value 读写值 ---")
	modifyValue()

	// --- 3. 结构体标签解析 ---
	fmt.Println("\n--- 3. 结构体标签解析 ---")
	fields := parseStructTags(User{})
	for _, f := range fields {
		required := ""
		if f.Required {
			required = " [必填]"
		}
		fmt.Printf("  %s → JSON: %q, 说明: %s%s\n",
			f.Name, f.JSONName, f.Doc, required)
	}

	// --- 4. 动态方法调用 ---
	fmt.Println("\n--- 4. 动态方法调用 ---")
	calc := Calculator{}

	result, err := callMethod(calc, "Add", 10, 20)
	if err == nil {
		fmt.Printf("  Add(10, 20) = %v\n", result[0])
	}

	result, err = callMethod(calc, "Multiply", 6, 7)
	if err == nil {
		fmt.Printf("  Multiply(6, 7) = %v\n", result[0])
	}

	result, err = callMethod(calc, "Greet", "Go 开发者")
	if err == nil {
		fmt.Printf("  Greet(\"Go 开发者\") = %v\n", result[0])
	}

	// 调用不存在的方法
	_, err = callMethod(calc, "Divide", 10, 3)
	if err != nil {
		fmt.Printf("  Divide: %v\n", err)
	}

	// --- 5. 性能对比 ---
	fmt.Println("\n--- 5. 反射性能对比（100万次访问）---")
	benchmarkComparison()

	// --- 6. 通用深度打印 ---
	fmt.Println("\n--- 6. 通用深度打印 ---")
	deepPrint(User{Name: "张三", Age: 25, Email: "test@example.com"}, "  ")
	deepPrint([]int{1, 2, 3, 4, 5}, "  ")
	deepPrint(map[string]int{"Go": 1, "Rust": 2}, "  ")

	fmt.Println("\n========== 示例结束 ==========")
}
