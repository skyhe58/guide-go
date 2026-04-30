// Gin REST API 完整示例
// 演示：路由分组、中间件、参数绑定与验证、统一错误处理、Swagger 注解
// Go 1.22+ | 验证日期 2025-01-01
//
// 依赖：github.com/gin-gonic/gin
// 运行方式：go run main.go
// 测试：
//   curl http://localhost:8080/api/v1/users
//   curl http://localhost:8080/api/v1/users/1
//   curl -X POST -H "Content-Type: application/json" \
//        -d '{"name":"Go开发者","email":"go@example.com","age":25}' \
//        http://localhost:8080/api/v1/users
//   curl -X PUT -H "Content-Type: application/json" \
//        -d '{"name":"新名字"}' \
//        http://localhost:8080/api/v1/users/1
//   curl -X DELETE http://localhost:8080/api/v1/users/1

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 数据模型与请求/响应结构
// ============================================================

// User 用户模型
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserReq 创建用户请求
// binding 标签用于参数验证（基于 go-playground/validator）
// @Description 创建用户请求体
type CreateUserReq struct {
	Name  string `json:"name" binding:"required,min=2,max=50"`  // 必填，2-50 字符
	Email string `json:"email" binding:"required,email"`         // 必填，邮箱格式
	Age   int    `json:"age" binding:"required,gte=1,lte=150"`   // 必填，1-150
}

// UpdateUserReq 更新用户请求
type UpdateUserReq struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=50"` // 可选，2-50 字符
	Email string `json:"email" binding:"omitempty,email"`        // 可选，邮箱格式
	Age   int    `json:"age" binding:"omitempty,gte=1,lte=150"`  // 可选，1-150
}

// APIResponse 统一 API 响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Items interface{} `json:"items"`
}

// ============================================================
// 内存数据存储（模拟数据库）
// ============================================================

// UserStore 线程安全的用户存储
type UserStore struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewUserStore 创建用户存储并初始化示例数据
func NewUserStore() *UserStore {
	store := &UserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}
	// 初始化示例数据
	store.Create("张三", "zhangsan@example.com", 28)
	store.Create("李四", "lisi@example.com", 32)
	store.Create("王五", "wangwu@example.com", 25)
	return store
}

func (s *UserStore) Create(name, email string, age int) *User {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := &User{
		ID:        s.nextID,
		Name:      name,
		Email:     email,
		Age:       age,
		CreatedAt: time.Now(),
	}
	s.users[s.nextID] = user
	s.nextID++
	return user
}

func (s *UserStore) GetByID(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

func (s *UserStore) List() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *UserStore) Update(id int, name, email string, age int) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return nil, false
	}
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	if age > 0 {
		user.Age = age
	}
	return user, true
}

func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return false
	}
	delete(s.users, id)
	return true
}

// ============================================================
// 自定义中间件
// ============================================================

// RequestLoggerMiddleware 请求日志中间件
// 记录每个请求的方法、路径、状态码、耗时和客户端 IP
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 执行后续中间件和 Handler
		c.Next()

		// 后处理：记录日志
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		log.Printf("[GIN] %s | %3d | %13v | %15s | %s %s",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// ErrorHandlerMiddleware 统一错误处理中间件
// 捕获 Handler 中通过 c.Error() 添加的错误，返回统一格式
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			// 如果还没有写入响应，返回错误
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Code:    500,
					Message: err.Error(),
				})
			}
		}
	}
}

// ============================================================
// 路由处理函数（Handler）
// ============================================================

// @Summary 获取用户列表
// @Description 获取所有用户，支持分页
// @Tags users
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} APIResponse
// @Router /api/v1/users [get]
func listUsers(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数（带默认值）
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 10
		}

		users := store.List()

		c.JSON(http.StatusOK, APIResponse{
			Code:    200,
			Message: "获取用户列表成功",
			Data: PaginatedResponse{
				Total: len(users),
				Page:  page,
				Size:  size,
				Items: users,
			},
		})
	}
}

// @Summary 获取用户详情
// @Description 根据 ID 获取用户信息
// @Tags users
// @Param id path int true "用户 ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/users/{id} [get]
func getUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取路径参数并转换类型
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "无效的用户 ID",
			})
			return
		}

		user, ok := store.GetByID(id)
		if !ok {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: fmt.Sprintf("用户 %d 不存在", id),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Code:    200,
			Message: "获取用户成功",
			Data:    user,
		})
	}
}

// @Summary 创建用户
// @Description 创建新用户
// @Tags users
// @Accept json
// @Param user body CreateUserReq true "用户信息"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Router /api/v1/users [post]
func createUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserReq

		// ShouldBindJSON 自动绑定 JSON 并验证
		// 如果验证失败，返回详细的错误信息
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "参数验证失败: " + err.Error(),
			})
			return // 重要：c.JSON 后必须 return，否则会继续执行
		}

		user := store.Create(req.Name, req.Email, req.Age)

		c.JSON(http.StatusCreated, APIResponse{
			Code:    201,
			Message: "创建用户成功",
			Data:    user,
		})
	}
}

// @Summary 更新用户
// @Description 更新用户信息
// @Tags users
// @Accept json
// @Param id path int true "用户 ID"
// @Param user body UpdateUserReq true "更新信息"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/users/{id} [put]
func updateUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "无效的用户 ID",
			})
			return
		}

		var req UpdateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "参数验证失败: " + err.Error(),
			})
			return
		}

		user, ok := store.Update(id, req.Name, req.Email, req.Age)
		if !ok {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: fmt.Sprintf("用户 %d 不存在", id),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Code:    200,
			Message: "更新用户成功",
			Data:    user,
		})
	}
}

// @Summary 删除用户
// @Description 根据 ID 删除用户
// @Tags users
// @Param id path int true "用户 ID"
// @Success 200 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Router /api/v1/users/{id} [delete]
func deleteUser(store *UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Code:    400,
				Message: "无效的用户 ID",
			})
			return
		}

		if !store.Delete(id) {
			c.JSON(http.StatusNotFound, APIResponse{
				Code:    404,
				Message: fmt.Sprintf("用户 %d 不存在", id),
			})
			return
		}

		c.JSON(http.StatusOK, APIResponse{
			Code:    200,
			Message: "删除用户成功",
		})
	}
}

// handleHealth 健康检查
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "服务运行正常",
		Data: gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

// ============================================================
// 主函数：配置路由和启动服务
// ============================================================

func main() {
	// 创建数据存储
	store := NewUserStore()

	// 创建 Gin 引擎
	// gin.New() 不带默认中间件，适合自定义中间件链
	// gin.Default() 自带 Logger 和 Recovery，适合快速开发
	r := gin.New()

	// 注册全局中间件（执行顺序：Recovery → ErrorHandler → RequestLogger）
	r.Use(gin.Recovery())             // panic 恢复（Gin 内置）
	r.Use(ErrorHandlerMiddleware())   // 统一错误处理
	r.Use(RequestLoggerMiddleware())  // 请求日志

	// 健康检查（不需要认证）
	r.GET("/health", handleHealth)

	// API v1 路由分组
	v1 := r.Group("/api/v1")
	{
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.GET("", listUsers(store))          // GET /api/v1/users
			users.GET("/:id", getUser(store))        // GET /api/v1/users/:id
			users.POST("", createUser(store))        // POST /api/v1/users
			users.PUT("/:id", updateUser(store))     // PUT /api/v1/users/:id
			users.DELETE("/:id", deleteUser(store))  // DELETE /api/v1/users/:id
		}
	}

	// 启动服务器
	log.Println("🚀 Gin REST API 服务器启动在 http://localhost:8080")
	log.Println("📋 可用端点:")
	log.Println("   GET    /health              - 健康检查")
	log.Println("   GET    /api/v1/users        - 获取用户列表")
	log.Println("   GET    /api/v1/users/:id    - 获取用户详情")
	log.Println("   POST   /api/v1/users        - 创建用户")
	log.Println("   PUT    /api/v1/users/:id    - 更新用户")
	log.Println("   DELETE /api/v1/users/:id    - 删除用户")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
