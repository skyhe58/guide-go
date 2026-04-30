// Go 1.22+ | 验证日期：2025-01-01
// HTTP 测试示例
// 演示 net/http/httptest 包的使用
// 包含：httptest.NewRecorder 测试 Handler、httptest.NewServer 测试客户端
package httpdemo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================
// httptest.NewRecorder 测试（直接调用 Handler）
// ============================================================

// TestHealthHandler 测试健康检查接口
// 使用 httptest.NewRecorder 直接调用 Handler，不启动 HTTP 服务器
func TestHealthHandler(t *testing.T) {
	// 创建请求
	req := httptest.NewRequest("GET", "/health", nil)
	// 创建响应记录器
	w := httptest.NewRecorder()

	// 调用 Handler
	HealthHandler(w, req)

	// 验证状态码
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 %d", resp.StatusCode, http.StatusOK)
	}

	// 验证响应体
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if result.Code != 200 {
		t.Errorf("响应 code = %d, 期望 200", result.Code)
	}
	if result.Message != "ok" {
		t.Errorf("响应 message = %s, 期望 ok", result.Message)
	}
}

// TestGetUserHandler 表驱动测试获取用户接口
func TestGetUserHandler(t *testing.T) {
	// 准备测试数据
	store := NewUserStore()
	store.Add("Alice", 30)

	handler := GetUserHandler(store)

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCode   int
	}{
		{"正常获取", "?id=1", http.StatusOK, 200},
		{"用户不存在", "?id=999", http.StatusNotFound, 404},
		{"缺少 id", "", http.StatusBadRequest, 400},
		{"id 格式错误", "?id=abc", http.StatusBadRequest, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/user"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// 验证状态码
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("状态码 = %d, 期望 %d", resp.StatusCode, tt.wantStatus)
			}

			// 验证响应 code
			var result Response
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}
			if result.Code != tt.wantCode {
				t.Errorf("响应 code = %d, 期望 %d", result.Code, tt.wantCode)
			}
		})
	}
}

// TestCreateUserHandler 测试创建用户接口
func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantCode   int
	}{
		{
			name:       "正常创建",
			method:     "POST",
			body:       `{"name":"Bob","age":25}`,
			wantStatus: http.StatusCreated,
			wantCode:   201,
		},
		{
			name:       "空用户名",
			method:     "POST",
			body:       `{"name":"","age":25}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   400,
		},
		{
			name:       "错误的请求方法",
			method:     "GET",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   405,
		},
		{
			name:       "无效 JSON",
			method:     "POST",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewUserStore()
			handler := CreateUserHandler(store)

			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/user", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("状态码 = %d, 期望 %d", resp.StatusCode, tt.wantStatus)
			}

			var result Response
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}
			if result.Code != tt.wantCode {
				t.Errorf("响应 code = %d, 期望 %d", result.Code, tt.wantCode)
			}
		})
	}
}

// ============================================================
// httptest.NewServer 测试（启动测试服务器）
// ============================================================

// TestWithServer 使用 httptest.NewServer 测试完整 HTTP 交互
// 适合测试 HTTP 客户端代码或需要完整 HTTP 栈的场景
func TestWithServer(t *testing.T) {
	store := NewUserStore()
	store.Add("Alice", 30)

	// 创建路由
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/user", GetUserHandler(store))

	// 启动测试服务器
	server := httptest.NewServer(mux)
	defer server.Close()

	// 测试健康检查
	t.Run("健康检查", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("状态码 = %d, 期望 %d", resp.StatusCode, http.StatusOK)
		}
	})

	// 测试获取用户
	t.Run("获取用户", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/user?id=1")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("状态码 = %d, 期望 %d", resp.StatusCode, http.StatusOK)
		}

		var result Response
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if result.Code != 200 {
			t.Errorf("响应 code = %d, 期望 200", result.Code)
		}
	})
}
