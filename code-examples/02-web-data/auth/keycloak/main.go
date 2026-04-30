// Keycloak 集成 — OIDC Token 验证 + 管理 API 调用
// 演示：JWKS 公钥验证、Token 解析、Keycloak REST API
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯 Go 模拟 OIDC/JWKS Token 验证（直接运行）
// Part B：连接真实 Keycloak（需传入参数 'real'）
//
// 运行方式：
//   go run ./keycloak/              # Part A：模拟 OIDC 验证
//   go run ./keycloak/ real         # Part B：连接真实 Keycloak
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.auth.yml up -d
//   管理控制台：http://localhost:8080，用户名：admin，密码：admin
//   需要先在 Keycloak 中创建 Realm 和 Client

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ============================================================
// Part A：模拟 OIDC/JWKS Token 验证
// ============================================================

// JWK JSON Web Key 结构（RSA 公钥）
type JWK struct {
	Kty string `json:"kty"` // 密钥类型：RSA
	Use string `json:"use"` // 用途：sig（签名）
	Kid string `json:"kid"` // 密钥 ID
	N   string `json:"n"`   // RSA 模数（Base64url）
	E   string `json:"e"`   // RSA 指数（Base64url）
	Alg string `json:"alg"` // 算法：RS256
}

// JWKS JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// KeycloakTokenClaims Keycloak JWT Token 的 Claims
type KeycloakTokenClaims struct {
	Exp           int64    `json:"exp"`
	Iat           int64    `json:"iat"`
	Iss           string   `json:"iss"`
	Sub           string   `json:"sub"`
	Aud           string   `json:"aud"`
	Typ           string   `json:"typ"`
	Azp           string   `json:"azp"`
	PreferredUser string   `json:"preferred_username"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	RealmAccess   struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// base64urlEncode Base64url 编码（无填充）
func base64urlEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// base64urlDecode Base64url 解码
func base64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// MockOIDCProvider 模拟 OIDC Provider（Keycloak）
type MockOIDCProvider struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
	issuer     string
}

// NewMockOIDCProvider 创建模拟 OIDC Provider
func NewMockOIDCProvider() (*MockOIDCProvider, error) {
	// 生成 RSA 密钥对（Keycloak 使用 RS256 签名）
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成 RSA 密钥失败: %w", err)
	}

	return &MockOIDCProvider{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		kid:        "keycloak-key-001",
		issuer:     "http://localhost:8080/realms/my-app",
	}, nil
}

// GetJWKS 返回 JWKS（公钥集合），对应 Keycloak 的 /protocol/openid-connect/certs 端点
func (p *MockOIDCProvider) GetJWKS() *JWKS {
	return &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: p.kid,
				N:   base64urlEncode(p.publicKey.N.Bytes()),
				E:   base64urlEncode(big.NewInt(int64(p.publicKey.E)).Bytes()),
				Alg: "RS256",
			},
		},
	}
}

// IssueToken 模拟 Keycloak 签发 JWT Token（RS256 签名）
func (p *MockOIDCProvider) IssueToken(sub, username, email, name string, roles []string) (string, error) {
	now := time.Now()

	// 构建 Claims
	claims := KeycloakTokenClaims{
		Exp:           now.Add(5 * time.Minute).Unix(),
		Iat:           now.Unix(),
		Iss:           p.issuer,
		Sub:           sub,
		Aud:           "go-app",
		Typ:           "Bearer",
		Azp:           "go-app",
		PreferredUser: username,
		Email:         email,
		Name:          name,
	}
	claims.RealmAccess.Roles = roles

	// 构建 JWT Header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": p.kid,
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	// Base64url 编码
	headerB64 := base64urlEncode(headerJSON)
	claimsB64 := base64urlEncode(claimsJSON)

	// 签名内容
	signingInput := headerB64 + "." + claimsB64

	// RS256 签名：SHA256 + RSA PKCS1v15
	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, p.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	signatureB64 := base64urlEncode(signature)

	return signingInput + "." + signatureB64, nil
}

// VerifyToken 使用 JWKS 公钥验证 Token 签名
func VerifyTokenWithJWKS(tokenString string, jwks *JWKS) (*KeycloakTokenClaims, error) {
	// 1. 分割 Token
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("无效的 JWT 格式：期望 3 部分，实际 %d 部分", len(parts))
	}

	// 2. 解析 Header，获取 kid
	headerJSON, err := base64urlDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("解码 Header 失败: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("解析 Header 失败: %w", err)
	}

	// 3. 根据 kid 查找对应的公钥
	var matchedKey *JWK
	for i := range jwks.Keys {
		if jwks.Keys[i].Kid == header.Kid {
			matchedKey = &jwks.Keys[i]
			break
		}
	}
	if matchedKey == nil {
		return nil, fmt.Errorf("未找到 kid=%s 对应的公钥", header.Kid)
	}

	// 4. 从 JWK 重建 RSA 公钥
	nBytes, err := base64urlDecode(matchedKey.N)
	if err != nil {
		return nil, fmt.Errorf("解码公钥 N 失败: %w", err)
	}
	eBytes, err := base64urlDecode(matchedKey.E)
	if err != nil {
		return nil, fmt.Errorf("解码公钥 E 失败: %w", err)
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}

	// 5. 验证签名
	signingInput := parts[0] + "." + parts[1]
	signatureBytes, err := base64urlDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("解码签名失败: %w", err)
	}

	hash := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signatureBytes); err != nil {
		return nil, fmt.Errorf("签名验证失败: %w", err)
	}

	// 6. 解析 Claims
	claimsJSON, err := base64urlDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("解码 Claims 失败: %w", err)
	}

	var claims KeycloakTokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("解析 Claims 失败: %w", err)
	}

	// 7. 检查过期时间
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("Token 已过期")
	}

	return &claims, nil
}

func partA() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Part A：OIDC/JWKS Token 验证（模拟 Keycloak）")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 创建模拟 OIDC Provider
	fmt.Println("\n--- 1. 初始化 OIDC Provider（模拟 Keycloak） ---")
	provider, err := NewMockOIDCProvider()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}
	fmt.Printf("Issuer: %s\n", provider.issuer)
	fmt.Printf("Key ID: %s\n", provider.kid)

	// 2. 获取 JWKS 公钥集合
	fmt.Println("\n--- 2. 获取 JWKS 公钥集合 ---")
	jwks := provider.GetJWKS()
	jwksJSON, _ := json.MarshalIndent(jwks, "  ", "  ")
	fmt.Printf("JWKS 端点响应:\n  %s\n", string(jwksJSON))

	// 3. 签发 Token（模拟用户登录 Keycloak）
	fmt.Println("\n--- 3. Keycloak 签发 Token ---")
	token, err := provider.IssueToken(
		"user-uuid-001",
		"zhangsan",
		"zhangsan@example.com",
		"张三",
		[]string{"admin", "user"},
	)
	if err != nil {
		fmt.Printf("签发失败: %v\n", err)
		return
	}
	fmt.Printf("JWT Token:\n  %s...（截取前 80 字符）\n", token[:80])

	// 4. 使用 JWKS 公钥验证 Token
	fmt.Println("\n--- 4. Go 服务验证 Token（使用 JWKS 公钥） ---")
	claims, err := VerifyTokenWithJWKS(token, jwks)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
		return
	}

	fmt.Println("✅ Token 签名验证通过")
	fmt.Printf("  用户 ID:    %s\n", claims.Sub)
	fmt.Printf("  用户名:     %s\n", claims.PreferredUser)
	fmt.Printf("  邮箱:       %s\n", claims.Email)
	fmt.Printf("  姓名:       %s\n", claims.Name)
	fmt.Printf("  签发者:     %s\n", claims.Iss)
	fmt.Printf("  受众:       %s\n", claims.Aud)
	fmt.Printf("  Realm 角色: %v\n", claims.RealmAccess.Roles)

	// 5. 篡改 Token 验证
	fmt.Println("\n--- 5. 篡改 Token 检测 ---")
	tamperedToken := token[:len(token)-10] + "XXXXXXXXXX"
	_, err = VerifyTokenWithJWKS(tamperedToken, jwks)
	if err != nil {
		fmt.Printf("✅ 篡改检测成功：%v\n", err)
	}
}

// ============================================================
// Part B：连接真实 Keycloak
// ============================================================

// KeycloakConfig Keycloak 连接配置
type KeycloakConfig struct {
	BaseURL      string // http://localhost:8080
	Realm        string // my-app
	ClientID     string // go-app
	ClientSecret string // 从 Keycloak 管理控制台获取
	AdminUser    string // admin
	AdminPass    string // admin
}

// KeycloakClient Keycloak HTTP 客户端
type KeycloakClient struct {
	config     *KeycloakConfig
	httpClient *http.Client
}

// NewKeycloakClient 创建 Keycloak 客户端
func NewKeycloakClient(config *KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAdminToken 获取管理员 Token（用于调用管理 API）
func (c *KeycloakClient) GetAdminToken() (string, error) {
	tokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", c.config.BaseURL)

	data := url.Values{
		"grant_type": {"password"},
		"client_id":  {"admin-cli"},
		"username":   {c.config.AdminUser},
		"password":   {c.config.AdminPass},
	}

	resp, err := c.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("请求管理员 Token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取管理员 Token 失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 Token 响应失败: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// GetJWKS 获取 Realm 的 JWKS 公钥集合
func (c *KeycloakClient) GetJWKS() (*JWKS, error) {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.config.BaseURL, c.config.Realm)

	resp, err := c.httpClient.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("请求 JWKS 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 JWKS 失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("解析 JWKS 失败: %w", err)
	}

	return &jwks, nil
}

// ListUsers 列出 Realm 中的用户（管理 API）
func (c *KeycloakClient) ListUsers(adminToken string) ([]map[string]interface{}, error) {
	usersURL := fmt.Sprintf("%s/admin/realms/%s/users", c.config.BaseURL, c.config.Realm)

	req, _ := http.NewRequest("GET", usersURL, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户列表失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("解析用户列表失败: %w", err)
	}

	return users, nil
}

func partB() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 Keycloak")
	fmt.Println(strings.Repeat("=", 60))

	config := &KeycloakConfig{
		BaseURL:   "http://localhost:8080",
		Realm:     "master",
		AdminUser: "admin",
		AdminPass: "admin",
	}

	client := NewKeycloakClient(config)

	// 1. 获取管理员 Token
	fmt.Println("\n--- 1. 获取管理员 Token ---")
	adminToken, err := client.GetAdminToken()
	if err != nil {
		fmt.Printf("❌ 获取管理员 Token 失败: %v\n", err)
		fmt.Println("请确保 Keycloak 已启动：docker compose -f docker/docker-compose.auth.yml up -d")
		return
	}
	fmt.Printf("✅ 管理员 Token: %s...（截取前 50 字符）\n", adminToken[:50])

	// 2. 获取 JWKS 公钥
	fmt.Println("\n--- 2. 获取 JWKS 公钥集合 ---")
	jwks, err := client.GetJWKS()
	if err != nil {
		fmt.Printf("❌ 获取 JWKS 失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 获取到 %d 个公钥\n", len(jwks.Keys))
	for _, key := range jwks.Keys {
		fmt.Printf("  - Kid: %s, Alg: %s, Use: %s\n", key.Kid, key.Alg, key.Use)
	}

	// 3. 列出用户
	fmt.Println("\n--- 3. 列出 Realm 用户（管理 API） ---")
	users, err := client.ListUsers(adminToken)
	if err != nil {
		fmt.Printf("❌ 获取用户列表失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 共 %d 个用户\n", len(users))
	for _, u := range users {
		fmt.Printf("  - 用户名: %v, 邮箱: %v\n", u["username"], u["email"])
	}
}

// ============================================================
// 入口
// ============================================================

func main() {
	partA()

	// Part B：连接真实 Keycloak，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		fmt.Println()
		partB()
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Keycloak 集成演示完成")
	fmt.Println(strings.Repeat("=", 60))
}
