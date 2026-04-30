// 短链接系统 — 架构设计场景
// 本示例演示短链接系统的核心功能：Base62 编码、哈希生成短码、冲突检测、URL 映射存储
// 使用纯 Go 实现，无需外部依赖
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：go run ./04-distributed/architecture/short-url/

package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// ============================================================
// Part A：短链接系统核心实现（纯内存模拟）
// ============================================================

// --- 组件 1：Base62 编码器 ---

// base62Chars Base62 字符集：0-9 a-z A-Z 共 62 个字符
// 6 位 Base62 短码可表示 62^6 ≈ 568 亿个不同的 URL
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62Encode 将无符号整数编码为 Base62 字符串
// 原理：类似十进制转换，不断除以 62 取余数
func Base62Encode(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	for num > 0 {
		remainder := num % 62
		result = append([]byte{base62Chars[remainder]}, result...)
		num /= 62
	}
	return string(result)
}

// Base62Decode 将 Base62 字符串解码为无符号整数
// 原理：类似十进制解析，从高位到低位逐位累加
func Base62Decode(s string) uint64 {
	var result uint64
	for _, c := range s {
		result *= 62
		idx := strings.IndexRune(base62Chars, c)
		if idx < 0 {
			return 0 // 非法字符
		}
		result += uint64(idx)
	}
	return result
}

// --- 组件 2：短码生成器 ---

// ShortCodeGenerator 短码生成器
// 支持两种生成策略：哈希截取 和 自增 ID
type ShortCodeGenerator struct {
	codeLength int // 短码长度（默认 6 位）
}

// NewShortCodeGenerator 创建短码生成器
func NewShortCodeGenerator(codeLength int) *ShortCodeGenerator {
	return &ShortCodeGenerator{codeLength: codeLength}
}

// GenerateByHash 基于哈希生成短码（推荐方案）
// 对长 URL 计算 MD5 哈希，取前 8 字节转为 uint64，再 Base62 编码
// 优点：分布式友好，无中心依赖
// 缺点：可能冲突，需要冲突检测
func (g *ShortCodeGenerator) GenerateByHash(longURL string) string {
	// 计算 MD5 哈希
	hash := md5.Sum([]byte(longURL))
	// 取前 8 字节转为 uint64
	num := binary.BigEndian.Uint64(hash[:8])
	// Base62 编码
	code := Base62Encode(num)
	// 截取指定长度
	if len(code) > g.codeLength {
		code = code[:g.codeLength]
	}
	// 如果长度不足，左侧补零
	for len(code) < g.codeLength {
		code = "0" + code
	}
	return code
}

// GenerateByHashWithSalt 带盐值的哈希生成（用于冲突重试）
// 在原始 URL 后追加随机盐值，生成不同的短码
func (g *ShortCodeGenerator) GenerateByHashWithSalt(longURL string, salt string) string {
	return g.GenerateByHash(longURL + salt)
}

// GenerateByID 基于自增 ID 生成短码
// 优点：无冲突，短码最短
// 缺点：依赖中心化 ID 生成器，短码可预测
func (g *ShortCodeGenerator) GenerateByID(id uint64) string {
	code := Base62Encode(id)
	for len(code) < g.codeLength {
		code = "0" + code
	}
	return code
}

// --- 组件 3：URL 映射存储 ---

// URLMapping URL 映射记录
type URLMapping struct {
	ShortCode string    // 短码
	LongURL   string    // 原始长 URL
	CreatedAt time.Time // 创建时间
	Clicks    int64     // 点击次数
	ExpiresAt time.Time // 过期时间（零值表示永不过期）
}

// URLStore URL 映射存储
// 使用 sync.Map 模拟 Redis/数据库存储
// 维护两个映射：短码→长URL（用于重定向）、长URL→短码（用于去重）
type URLStore struct {
	shortToLong sync.Map // shortCode → URLMapping
	longToShort sync.Map // longURL → shortCode（去重用）
	generator   *ShortCodeGenerator
	idCounter   uint64 // 自增 ID 计数器
	mu          sync.Mutex
}

// NewURLStore 创建 URL 存储
func NewURLStore(codeLength int) *URLStore {
	return &URLStore{
		generator: NewShortCodeGenerator(codeLength),
	}
}

// maxRetries 冲突重试最大次数
const maxRetries = 5

// CreateShortURL 创建短链接
// 流程：去重检查 → 哈希生成 → 冲突检测 → 存储映射
func (s *URLStore) CreateShortURL(longURL string) (string, bool) {
	// 去重：检查长 URL 是否已有对应短码
	if code, ok := s.longToShort.Load(longURL); ok {
		return code.(string), false // 返回已有短码，isNew=false
	}

	// 生成短码（哈希方式）
	code := s.generator.GenerateByHash(longURL)

	// 冲突检测与重试
	for i := 0; i < maxRetries; i++ {
		if existing, loaded := s.shortToLong.LoadOrStore(code, URLMapping{
			ShortCode: code,
			LongURL:   longURL,
			CreatedAt: time.Now(),
		}); !loaded {
			// 存储成功（无冲突）
			s.longToShort.Store(longURL, code)
			return code, true // isNew=true
		} else if existing.(URLMapping).LongURL == longURL {
			// 相同 URL 已存在，返回已有短码
			return code, false
		}
		// 冲突：追加随机盐值重新生成
		salt := fmt.Sprintf("_%d_%d", i, rand.Int63())
		code = s.generator.GenerateByHashWithSalt(longURL, salt)
	}

	// 所有重试都冲突，使用自增 ID 兜底
	s.mu.Lock()
	s.idCounter++
	id := s.idCounter
	s.mu.Unlock()
	code = s.generator.GenerateByID(id)
	s.shortToLong.Store(code, URLMapping{
		ShortCode: code,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	})
	s.longToShort.Store(longURL, code)
	return code, true
}

// Resolve 解析短码，返回原始长 URL（模拟重定向）
// 同时增加点击计数
func (s *URLStore) Resolve(shortCode string) (string, bool) {
	val, ok := s.shortToLong.Load(shortCode)
	if !ok {
		return "", false
	}
	mapping := val.(URLMapping)

	// 检查是否过期
	if !mapping.ExpiresAt.IsZero() && time.Now().After(mapping.ExpiresAt) {
		s.Delete(shortCode)
		return "", false
	}

	// 增加点击计数（非原子操作，生产环境应使用 atomic 或 Redis INCR）
	mapping.Clicks++
	s.shortToLong.Store(shortCode, mapping)

	return mapping.LongURL, true
}

// Delete 删除短链接
func (s *URLStore) Delete(shortCode string) bool {
	val, ok := s.shortToLong.Load(shortCode)
	if !ok {
		return false
	}
	mapping := val.(URLMapping)
	s.shortToLong.Delete(shortCode)
	s.longToShort.Delete(mapping.LongURL)
	return true
}

// GetStats 获取短链接统计信息
func (s *URLStore) GetStats(shortCode string) (URLMapping, bool) {
	val, ok := s.shortToLong.Load(shortCode)
	if !ok {
		return URLMapping{}, false
	}
	return val.(URLMapping), true
}

// ListAll 列出所有短链接（调试用）
func (s *URLStore) ListAll() []URLMapping {
	var mappings []URLMapping
	s.shortToLong.Range(func(key, value interface{}) bool {
		mappings = append(mappings, value.(URLMapping))
		return true
	})
	return mappings
}

func main() {
	fmt.Println("========== 短链接系统演示 ==========")
	fmt.Println()

	// --- 演示 1：Base62 编码/解码 ---
	fmt.Println("--- 1. Base62 编码/解码 ---")
	testNumbers := []uint64{0, 1, 61, 62, 1000, 568800000, 3521614606207}
	for _, num := range testNumbers {
		encoded := Base62Encode(num)
		decoded := Base62Decode(encoded)
		fmt.Printf("  %15d → %-10s → %d", num, encoded, decoded)
		if num == decoded {
			fmt.Println(" ✅")
		} else {
			fmt.Println(" ❌")
		}
	}
	fmt.Printf("\n  6 位 Base62 最大值: 62^6 = %d（约 568 亿）\n", 62*62*62*62*62*62)
	fmt.Println()

	// --- 演示 2：短码生成 ---
	fmt.Println("--- 2. 短码生成（哈希方式） ---")
	gen := NewShortCodeGenerator(6)
	testURLs := []string{
		"https://www.example.com/very/long/path/to/some/resource?param=value",
		"https://github.com/golang/go/issues/12345",
		"https://pkg.go.dev/sync/atomic",
		"https://redis.io/commands/decr/",
	}
	for _, url := range testURLs {
		code := gen.GenerateByHash(url)
		fmt.Printf("  %s\n    → 短码: %s\n", url, code)
	}
	fmt.Println()

	// 相同 URL 生成相同短码
	code1 := gen.GenerateByHash(testURLs[0])
	code2 := gen.GenerateByHash(testURLs[0])
	fmt.Printf("  相同 URL 生成相同短码: %s == %s → %v\n", code1, code2, code1 == code2)
	fmt.Println()

	// --- 演示 3：完整的短链接 CRUD ---
	fmt.Println("--- 3. 短链接 CRUD 操作 ---")
	store := NewURLStore(6)

	// 创建短链接
	fmt.Println("  [创建]")
	urls := []string{
		"https://www.google.com/search?q=golang+tutorial",
		"https://go.dev/doc/effective_go",
		"https://github.com/gin-gonic/gin",
	}
	codes := make([]string, len(urls))
	for i, url := range urls {
		code, isNew := store.CreateShortURL(url)
		codes[i] = code
		fmt.Printf("    %s → %s (新建: %v)\n", url, code, isNew)
	}
	fmt.Println()

	// 去重测试：相同 URL 返回已有短码
	fmt.Println("  [去重测试]")
	code, isNew := store.CreateShortURL(urls[0])
	fmt.Printf("    重复创建: %s → %s (新建: %v)\n", urls[0], code, isNew)
	fmt.Println()

	// 解析短码（模拟重定向）
	fmt.Println("  [解析/重定向]")
	for _, c := range codes {
		if longURL, ok := store.Resolve(c); ok {
			fmt.Printf("    %s → %s ✅\n", c, longURL)
		}
	}
	// 解析不存在的短码
	if _, ok := store.Resolve("XXXXXX"); !ok {
		fmt.Println("    XXXXXX → 404 Not Found ✅")
	}
	fmt.Println()

	// 点击统计
	fmt.Println("  [点击统计]")
	// 模拟多次访问
	for i := 0; i < 5; i++ {
		store.Resolve(codes[0])
	}
	if stats, ok := store.GetStats(codes[0]); ok {
		fmt.Printf("    短码 %s 点击次数: %d\n", codes[0], stats.Clicks)
	}
	fmt.Println()

	// 删除短链接
	fmt.Println("  [删除]")
	deleted := store.Delete(codes[2])
	fmt.Printf("    删除 %s: %v\n", codes[2], deleted)
	if _, ok := store.Resolve(codes[2]); !ok {
		fmt.Printf("    解析 %s: 已删除 ✅\n", codes[2])
	}
	fmt.Println()

	// --- 演示 4：并发安全测试 ---
	fmt.Println("--- 4. 并发安全测试 ---")
	concurrentStore := NewURLStore(6)
	var wg sync.WaitGroup
	createdCodes := sync.Map{}

	// 100 个 goroutine 并发创建短链接
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			url := fmt.Sprintf("https://example.com/page/%d", id)
			code, _ := concurrentStore.CreateShortURL(url)
			createdCodes.Store(code, url)
		}(i)
	}
	wg.Wait()

	// 统计创建结果
	totalCreated := 0
	concurrentStore.shortToLong.Range(func(key, value interface{}) bool {
		totalCreated++
		return true
	})
	fmt.Printf("  100 个并发请求创建了 %d 个短链接\n", totalCreated)

	// 验证所有短链接都能正确解析
	allResolved := true
	concurrentStore.shortToLong.Range(func(key, value interface{}) bool {
		code := key.(string)
		mapping := value.(URLMapping)
		if resolved, ok := concurrentStore.Resolve(code); !ok || resolved != mapping.LongURL {
			allResolved = false
			return false
		}
		return true
	})
	if allResolved {
		fmt.Println("  所有短链接解析验证通过 ✅")
	} else {
		fmt.Println("  部分短链接解析失败 ❌")
	}
}
