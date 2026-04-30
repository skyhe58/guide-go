// Gin JWT 鉴权中间件 — 完整链路演示
// 演示：JWT 认证中间件、RBAC 权限中间件、路由级别权限控制
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./gin-auth-middleware/
//
// 说明：使用 httptest 模拟完整的请求链路，无需启动真实 HTTP 服务器。
//       演示 注册 → 登录 → 访问受保护路由 → 权限校验 的完整流程。
//
// 依赖：github.com/gin-gonic/gin, github.com/golang-jwt/jwt/v5

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// 配置与数据模型
// ============================================================

const jwtSecret = "gin-auth-middleware-secret-2025"

// User 用户模型（内存存储，生产环境用数据库）
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // 不序列化密码（生产环境用 bcrypt）
	Role     string `json:"role"` // admin / author / reader
}

// UserStore 内存用户存储
type UserStore struct {
	mu      sync.RWMutex
	users   map[string]*User // username -> User
	nextID  int64
}

// NewUserStore 创建用户存储
func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[string]*User),
		nextID: 1,
	}
}

// Register 注册用户
func (s *UserStore) Register(username, password, role string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return nil, fmt.Errorf("用户名 %s 已存在", username)
	}

	user := &User{
		ID:       s.nextID,
		Username: username,
		Password: password, // 生产环境应使用 bcrypt 加密
		Role:     role,
	}
	s.users[username] = user
	s.nextID++
	return user, nil
}

// Authenticate 验证用户名密码
func (s *UserStore) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.Password != password { // 生产环境应使用 bcrypt.CompareHashAndPassword
		return nil, fmt.Errorf("密码错误")
	}
	return user, nil
}

// ============================================================
// JWT 工具函数
// ============================================================

// Claims 自定义 JWT Claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// generateToken 生成 JWT Token
func generateToken(user *User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-auth-demo",
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// parseToken 解析 JWT Token
func parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("无效的 Token")
	}
	return claims, nil
}

// ============================================================
// Gin 中间件
// ============================================================

// JWTAuthMiddleware JWT 认证中间件
// 从 Authorization Header 提取并验证 JWT Token
// 验证通过后将用户信息写入 Gin Context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少 Authorization Header",
			})
			c.Abort()
			return
		}

		// 2. 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization 格式错误，应为 Bearer <token>",
			})
			c.Abort()
			return
		}

		// 3. 解析并验证 Token
		claims, err := parseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": fmt.Sprintf("Token 验证失败: %v", err),
			})
			c.Abort()
			return
		}

		// 4. 将用户信息写入 Context，供后续 Handler 使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RBACMiddleware RBAC 权限中间件
// 检查用户角色是否在允许的角色列表中
// 角色层级：admin > author > reader
func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	// 角色层级映射（数字越大权限越高）
	roleLevel := map[string]int{
		"reader": 1,
		"author": 2,
		"admin":  3,
	}

	return func(c *gin.Context) {
		// 从 Context 获取用户角色（由 JWT 中间件写入）
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无法获取用户角色",
			})
			c.Abort()
			return
		}

		userRole := role.(string)
		userLevel := roleLevel[userRole]

		// 检查用户角色是否满足要求（支持角色继承）
		allowed := false
		for _, r := range allowedRoles {
			if userLevel >= roleLevel[r] {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": fmt.Sprintf("权限不足：需要 %v 角色，当前角色为 %s", allowedRoles, userRole),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================================================
// HTTP Handlers
// ============================================================

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func setupRouter(store *UserStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// --- 公开路由（无需认证） ---
	public := r.Group("/api")
	{
		// 注册
		public.POST("/register", func(c *gin.Context) {
			var req RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
				return
			}

			user, err := store.Register(req.Username, req.Password, req.Role)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"code":    201,
				"message": "注册成功",
				"data":    user,
			})
		})

		// 登录
		public.POST("/login", func(c *gin.Context) {
			var req LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
				return
			}

			user, err := store.Authenticate(req.Username, req.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
				return
			}

			token, err := generateToken(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Token 生成失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "登录成功",
				"data": gin.H{
					"token": token,
					"user":  user,
				},
			})
		})
	}

	// --- 需要认证的路由 ---
	auth := r.Group("/api")
	auth.Use(JWTAuthMiddleware())
	{
		// 获取当前用户信息（所有已认证用户）
		auth.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "获取个人信息成功",
				"data": gin.H{
					"user_id":  c.MustGet("user_id"),
					"username": c.MustGet("username"),
					"role":     c.MustGet("role"),
				},
			})
		})

		// 创建文章（需要 author 或 admin 角色）
		auth.POST("/articles", RBACMiddleware("author"), func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"code":    201,
				"message": fmt.Sprintf("文章创建成功（操作者：%s，角色：%s）", c.MustGet("username"), c.MustGet("role")),
			})
		})

		// 获取文章列表（所有已认证用户）
		auth.GET("/articles", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "获取文章列表成功",
				"data":    []string{"Go 并发编程", "JWT 认证实战", "Gin 中间件详解"},
			})
		})
	}

	// --- 管理员路由 ---
	admin := r.Group("/api/admin")
	admin.Use(JWTAuthMiddleware(), RBACMiddleware("admin"))
	{
		// 用户管理（仅 admin）
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "获取用户列表成功（管理员操作）",
			})
		})

		// 删除用户（仅 admin）
		admin.DELETE("/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": fmt.Sprintf("用户 %s 已删除（管理员操作）", c.Param("id")),
			})
		})
	}

	return r
}

// ============================================================
// 测试辅助函数
// ============================================================

// doRequest 发送 HTTP 请求并返回响应
func doRequest(router *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// printResponse 格式化打印响应
func printResponse(label string, w *httptest.ResponseRecorder) {
	var result map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	resultJSON, _ := json.MarshalIndent(result, "    ", "  ")
	fmt.Printf("  [%d] %s\n    %s\n", w.Code, label, string(resultJSON))
}

// ============================================================
// 演示入口
// ============================================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Gin JWT 鉴权中间件 — 完整链路演示")
	fmt.Println(strings.Repeat("=", 60))

	store := NewUserStore()
	router := setupRouter(store)

	// --- 1. 注册用户 ---
	fmt.Println("\n--- 1. 注册用户（公开路由） ---")

	w := doRequest(router, "POST", "/api/register", RegisterRequest{
		Username: "admin_user", Password: "admin123", Role: "admin",
	}, "")
	printResponse("注册管理员", w)

	w = doRequest(router, "POST", "/api/register", RegisterRequest{
		Username: "author_user", Password: "author123", Role: "author",
	}, "")
	printResponse("注册作者", w)

	w = doRequest(router, "POST", "/api/register", RegisterRequest{
		Username: "reader_user", Password: "reader123", Role: "reader",
	}, "")
	printResponse("注册读者", w)

	// --- 2. 用户登录 ---
	fmt.Println("\n--- 2. 用户登录（获取 JWT Token） ---")

	// 管理员登录
	w = doRequest(router, "POST", "/api/login", LoginRequest{
		Username: "admin_user", Password: "admin123",
	}, "")
	var adminLoginResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &adminLoginResp)
	adminToken := adminLoginResp["data"].(map[string]interface{})["token"].(string)
	printResponse("管理员登录", w)

	// 作者登录
	w = doRequest(router, "POST", "/api/login", LoginRequest{
		Username: "author_user", Password: "author123",
	}, "")
	var authorLoginResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &authorLoginResp)
	authorToken := authorLoginResp["data"].(map[string]interface{})["token"].(string)
	printResponse("作者登录", w)

	// 读者登录
	w = doRequest(router, "POST", "/api/login", LoginRequest{
		Username: "reader_user", Password: "reader123",
	}, "")
	var readerLoginResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &readerLoginResp)
	readerToken := readerLoginResp["data"].(map[string]interface{})["token"].(string)
	printResponse("读者登录", w)

	// --- 3. 访问受保护路由 ---
	fmt.Println("\n--- 3. 访问受保护路由（JWT 认证） ---")

	// 无 Token 访问
	w = doRequest(router, "GET", "/api/profile", nil, "")
	printResponse("无 Token 访问 /profile", w)

	// 有 Token 访问
	w = doRequest(router, "GET", "/api/profile", nil, adminToken)
	printResponse("管理员访问 /profile", w)

	// --- 4. RBAC 权限校验 ---
	fmt.Println("\n--- 4. RBAC 权限校验 ---")

	// 管理员创建文章（✅ admin >= author）
	w = doRequest(router, "POST", "/api/articles", map[string]string{"title": "Go 入门"}, adminToken)
	printResponse("管理员创建文章", w)

	// 作者创建文章（✅ author >= author）
	w = doRequest(router, "POST", "/api/articles", map[string]string{"title": "JWT 实战"}, authorToken)
	printResponse("作者创建文章", w)

	// 读者创建文章（❌ reader < author）
	w = doRequest(router, "POST", "/api/articles", map[string]string{"title": "我的文章"}, readerToken)
	printResponse("读者创建文章（应被拒绝）", w)

	// --- 5. 管理员专属路由 ---
	fmt.Println("\n--- 5. 管理员专属路由 ---")

	// 管理员访问用户列表（✅）
	w = doRequest(router, "GET", "/api/admin/users", nil, adminToken)
	printResponse("管理员访问用户列表", w)

	// 作者访问用户列表（❌）
	w = doRequest(router, "GET", "/api/admin/users", nil, authorToken)
	printResponse("作者访问用户列表（应被拒绝）", w)

	// 管理员删除用户（✅）
	w = doRequest(router, "DELETE", "/api/admin/users/2", nil, adminToken)
	printResponse("管理员删除用户", w)

	// --- 6. 错误场景 ---
	fmt.Println("\n--- 6. 错误场景 ---")

	// 无效 Token
	w = doRequest(router, "GET", "/api/profile", nil, "invalid.token.here")
	printResponse("无效 Token", w)

	// 错误密码登录
	w = doRequest(router, "POST", "/api/login", LoginRequest{
		Username: "admin_user", Password: "wrong_password",
	}, "")
	printResponse("错误密码登录", w)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Gin JWT 鉴权中间件演示完成")
	fmt.Println(strings.Repeat("=", 60))
}
