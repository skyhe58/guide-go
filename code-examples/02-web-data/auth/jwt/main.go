// JWT 签发与验证 — Access Token + Refresh Token 双令牌机制
// 演示：Token 生成、解析、刷新、黑名单
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./jwt/
//
// 依赖：github.com/golang-jwt/jwt/v5

package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// 配置常量
// ============================================================

const (
	// 签名密钥（生产环境应从环境变量或配置中心读取，绝不能硬编码）
	accessTokenSecret  = "my-access-token-secret-key-2025"
	refreshTokenSecret = "my-refresh-token-secret-key-2025"

	// Token 有效期
	accessTokenExpiry  = 15 * time.Minute // Access Token: 15 分钟
	refreshTokenExpiry = 7 * 24 * time.Hour // Refresh Token: 7 天
)

// ============================================================
// 自定义 Claims 结构体
// ============================================================

// UserClaims 自定义 JWT Claims，包含用户业务信息
type UserClaims struct {
	jwt.RegisteredClaims
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// RefreshClaims Refresh Token 的 Claims（只包含最少信息）
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID  int64  `json:"user_id"`
	TokenID string `json:"token_id"` // 唯一标识，用于黑名单
}

// ============================================================
// Token 黑名单（内存实现，生产环境用 Redis）
// ============================================================

// TokenBlacklist 基于内存的 Token 黑名单
// 生产环境应使用 Redis，key 为 token_id，TTL 为 Token 剩余有效期
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token_id -> 过期时间
}

// NewTokenBlacklist 创建黑名单实例
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

// Add 将 Token 加入黑名单
func (b *TokenBlacklist) Add(tokenID string, expiry time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[tokenID] = expiry
}

// IsBlacklisted 检查 Token 是否在黑名单中
func (b *TokenBlacklist) IsBlacklisted(tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, exists := b.tokens[tokenID]
	return exists
}

// Cleanup 清理已过期的黑名单条目（节省内存）
func (b *TokenBlacklist) Cleanup() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	cleaned := 0
	for id, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, id)
			cleaned++
		}
	}
	return cleaned
}

// ============================================================
// JWT 服务
// ============================================================

// JWTService JWT 签发与验证服务
type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	blacklist     *TokenBlacklist
}

// NewJWTService 创建 JWT 服务实例
func NewJWTService() *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessTokenSecret),
		refreshSecret: []byte(refreshTokenSecret),
		blacklist:     NewTokenBlacklist(),
	}
}

// generateTokenID 生成唯一的 Token ID
func generateTokenID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// TokenPair Access Token + Refresh Token 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access Token 过期时间（秒）
}

// GenerateTokenPair 生成 Access Token + Refresh Token 令牌对
func (s *JWTService) GenerateTokenPair(userID int64, username, role string) (*TokenPair, error) {
	now := time.Now()
	tokenID := generateTokenID()

	// --- 生成 Access Token ---
	accessClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "guide-go-auth",
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID, // jti: 关联到 Refresh Token 的 token_id
		},
		UserID:   userID,
		Username: username,
		Role:     role,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("签发 Access Token 失败: %w", err)
	}

	// --- 生成 Refresh Token ---
	refreshClaims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "guide-go-auth",
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
		UserID:  userID,
		TokenID: tokenID,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("签发 Refresh Token 失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(accessTokenExpiry.Seconds()),
	}, nil
}

// ParseAccessToken 解析并验证 Access Token
func (s *JWTService) ParseAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 安全检查：确保签名算法是 HMAC，防止算法混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("解析 Access Token 失败: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Access Token")
	}

	// 检查黑名单
	if s.blacklist.IsBlacklisted(claims.ID) {
		return nil, errors.New("Token 已被注销（在黑名单中）")
	}

	return claims, nil
}

// ParseRefreshToken 解析并验证 Refresh Token
func (s *JWTService) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("解析 Refresh Token 失败: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Refresh Token")
	}

	// 检查黑名单
	if s.blacklist.IsBlacklisted(claims.TokenID) {
		return nil, errors.New("Refresh Token 已被注销（在黑名单中）")
	}

	return claims, nil
}

// RefreshTokenPair 使用 Refresh Token 刷新令牌对
// 旧的 Refresh Token 会被加入黑名单（Refresh Token Rotation）
func (s *JWTService) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	// 1. 解析旧的 Refresh Token
	claims, err := s.ParseRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新失败: %w", err)
	}

	// 2. 将旧的 Token ID 加入黑名单（Refresh Token Rotation 安全策略）
	s.blacklist.Add(claims.TokenID, claims.ExpiresAt.Time)

	// 3. 生成新的令牌对（生产环境应从数据库查询最新的用户信息）
	return s.GenerateTokenPair(claims.UserID, "user_from_db", "author")
}

// RevokeToken 注销 Token（加入黑名单）
func (s *JWTService) RevokeToken(tokenID string, expiry time.Time) {
	s.blacklist.Add(tokenID, expiry)
}

// ============================================================
// 演示入口
// ============================================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("JWT 签发与验证 — Access Token + Refresh Token 双令牌机制")
	fmt.Println(strings.Repeat("=", 60))

	svc := NewJWTService()

	// --- 1. 模拟用户登录，签发令牌对 ---
	fmt.Println("\n--- 1. 用户登录：签发 Access Token + Refresh Token ---")
	pair, err := svc.GenerateTokenPair(1001, "zhangsan", "admin")
	if err != nil {
		fmt.Printf("签发失败: %v\n", err)
		return
	}
	fmt.Printf("Access Token:  %s...（截取前 50 字符）\n", pair.AccessToken[:50])
	fmt.Printf("Refresh Token: %s...（截取前 50 字符）\n", pair.RefreshToken[:50])
	fmt.Printf("Access Token 有效期: %d 秒\n", pair.ExpiresIn)

	// --- 2. 解析 Access Token ---
	fmt.Println("\n--- 2. 解析 Access Token（提取用户信息） ---")
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}
	fmt.Printf("用户 ID:   %d\n", claims.UserID)
	fmt.Printf("用户名:    %s\n", claims.Username)
	fmt.Printf("角色:      %s\n", claims.Role)
	fmt.Printf("签发者:    %s\n", claims.Issuer)
	fmt.Printf("过期时间:  %s\n", claims.ExpiresAt.Time.Format(time.RFC3339))

	// --- 3. 模拟 Token 刷新 ---
	fmt.Println("\n--- 3. 使用 Refresh Token 刷新令牌对 ---")
	newPair, err := svc.RefreshTokenPair(pair.RefreshToken)
	if err != nil {
		fmt.Printf("刷新失败: %v\n", err)
		return
	}
	fmt.Printf("新 Access Token:  %s...（截取前 50 字符）\n", newPair.AccessToken[:50])
	fmt.Printf("新 Refresh Token: %s...（截取前 50 字符）\n", newPair.RefreshToken[:50])

	// --- 4. 验证旧的 Refresh Token 已失效（Rotation 安全策略） ---
	fmt.Println("\n--- 4. 验证旧 Refresh Token 已失效（Rotation 策略） ---")
	_, err = svc.RefreshTokenPair(pair.RefreshToken)
	if err != nil {
		fmt.Printf("✅ 预期行为：旧 Refresh Token 已失效 — %v\n", err)
	}

	// --- 5. 模拟 Token 注销（黑名单） ---
	fmt.Println("\n--- 5. Token 注销（黑名单机制） ---")
	newClaims, _ := svc.ParseAccessToken(newPair.AccessToken)
	svc.RevokeToken(newClaims.ID, newClaims.ExpiresAt.Time)
	fmt.Println("已将当前 Access Token 加入黑名单")

	_, err = svc.ParseAccessToken(newPair.AccessToken)
	if err != nil {
		fmt.Printf("✅ 预期行为：已注销的 Token 无法使用 — %v\n", err)
	}

	// --- 6. 黑名单清理 ---
	fmt.Println("\n--- 6. 黑名单清理（清除已过期条目） ---")
	cleaned := svc.blacklist.Cleanup()
	fmt.Printf("清理了 %d 条已过期的黑名单记录\n", cleaned)

	// --- 7. 演示错误场景 ---
	fmt.Println("\n--- 7. 错误场景演示 ---")

	// 7a. 篡改 Token
	fmt.Println("\n[7a] 篡改 Token:")
	tamperedToken := newPair.AccessToken[:len(newPair.AccessToken)-5] + "XXXXX"
	_, err = svc.ParseAccessToken(tamperedToken)
	if err != nil {
		fmt.Printf("✅ 篡改检测：%v\n", err)
	}

	// 7b. 完全无效的 Token
	fmt.Println("\n[7b] 无效 Token:")
	_, err = svc.ParseAccessToken("invalid.token.string")
	if err != nil {
		fmt.Printf("✅ 无效 Token：%v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("JWT 双令牌机制演示完成")
	fmt.Println(strings.Repeat("=", 60))
}
