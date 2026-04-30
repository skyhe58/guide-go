// Go 1.22+ | 验证日期：2025-01-01
// HTTP Handler 示例
// 演示 net/http 标准库的 Handler 实现
// 配合 httptest 包进行测试
package httpdemo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// ============================================================
// 数据模型
// ============================================================

// User 用户模型
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ============================================================
// 内存存储（模拟数据库）
// ============================================================

// UserStore 用户内存存储
type UserStore struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewUserStore 创建用户存储
func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// Get 获取用户
func (s *UserStore) Get(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

// Add 添加用户
func (s *UserStore) Add(name string, age int) *User {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := &User{ID: s.nextID, Name: name, Age: age}
	s.users[s.nextID] = user
	s.nextID++
	return user
}

// ============================================================
// HTTP Handler
// ============================================================

// HealthHandler 健康检查接口
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "ok",
	})
}

// GetUserHandler 获取用户接口
// 根据 URL 查询参数 id 获取用户信息
func GetUserHandler(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Code:    400,
				Message: "缺少 id 参数",
			})
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Code:    400,
				Message: "id 参数格式错误",
			})
			return
		}

		user, ok := store.Get(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{
				Code:    404,
				Message: "用户不存在",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Code:    200,
			Message: "success",
			Data:    user,
		})
	}
}

// CreateUserHandler 创建用户接口
func CreateUserHandler(store *UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(Response{
				Code:    405,
				Message: "仅支持 POST 方法",
			})
			return
		}

		var req struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Code:    400,
				Message: "请求体格式错误",
			})
			return
		}

		if req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Code:    400,
				Message: "用户名不能为空",
			})
			return
		}

		user := store.Add(req.Name, req.Age)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response{
			Code:    201,
			Message: "创建成功",
			Data:    user,
		})
	}
}
