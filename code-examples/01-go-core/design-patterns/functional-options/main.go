// Functional Options Pattern — Go 最推崇的配置模式
// 演示如何使用 Functional Options 实现灵活的对象构造
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式: go run ./functional-options/

package main

import (
	"fmt"
	"time"
)

// =============================================================================
// Part A: Functional Options Pattern 完整演示
// =============================================================================

// Server 表示一个 HTTP 服务器配置
type Server struct {
	host       string
	port       int
	timeout    time.Duration
	maxConn    int
	enableTLS  bool
	certFile   string
	keyFile    string
}

// Option 定义服务器配置选项的函数类型
type Option func(*Server)

// WithPort 设置服务器端口
func WithPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// WithMaxConn 设置最大连接数
func WithMaxConn(maxConn int) Option {
	return func(s *Server) {
		s.maxConn = maxConn
	}
}

// WithTLS 启用 TLS 加密
func WithTLS(certFile, keyFile string) Option {
	return func(s *Server) {
		s.enableTLS = true
		s.certFile = certFile
		s.keyFile = keyFile
	}
}

// NewServer 使用 Functional Options 创建服务器
// host 是必需参数，其余通过 Option 可选配置
func NewServer(host string, opts ...Option) *Server {
	// 设置默认值
	s := &Server{
		host:    host,
		port:    8080,
		timeout: 30 * time.Second,
		maxConn: 100,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// String 返回服务器配置的字符串表示
func (s *Server) String() string {
	tlsInfo := "关闭"
	if s.enableTLS {
		tlsInfo = fmt.Sprintf("开启 (cert: %s)", s.certFile)
	}
	return fmt.Sprintf(
		"Server{host: %s, port: %d, timeout: %v, maxConn: %d, TLS: %s}",
		s.host, s.port, s.timeout, s.maxConn, tlsInfo,
	)
}

// =============================================================================
// 进阶：带校验的 Functional Options
// =============================================================================

// OptionWithError 带错误返回的选项类型
type OptionWithError func(*Server) error

// WithPortValidated 带校验的端口设置
func WithPortValidated(port int) OptionWithError {
	return func(s *Server) error {
		if port < 1 || port > 65535 {
			return fmt.Errorf("无效端口号: %d (有效范围: 1-65535)", port)
		}
		s.port = port
		return nil
	}
}

// NewServerValidated 带校验的服务器创建
func NewServerValidated(host string, opts ...OptionWithError) (*Server, error) {
	s := &Server{
		host:    host,
		port:    8080,
		timeout: 30 * time.Second,
		maxConn: 100,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("配置服务器失败: %w", err)
		}
	}

	return s, nil
}

func main() {
	fmt.Println("=== Functional Options Pattern 演示 ===")
	fmt.Println()

	// 1. 使用默认配置
	fmt.Println("--- 1. 默认配置 ---")
	s1 := NewServer("localhost")
	fmt.Println(s1)
	fmt.Println()

	// 2. 自定义部分配置
	fmt.Println("--- 2. 自定义部分配置 ---")
	s2 := NewServer("0.0.0.0",
		WithPort(9090),
		WithTimeout(60*time.Second),
	)
	fmt.Println(s2)
	fmt.Println()

	// 3. 完整自定义配置
	fmt.Println("--- 3. 完整自定义配置 ---")
	s3 := NewServer("api.example.com",
		WithPort(443),
		WithTimeout(120*time.Second),
		WithMaxConn(1000),
		WithTLS("/etc/ssl/cert.pem", "/etc/ssl/key.pem"),
	)
	fmt.Println(s3)
	fmt.Println()

	// 4. 带校验的 Options
	fmt.Println("--- 4. 带校验的 Options ---")
	s4, err := NewServerValidated("localhost",
		WithPortValidated(8080),
	)
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
	} else {
		fmt.Printf("创建成功: %s\n", s4)
	}

	// 校验失败的情况
	_, err = NewServerValidated("localhost",
		WithPortValidated(99999), // 无效端口
	)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	}

	fmt.Println()
	fmt.Println("=== 实际应用场景 ===")
	fmt.Println("- gRPC-Go: grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock())")
	fmt.Println("- zap 日志: zap.New(core, zap.AddCaller(), zap.AddStacktrace(...))")
	fmt.Println("- Docker client: client.NewClientWithOpts(client.FromEnv)")
}
