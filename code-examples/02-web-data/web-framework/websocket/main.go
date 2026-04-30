// WebSocket 示例
// 演示：WebSocket 服务端、Echo 回显、Hub 广播模式
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：go run main.go
// 测试：使用浏览器打开 http://localhost:8080 查看内置测试页面
//       或使用 wscat: wscat -c ws://localhost:8080/ws
//
// 依赖：github.com/gorilla/websocket

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ============================================================
// WebSocket 升级器配置
// ============================================================

// upgrader 将 HTTP 连接升级为 WebSocket 连接
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin 检查请求来源
	// 生产环境应该严格检查 Origin，防止跨站 WebSocket 劫持
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源
	},
}

// ============================================================
// Client 客户端连接管理
// ============================================================

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // 发送消息的缓冲通道
	name string
}

// readPump 读取客户端消息的 goroutine
// 每个连接一个读 goroutine，负责从 WebSocket 读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// 设置读取限制和超时
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置 Pong 处理器：收到 Pong 时重置读取超时
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				log.Printf("[WebSocket] 读取错误: %v", err)
			}
			break
		}

		// 广播消息给所有客户端
		broadcastMsg := fmt.Sprintf("[%s] %s", c.name, string(message))
		log.Printf("[WebSocket] 收到消息: %s", broadcastMsg)
		c.hub.broadcast <- []byte(broadcastMsg)
	}
}

// writePump 写入消息到客户端的 goroutine
// 每个连接一个写 goroutine，负责向 WebSocket 写入消息
// 注意：gorilla/websocket 不支持多个 goroutine 同时写入
func (c *Client) writePump() {
	// 心跳定时器：每 30 秒发送一次 Ping
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 设置写入超时
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// Hub 关闭了 send channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 写入文本消息
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将缓冲区中的消息一起发送（减少系统调用）
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 发送 Ping 心跳
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ============================================================
// Hub 消息中心（广播模式）
// ============================================================

// Hub 管理所有客户端连接和消息广播
type Hub struct {
	clients    map[*Client]bool // 已注册的客户端
	broadcast  chan []byte      // 广播消息通道
	register   chan *Client     // 注册通道
	unregister chan *Client     // 注销通道
	mu         sync.RWMutex
}

// NewHub 创建消息中心
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 运行消息中心（在独立 goroutine 中运行）
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// 新客户端注册
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("[Hub] 客户端 %s 已连接，当前在线: %d", client.name, count)

			// 通知所有人
			h.broadcast <- []byte(fmt.Sprintf("📢 %s 加入了聊天室（在线人数: %d）", client.name, count))

		case client := <-h.unregister:
			// 客户端注销
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("[Hub] 客户端 %s 已断开，当前在线: %d", client.name, count)

			h.broadcast <- []byte(fmt.Sprintf("📢 %s 离开了聊天室（在线人数: %d）", client.name, count))

		case message := <-h.broadcast:
			// 广播消息给所有客户端
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 客户端发送缓冲区满，关闭连接
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ============================================================
// HTTP 处理函数
// ============================================================

var clientCounter int
var counterMu sync.Mutex

// handleWebSocket 处理 WebSocket 连接
func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 将 HTTP 连接升级为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] 升级失败: %v", err)
		return
	}

	// 分配客户端名称
	counterMu.Lock()
	clientCounter++
	name := fmt.Sprintf("用户_%d", clientCounter)
	counterMu.Unlock()

	// 创建客户端
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		name: name,
	}

	// 注册到 Hub
	hub.register <- client

	// 启动读写 goroutine
	// 注意：每个连接两个 goroutine（一读一写），这是 WebSocket 的标准模式
	go client.writePump()
	go client.readPump()
}

// handleEcho 简单的 Echo WebSocket（不使用 Hub）
func handleEcho(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Echo] 升级失败: %v", err)
		return
	}
	defer conn.Close()

	log.Println("[Echo] 新的 Echo 连接")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Echo] 读取错误: %v", err)
			break
		}

		log.Printf("[Echo] 收到: %s", string(message))

		// 回显消息
		reply := fmt.Sprintf("Echo: %s", string(message))
		if err := conn.WriteMessage(messageType, []byte(reply)); err != nil {
			log.Printf("[Echo] 写入错误: %v", err)
			break
		}
	}
}

// handleHome 提供简单的 HTML 测试页面
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, testPageHTML)
}

// ============================================================
// 主函数
// ============================================================

func main() {
	hub := NewHub()
	go hub.Run()

	mux := http.NewServeMux()

	// 路由
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})
	mux.HandleFunc("/echo", handleEcho)

	log.Println("🚀 WebSocket 服务器启动在 http://localhost:8080")
	log.Println("📋 端点:")
	log.Println("   /     - 测试页面（浏览器打开）")
	log.Println("   /ws   - 聊天室 WebSocket（Hub 广播模式）")
	log.Println("   /echo - Echo WebSocket（简单回显）")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// ============================================================
// 内置测试页面 HTML
// ============================================================

const testPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>WebSocket 测试</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 40px auto; padding: 0 20px; }
        #log { border: 1px solid #ccc; padding: 10px; height: 300px; overflow-y: auto; background: #f9f9f9; }
        input { width: 70%; padding: 8px; }
        button { padding: 8px 16px; }
        .msg { margin: 2px 0; }
    </style>
</head>
<body>
    <h2>WebSocket 聊天室测试</h2>
    <div id="log"></div>
    <br>
    <input type="text" id="msg" placeholder="输入消息..." onkeypress="if(event.key==='Enter')send()">
    <button onclick="send()">发送</button>
    <button onclick="connect()">连接</button>
    <button onclick="disconnect()">断开</button>
    <script>
        var ws;
        var logDiv = document.getElementById('log');
        function appendLog(msg) {
            var p = document.createElement('div');
            p.className = 'msg';
            p.textContent = msg;
            logDiv.appendChild(p);
            logDiv.scrollTop = logDiv.scrollHeight;
        }
        function connect() {
            if (ws) { ws.close(); }
            ws = new WebSocket('ws://' + location.host + '/ws');
            ws.onopen = function() { appendLog('✅ 已连接'); };
            ws.onclose = function() { appendLog('❌ 已断开'); };
            ws.onmessage = function(e) { appendLog(e.data); };
            ws.onerror = function(e) { appendLog('⚠️ 错误'); };
        }
        function send() {
            var input = document.getElementById('msg');
            if (ws && input.value) {
                ws.send(input.value);
                input.value = '';
            }
        }
        function disconnect() { if (ws) ws.close(); }
        connect();
    </script>
</body>
</html>`
