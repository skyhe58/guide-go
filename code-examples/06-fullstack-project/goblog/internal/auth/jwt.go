// Package auth 提供 GoBlog 的认证鉴权功能
// 包含 JWT 双令牌签发与验证、密码加密、RBAC 角色权限管理
package auth

import (
	"errors"
	"time"

	"guide-go/goblog/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType 令牌类型
type TokenType string

const (
	// AccessToken 访问令牌，有效期短（15 分钟）
	AccessToken TokenType = "access"
	// RefreshToken 刷新令牌，有效期长（7 天）
	RefreshToken TokenType = "refresh"
)

// Claims 自定义 JWT Claims，包含用户信息
type Claims struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenPair 双令牌对，包含 Access Token 和 Refresh Token
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access Token 过期时间（秒）
}

// JWTService JWT 服务，负责令牌的签发、验证和刷新
type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewJWTService 创建 JWT 服务实例
func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		issuer:          cfg.Issuer,
	}
}

// GenerateTokenPair 签发双令牌（Access Token + Refresh Token）
func (s *JWTService) GenerateTokenPair(userID uint, username, role string) (*TokenPair, error) {
	// 签发 Access Token
	accessToken, err := s.generateToken(userID, username, role, AccessToken, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	// 签发 Refresh Token
	refreshToken, err := s.generateToken(userID, username, role, RefreshToken, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
	}, nil
}

// ParseToken 解析并验证 JWT Token，返回 Claims
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名算法")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Token")
	}

	return claims, nil
}

// RefreshAccessToken 使用 Refresh Token 刷新 Access Token
func (s *JWTService) RefreshAccessToken(refreshTokenStr string) (*TokenPair, error) {
	// 解析 Refresh Token
	claims, err := s.ParseToken(refreshTokenStr)
	if err != nil {
		return nil, err
	}

	// 验证是否为 Refresh Token 类型
	if claims.Type != RefreshToken {
		return nil, errors.New("不是有效的 Refresh Token")
	}

	// 签发新的令牌对
	return s.GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
}

// GetAccessTokenTTL 获取 Access Token 有效期
func (s *JWTService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

// generateToken 内部方法：签发指定类型的 JWT Token
func (s *JWTService) generateToken(userID uint, username, role string, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			ID:        uuid.New().String(), // JTI 用于 Token 黑名单
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}
