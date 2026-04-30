// gRPC 四种通信模式演示（简化版）
// 演示：Unary、Server Streaming、Client Streaming、Bidirectional Streaming
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例使用简化方式演示 gRPC 核心概念，无需 protoc 编译。
// 实际项目中应使用 .proto 文件定义服务，通过 protoc 生成代码。
//
// 运行方式：go run main.go
//
// 说明：
//   本 demo 通过 google.golang.org/grpc 库创建真实的 gRPC 服务端和客户端，
//   演示四种通信模式的核心概念。为简化演示，使用手动注册而非 protoc 生成代码。

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ============================================================
// 概念说明：gRPC 四种通信模式
// ============================================================
//
// 1. Unary（一元调用）：
//    客户端发送一个请求，服务端返回一个响应。
//    最常用的模式，类似普通的 HTTP 请求-响应。
//    场景：获取用户信息、创建订单等。
//
// 2. Server Streaming（服务端流）：
//    客户端发送一个请求，服务端返回一个数据流。
//    场景：实时日志推送、股票行情、文件下载。
//
// 3. Client Streaming（客户端流）：
//    客户端发送一个数据流，服务端返回一个响应。
//    场景：文件上传、批量数据导入、传感器数据上报。
//
// 4. Bidirectional Streaming（双向流）：
//    客户端和服务端同时发送数据流。
//    场景：聊天应用、实时协作、游戏通信。

// ============================================================
// 模拟 gRPC 四种通信模式（纯 Go 演示）
// ============================================================

// User 用户模型（模拟 protobuf 消息）
type User struct {
	ID    int64
	Name  string
	Email string
}

// ChatMessage 聊天消息（模拟 protobuf 消息）
type ChatMessage struct {
	From    string
	Content string
	Time    time.Time
}

// ============================================================
// 演示 1：Unary（一元调用）
// ============================================================

func demoUnary() {
	fmt.Println("=" + repeatStr("=", 59))
	fmt.Println("📡 模式 1：Unary（一元调用）")
	fmt.Println("   客户端发送一个请求，服务端返回一个响应")
	fmt.Println(repeatStr("=", 60))

	// 模拟 Unary RPC：GetUser
	getUser := func(ctx context.Context, userID int64) (*User, error) {
		// 模拟服务端处理
		log.Printf("[服务端] 收到 GetUser 请求，ID=%d", userID)
		time.Sleep(50 * time.Millisecond) // 模拟处理耗时

		return &User{
			ID:    userID,
			Name:  fmt.Sprintf("用户_%d", userID),
			Email: fmt.Sprintf("user%d@example.com", userID),
		}, nil
	}

	// 客户端调用
	ctx := context.Background()
	user, err := getUser(ctx, 42)
	if err != nil {
		log.Printf("调用失败: %v", err)
		return
	}
	log.Printf("[客户端] 收到响应: ID=%d, Name=%s, Email=%s",
		user.ID, user.Name, user.Email)
	fmt.Println()
}

// ============================================================
// 演示 2：Server Streaming（服务端流）
// ============================================================

func demoServerStreaming() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📡 模式 2：Server Streaming（服务端流）")
	fmt.Println("   客户端发送一个请求，服务端返回数据流")
	fmt.Println(repeatStr("=", 60))

	// 模拟 Server Streaming RPC：ListUsers
	// 服务端通过 channel 发送数据流
	listUsers := func(ctx context.Context, prefix string) <-chan *User {
		ch := make(chan *User)
		go func() {
			defer close(ch)
			log.Printf("[服务端] 收到 ListUsers 请求，前缀=%s", prefix)

			// 模拟逐条返回用户数据
			for i := 1; i <= 5; i++ {
				select {
				case <-ctx.Done():
					log.Println("[服务端] 客户端取消，停止发送")
					return
				default:
					user := &User{
						ID:   int64(i),
						Name: fmt.Sprintf("%s_用户_%d", prefix, i),
					}
					ch <- user
					time.Sleep(100 * time.Millisecond) // 模拟逐条发送
				}
			}
			log.Println("[服务端] 数据流发送完毕")
		}()
		return ch
	}

	// 客户端接收数据流
	ctx := context.Background()
	stream := listUsers(ctx, "Go")
	for user := range stream {
		log.Printf("[客户端] 收到流数据: ID=%d, Name=%s", user.ID, user.Name)
	}
	log.Println("[客户端] 数据流接收完毕")
	fmt.Println()
}

// ============================================================
// 演示 3：Client Streaming（客户端流）
// ============================================================

func demoClientStreaming() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📡 模式 3：Client Streaming（客户端流）")
	fmt.Println("   客户端发送数据流，服务端返回一个响应")
	fmt.Println(repeatStr("=", 60))

	// 模拟 Client Streaming RPC：BatchCreateUsers
	// 客户端通过 channel 发送数据流，服务端返回汇总结果
	batchCreateUsers := func(ctx context.Context, users <-chan *User) (int, error) {
		count := 0
		for user := range users {
			log.Printf("[服务端] 收到用户数据: Name=%s", user.Name)
			count++
			time.Sleep(50 * time.Millisecond) // 模拟处理
		}
		log.Printf("[服务端] 批量创建完成，共 %d 个用户", count)
		return count, nil
	}

	// 客户端发送数据流
	ch := make(chan *User)
	go func() {
		defer close(ch)
		names := []string{"张三", "李四", "王五", "赵六"}
		for i, name := range names {
			user := &User{ID: int64(i + 1), Name: name}
			log.Printf("[客户端] 发送用户: %s", name)
			ch <- user
			time.Sleep(80 * time.Millisecond)
		}
		log.Println("[客户端] 数据流发送完毕")
	}()

	ctx := context.Background()
	count, err := batchCreateUsers(ctx, ch)
	if err != nil {
		log.Printf("调用失败: %v", err)
		return
	}
	log.Printf("[客户端] 服务端响应: 成功创建 %d 个用户", count)
	fmt.Println()
}

// ============================================================
// 演示 4：Bidirectional Streaming（双向流）
// ============================================================

func demoBidirectionalStreaming() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📡 模式 4：Bidirectional Streaming（双向流）")
	fmt.Println("   客户端和服务端同时发送数据流")
	fmt.Println(repeatStr("=", 60))

	// 模拟 Bidirectional Streaming RPC：Chat
	// 使用两个 channel 模拟双向流
	clientToServer := make(chan *ChatMessage, 10)
	serverToClient := make(chan *ChatMessage, 10)

	var wg sync.WaitGroup

	// 服务端：接收消息并回复
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(serverToClient)

		for msg := range clientToServer {
			log.Printf("[服务端] 收到消息: [%s] %s", msg.From, msg.Content)

			// 服务端回复
			reply := &ChatMessage{
				From:    "服务端",
				Content: fmt.Sprintf("收到你的消息: '%s'", msg.Content),
				Time:    time.Now(),
			}
			serverToClient <- reply
		}
		log.Println("[服务端] 客户端关闭了连接")
	}()

	// 客户端：发送消息并接收回复
	wg.Add(1)
	go func() {
		defer wg.Done()

		for msg := range serverToClient {
			log.Printf("[客户端] 收到回复: [%s] %s", msg.From, msg.Content)
		}
		log.Println("[客户端] 服务端关闭了连接")
	}()

	// 客户端发送消息
	messages := []string{"你好！", "Go 语言真棒", "gRPC 双向流很强大", "再见！"}
	for _, content := range messages {
		msg := &ChatMessage{
			From:    "客户端",
			Content: content,
			Time:    time.Now(),
		}
		log.Printf("[客户端] 发送消息: %s", content)
		clientToServer <- msg
		time.Sleep(200 * time.Millisecond)
	}
	close(clientToServer)

	wg.Wait()
	fmt.Println()
}

// ============================================================
// 演示 5：真实 gRPC 服务端和客户端（Unary 模式）
// ============================================================

// echoServer 简单的 Echo 服务端（演示真实 gRPC 连接）
type echoServer struct{}

func demoRealGRPC() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📡 演示 5：真实 gRPC 服务端和客户端连接")
	fmt.Println("   使用 google.golang.org/grpc 建立真实连接")
	fmt.Println(repeatStr("=", 60))

	// 启动 gRPC 服务端
	listener, err := net.Listen("tcp", ":0") // 随机端口
	if err != nil {
		log.Printf("监听失败: %v", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// 创建 gRPC 服务器（带拦截器演示）
	server := grpc.NewServer(
		// Unary 拦截器：类似 HTTP 中间件
		grpc.UnaryInterceptor(func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (interface{}, error) {
			start := time.Now()
			log.Printf("[拦截器] 方法: %s 开始处理", info.FullMethod)

			// 调用实际处理函数
			resp, err := handler(ctx, req)

			log.Printf("[拦截器] 方法: %s 处理完成，耗时: %v",
				info.FullMethod, time.Since(start))
			return resp, err
		}),
	)

	// 启动服务端
	go func() {
		log.Printf("[服务端] gRPC 服务器启动在端口 %d", port)
		if err := server.Serve(listener); err != nil {
			log.Printf("服务端错误: %v", err)
		}
	}()

	// 等待服务端启动
	time.Sleep(100 * time.Millisecond)

	// 创建 gRPC 客户端连接
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[客户端] 已连接到 gRPC 服务端 localhost:%d", port)
	log.Println("[客户端] 连接状态:", conn.GetState().String())

	// 注意：由于没有注册具体服务，这里只演示连接建立
	// 实际项目中，需要通过 protoc 生成的代码注册服务和调用方法
	log.Println("[说明] 实际项目中，需要：")
	log.Println("  1. 编写 .proto 文件定义服务")
	log.Println("  2. 使用 protoc 生成 Go 代码")
	log.Println("  3. 实现生成的服务接口")
	log.Println("  4. 使用生成的客户端代码调用服务")

	// 关闭服务端
	server.GracefulStop()
	log.Println("[服务端] gRPC 服务器已优雅关闭")
	fmt.Println()
}

// ============================================================
// 演示 6：gRPC 拦截器概念
// ============================================================

func demoInterceptor() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📡 演示 6：gRPC 拦截器（Interceptor）概念")
	fmt.Println("   拦截器类似 HTTP 中间件，用于横切关注点")
	fmt.Println(repeatStr("=", 60))

	// 模拟拦截器链
	type InterceptorFunc func(method string, req interface{}, handler func() (interface{}, error)) (interface{}, error)

	// 日志拦截器
	loggingInterceptor := func(method string, req interface{}, handler func() (interface{}, error)) (interface{}, error) {
		start := time.Now()
		log.Printf("[日志拦截器] 开始调用: %s", method)

		resp, err := handler()

		log.Printf("[日志拦截器] 调用完成: %s, 耗时: %v, 错误: %v",
			method, time.Since(start), err)
		return resp, err
	}

	// 认证拦截器
	authInterceptor := func(method string, req interface{}, handler func() (interface{}, error)) (interface{}, error) {
		log.Printf("[认证拦截器] 验证请求: %s", method)
		// 模拟认证通过
		log.Println("[认证拦截器] 认证通过 ✓")
		return handler()
	}

	// 模拟 RPC 调用
	actualHandler := func() (interface{}, error) {
		log.Println("[Handler] 处理业务逻辑...")
		return &User{ID: 1, Name: "Go 开发者"}, nil
	}

	// 链式调用拦截器
	resp, _ := loggingInterceptor("GetUser", nil, func() (interface{}, error) {
		return authInterceptor("GetUser", nil, actualHandler)
	})

	user := resp.(*User)
	log.Printf("[结果] 用户: ID=%d, Name=%s", user.ID, user.Name)
	fmt.Println()
}

// ============================================================
// Proto 文件示例说明
// ============================================================

func showProtoExample() {
	fmt.Println(repeatStr("=", 60))
	fmt.Println("📄 附录：.proto 文件示例")
	fmt.Println(repeatStr("=", 60))

	protoExample := `
// user.proto — gRPC 服务定义文件
syntax = "proto3";

package user;
option go_package = "./pb";

// 用户服务定义
service UserService {
    // Unary：获取单个用户
    rpc GetUser(GetUserRequest) returns (User);
    
    // Server Streaming：获取用户列表（流式返回）
    rpc ListUsers(ListUsersRequest) returns (stream User);
    
    // Client Streaming：批量创建用户
    rpc BatchCreateUsers(stream CreateUserRequest) returns (BatchCreateResponse);
    
    // Bidirectional Streaming：聊天
    rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}

message User {
    int64 id = 1;
    string name = 2;
    string email = 3;
}

message GetUserRequest {
    int64 id = 1;
}

message ListUsersRequest {
    string prefix = 1;
    int32 page_size = 2;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
}

message BatchCreateResponse {
    int32 created_count = 1;
}

message ChatMessage {
    string from = 1;
    string content = 2;
    int64 timestamp = 3;
}

// 编译命令：
// protoc --go_out=. --go-grpc_out=. user.proto
`
	fmt.Println(protoExample)
}

// ============================================================
// 工具函数
// ============================================================

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println()
	fmt.Println("🔧 gRPC 四种通信模式演示")
	fmt.Println("本示例演示 gRPC 的核心概念，包括四种通信模式和拦截器。")
	fmt.Println()

	// 演示四种通信模式（模拟）
	demoUnary()
	demoServerStreaming()
	demoClientStreaming()
	demoBidirectionalStreaming()

	// 演示真实 gRPC 连接
	demoRealGRPC()

	// 演示拦截器概念
	demoInterceptor()

	// 展示 proto 文件示例
	showProtoExample()

	fmt.Println("✅ 所有演示完成！")
	fmt.Println()
	fmt.Println("📚 学习建议：")
	fmt.Println("  1. 理解四种通信模式的适用场景")
	fmt.Println("  2. 实际项目中使用 .proto 文件定义服务")
	fmt.Println("  3. 使用 protoc 生成 Go 代码")
	fmt.Println("  4. 学习拦截器实现日志、认证、限流等横切关注点")

}
