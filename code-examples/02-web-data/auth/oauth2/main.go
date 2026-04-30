// OAuth2 第三方登录 — GitHub 授权码模式模拟
// 演示：OAuth2 授权码流程、State 防 CSRF、Token 交换、用户信息获取
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./oauth2/
//
// 说明：本示例完整模拟 OAuth2 授权码流程，无需真实 GitHub 应用。
//       通过内存模拟 Authorization Server 和 Resource Server 的行为。

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ============================================================
// OAuth2 核心数据结构
// ============================================================

// OAuthConfig OAuth2 客户端配置（对应 golang.org/x/oauth2.Config）
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string   // 授权端点
	TokenURL     string   // Token 端点
	Scopes       []string // 请求的权限范围
}

// OAuthToken OAuth2 Token 响应
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// GitHubUser GitHub 用户信息
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

// AuthorizationCode 授权码（Authorization Server 内部存储）
type AuthorizationCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	Scope       string
	UserID      int64
	ExpiresAt   time.Time
	Used        bool // 授权码只能使用一次
}

// ============================================================
// 模拟 Authorization Server（GitHub OAuth）
// ============================================================

// MockAuthServer 模拟 GitHub 授权服务器
type MockAuthServer struct {
	mu    sync.Mutex
	codes map[string]*AuthorizationCode // code -> AuthorizationCode
	users map[int64]*GitHubUser         // 模拟用户数据库
}

// NewMockAuthServer 创建模拟授权服务器
func NewMockAuthServer() *MockAuthServer {
	return &MockAuthServer{
		codes: make(map[string]*AuthorizationCode),
		users: map[int64]*GitHubUser{
			1001: {
				ID:        1001,
				Login:     "zhangsan",
				Name:      "张三",
				Email:     "zhangsan@example.com",
				AvatarURL: "https://avatars.example.com/zhangsan.png",
				Bio:       "Go 开发者 | 云原生爱好者",
			},
		},
	}
}

// generateCode 生成随机授权码
func generateCode() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// generateState 生成随机 State 参数（防 CSRF）
func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Authorize 模拟授权端点：用户授权后生成授权码
// 真实场景中，这是一个 HTTP 端点，用户在浏览器中完成授权
func (s *MockAuthServer) Authorize(clientID, redirectURI, scope string, userID int64) (code string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证用户存在
	if _, ok := s.users[userID]; !ok {
		return "", errors.New("用户不存在")
	}

	// 生成授权码（有效期 10 分钟，只能使用一次）
	code = generateCode()
	s.codes[code] = &AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		UserID:      userID,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Used:        false,
	}

	return code, nil
}

// ExchangeToken 模拟 Token 端点：用授权码换取 Access Token
func (s *MockAuthServer) ExchangeToken(code, clientID, clientSecret, redirectURI string) (*OAuthToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 查找授权码
	authCode, ok := s.codes[code]
	if !ok {
		return nil, errors.New("无效的授权码")
	}

	// 2. 检查授权码是否已使用（防止重放攻击）
	if authCode.Used {
		return nil, errors.New("授权码已被使用（防重放攻击）")
	}

	// 3. 检查授权码是否过期
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.New("授权码已过期")
	}

	// 4. 验证 client_id 和 redirect_uri 一致性
	if authCode.ClientID != clientID {
		return nil, errors.New("client_id 不匹配")
	}
	if authCode.RedirectURI != redirectURI {
		return nil, errors.New("redirect_uri 不匹配")
	}

	// 5. 标记授权码已使用
	authCode.Used = true

	// 6. 生成 Access Token（真实场景中可能是 JWT 或不透明字符串）
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)

	return &OAuthToken{
		AccessToken:  "gho_" + hex.EncodeToString(tokenBytes), // GitHub 风格的 Token 前缀
		TokenType:    "bearer",
		ExpiresIn:    3600,
		RefreshToken: "", // GitHub OAuth 不返回 Refresh Token
		Scope:        authCode.Scope,
	}, nil
}

// GetUserInfo 模拟 Resource Server：用 Access Token 获取用户信息
func (s *MockAuthServer) GetUserInfo(accessToken string) (*GitHubUser, error) {
	// 真实场景中需要验证 Access Token 的有效性
	// 这里简化处理，直接返回模拟用户
	if !strings.HasPrefix(accessToken, "gho_") {
		return nil, errors.New("无效的 Access Token")
	}
	return s.users[1001], nil
}

// ============================================================
// OAuth2 客户端（你的 Go 应用）
// ============================================================

// OAuthClient OAuth2 客户端
type OAuthClient struct {
	config     *OAuthConfig
	stateStore map[string]bool // 存储已生成的 state，用于验证回调
	mu         sync.Mutex
}

// NewOAuthClient 创建 OAuth2 客户端
func NewOAuthClient(config *OAuthConfig) *OAuthClient {
	return &OAuthClient{
		config:     config,
		stateStore: make(map[string]bool),
	}
}

// GetAuthorizationURL 生成授权 URL（步骤 1：重定向用户到授权服务器）
func (c *OAuthClient) GetAuthorizationURL() (authURL string, state string) {
	state = generateState()

	// 存储 state 用于后续验证
	c.mu.Lock()
	c.stateStore[state] = true
	c.mu.Unlock()

	// 构建授权 URL
	params := url.Values{
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURI},
		"scope":         {strings.Join(c.config.Scopes, " ")},
		"state":         {state},
		"response_type": {"code"},
	}

	return c.config.AuthURL + "?" + params.Encode(), state
}

// ValidateState 验证回调中的 state 参数（防 CSRF）
func (c *OAuthClient) ValidateState(state string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.stateStore[state]; ok {
		delete(c.stateStore, state) // state 只能使用一次
		return true
	}
	return false
}

// ============================================================
// 演示入口
// ============================================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("OAuth2 第三方登录 — GitHub 授权码模式模拟")
	fmt.Println(strings.Repeat("=", 60))

	// --- 初始化 ---
	authServer := NewMockAuthServer()

	config := &OAuthConfig{
		ClientID:     "Iv1.abc123def456",
		ClientSecret: "secret_xyz789",
		RedirectURI:  "http://localhost:8080/callback",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		Scopes:       []string{"user:email", "read:user"},
	}

	client := NewOAuthClient(config)

	// ============================================================
	// 步骤 1：生成授权 URL，引导用户跳转到 GitHub
	// ============================================================
	fmt.Println("\n--- 步骤 1：生成授权 URL ---")
	authURL, state := client.GetAuthorizationURL()
	fmt.Printf("授权 URL:\n  %s\n", authURL)
	fmt.Printf("State 参数: %s\n", state)
	fmt.Println("（用户在浏览器中打开此 URL，登录 GitHub 并授权）")

	// ============================================================
	// 步骤 2：用户在 GitHub 授权后，GitHub 回调应用
	// ============================================================
	fmt.Println("\n--- 步骤 2：用户授权，GitHub 回调 ---")
	fmt.Println("（模拟用户 zhangsan 在 GitHub 上点击「授权」按钮）")

	code, err := authServer.Authorize(config.ClientID, config.RedirectURI, "user:email read:user", 1001)
	if err != nil {
		fmt.Printf("授权失败: %v\n", err)
		return
	}

	callbackURL := fmt.Sprintf("%s?code=%s&state=%s", config.RedirectURI, code, state)
	fmt.Printf("回调 URL:\n  %s\n", callbackURL)

	// ============================================================
	// 步骤 3：验证 State 参数（防 CSRF 攻击）
	// ============================================================
	fmt.Println("\n--- 步骤 3：验证 State 参数（防 CSRF） ---")
	if client.ValidateState(state) {
		fmt.Println("✅ State 验证通过")
	} else {
		fmt.Println("❌ State 验证失败，可能是 CSRF 攻击！")
		return
	}

	// 演示 CSRF 攻击场景
	fmt.Println("\n[CSRF 攻击模拟] 攻击者伪造 state:")
	if client.ValidateState("fake-state-from-attacker") {
		fmt.Println("❌ 不应该到这里")
	} else {
		fmt.Println("✅ 伪造的 state 被拒绝，CSRF 攻击被阻止")
	}

	// ============================================================
	// 步骤 4：用授权码换取 Access Token
	// ============================================================
	fmt.Println("\n--- 步骤 4：用授权码换取 Access Token ---")
	token, err := authServer.ExchangeToken(code, config.ClientID, config.ClientSecret, config.RedirectURI)
	if err != nil {
		fmt.Printf("Token 交换失败: %v\n", err)
		return
	}

	tokenJSON, _ := json.MarshalIndent(token, "  ", "  ")
	fmt.Printf("Token 响应:\n  %s\n", string(tokenJSON))

	// 演示授权码重放攻击防护
	fmt.Println("\n[重放攻击防护] 尝试重复使用授权码:")
	_, err = authServer.ExchangeToken(code, config.ClientID, config.ClientSecret, config.RedirectURI)
	if err != nil {
		fmt.Printf("✅ 重放攻击被阻止：%v\n", err)
	}

	// ============================================================
	// 步骤 5：用 Access Token 获取用户信息
	// ============================================================
	fmt.Println("\n--- 步骤 5：用 Access Token 获取用户信息 ---")
	user, err := authServer.GetUserInfo(token.AccessToken)
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}

	userJSON, _ := json.MarshalIndent(user, "  ", "  ")
	fmt.Printf("用户信息:\n  %s\n", string(userJSON))

	// ============================================================
	// 步骤 6：创建本地用户会话
	// ============================================================
	fmt.Println("\n--- 步骤 6：创建本地用户会话 ---")
	fmt.Printf("GitHub 用户 %s（%s）登录成功\n", user.Login, user.Name)
	fmt.Println("应用可以：")
	fmt.Println("  1. 在本地数据库创建/更新用户记录")
	fmt.Println("  2. 签发应用自己的 JWT Token")
	fmt.Println("  3. 将 GitHub access_token 存储用于后续 API 调用")

	// ============================================================
	// 错误场景演示
	// ============================================================
	fmt.Println("\n--- 错误场景演示 ---")

	// 无效的 Access Token
	fmt.Println("\n[错误] 无效的 Access Token:")
	_, err = authServer.GetUserInfo("invalid_token")
	if err != nil {
		fmt.Printf("✅ 预期错误：%v\n", err)
	}

	// client_id 不匹配
	fmt.Println("\n[错误] client_id 不匹配:")
	code2, _ := authServer.Authorize(config.ClientID, config.RedirectURI, "user:email", 1001)
	_, err = authServer.ExchangeToken(code2, "wrong-client-id", config.ClientSecret, config.RedirectURI)
	if err != nil {
		fmt.Printf("✅ 预期错误：%v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("OAuth2 授权码模式演示完成")
	fmt.Println(strings.Repeat("=", 60))
}
