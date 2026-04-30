// STS AssumeRole 临时凭证 — 完整模拟示例
// 演示：AssumeRole 流程 / 临时凭证生成 / Token 过期与刷新 / 跨账号访问 / 凭证缓存
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例为纯 Go 实现，无需 Docker 或 AWS 账号
// 通过内存模拟完整的 STS AssumeRole 工作流程
//
// 运行方式：
//   go run ./sts/

package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================
// STS 核心数据结构
// ============================================================

// IAMRole IAM 角色定义
type IAMRole struct {
	RoleARN       string            // 角色 ARN（Amazon Resource Name）
	RoleName      string            // 角色名称
	TrustPolicy   TrustPolicy       // 信任策略（谁可以 AssumeRole）
	Permissions   []string          // 权限列表
	MaxSessionDur time.Duration     // 最大会话时长
	Tags          map[string]string // 角色标签
}

// TrustPolicy 信任策略（定义谁可以扮演此角色）
type TrustPolicy struct {
	AllowedPrincipals []string // 允许的调用者 ARN
	RequireExternalID bool     // 是否要求 ExternalID（防止混淆代理攻击）
	ExternalID        string   // 外部 ID
}

// TemporaryCredentials 临时安全凭证
type TemporaryCredentials struct {
	AccessKeyID     string    // 临时访问密钥 ID（以 ASIA 开头）
	SecretAccessKey string    // 临时访问密钥
	SessionToken    string    // 会话令牌（必须携带）
	Expiration      time.Time // 过期时间
	RoleARN         string    // 扮演的角色
	SessionName     string    // 会话名称
}

// IsExpired 检查凭证是否已过期
func (c *TemporaryCredentials) IsExpired() bool {
	return time.Now().After(c.Expiration)
}

// IsExpiringSoon 检查凭证是否即将过期（5 分钟内）
func (c *TemporaryCredentials) IsExpiringSoon() bool {
	return time.Until(c.Expiration) < 5*time.Minute
}

// CallerIdentity 调用者身份信息
type CallerIdentity struct {
	Account string // AWS 账号 ID
	ARN     string // 调用者 ARN
	UserID  string // 用户 ID
}

// ============================================================
// STS 服务模拟
// ============================================================

// STSService 模拟 AWS STS 服务
type STSService struct {
	mu        sync.RWMutex
	roles     map[string]*IAMRole                // roleARN -> role
	sessions  map[string]*TemporaryCredentials    // accessKeyID -> credentials
	accounts  map[string]*CallerIdentity          // accessKeyID -> identity
	secretKey string                              // 用于生成凭证的密钥
}

// NewSTSService 创建 STS 服务实例
func NewSTSService() *STSService {
	svc := &STSService{
		roles:     make(map[string]*IAMRole),
		sessions:  make(map[string]*TemporaryCredentials),
		accounts:  make(map[string]*CallerIdentity),
		secretKey: "sts-demo-master-secret-2025",
	}
	svc.setupDemoData()
	return svc
}

// setupDemoData 初始化演示数据
func (s *STSService) setupDemoData() {
	// 账号 A 的 IAM 用户
	s.accounts["AKIAIOSFODNN7EXAMPLE"] = &CallerIdentity{
		Account: "111111111111",
		ARN:     "arn:aws:iam::111111111111:user/developer",
		UserID:  "AIDAIOSFODNN7EXAMPLE",
	}

	// 账号 B 的 S3 只读角色
	s.roles["arn:aws:iam::222222222222:role/S3ReadOnlyRole"] = &IAMRole{
		RoleARN:  "arn:aws:iam::222222222222:role/S3ReadOnlyRole",
		RoleName: "S3ReadOnlyRole",
		TrustPolicy: TrustPolicy{
			AllowedPrincipals: []string{"arn:aws:iam::111111111111:root"},
			RequireExternalID: false,
		},
		Permissions:   []string{"s3:GetObject", "s3:ListBucket"},
		MaxSessionDur: 1 * time.Hour,
	}

	// 账号 B 的管理员角色（需要 ExternalID）
	s.roles["arn:aws:iam::222222222222:role/AdminRole"] = &IAMRole{
		RoleARN:  "arn:aws:iam::222222222222:role/AdminRole",
		RoleName: "AdminRole",
		TrustPolicy: TrustPolicy{
			AllowedPrincipals: []string{"arn:aws:iam::111111111111:root"},
			RequireExternalID: true,
			ExternalID:        "partner-external-id-12345",
		},
		Permissions:   []string{"s3:*", "sqs:*", "sts:*"},
		MaxSessionDur: 12 * time.Hour,
	}

	// 账号 C 的 IoT 角色
	s.roles["arn:aws:iam::333333333333:role/IoTDeviceRole"] = &IAMRole{
		RoleARN:  "arn:aws:iam::333333333333:role/IoTDeviceRole",
		RoleName: "IoTDeviceRole",
		TrustPolicy: TrustPolicy{
			AllowedPrincipals: []string{"arn:aws:iam::111111111111:root"},
		},
		Permissions:   []string{"iot:Connect", "iot:Publish", "iot:Subscribe"},
		MaxSessionDur: 1 * time.Hour,
	}
}

// generateAccessKeyID 生成临时 Access Key ID（以 ASIA 开头表示临时凭证）
func generateAccessKeyID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "ASIA" + strings.ToUpper(hex.EncodeToString(b))[:16]
}

// generateSecretKey 生成临时 Secret Access Key
func (s *STSService) generateSecretKey(roleARN, sessionName string) string {
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(fmt.Sprintf("%s:%s:%d", roleARN, sessionName, time.Now().UnixNano())))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))[:40]
}

// generateSessionToken 生成 Session Token
func (s *STSService) generateSessionToken(accessKeyID, roleARN string) string {
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write([]byte(fmt.Sprintf("session:%s:%s:%d", accessKeyID, roleARN, time.Now().UnixNano())))
	encoded := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return "FwoGZXIvYXdzE" + encoded
}

// AssumeRole 扮演角色，获取临时凭证
func (s *STSService) AssumeRole(callerKeyID, roleARN, sessionName string, duration time.Duration, externalID string) (*TemporaryCredentials, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 验证调用者身份
	caller, exists := s.accounts[callerKeyID]
	if !exists {
		return nil, fmt.Errorf("invalid caller credentials: %s", callerKeyID)
	}

	// 2. 查找目标角色
	role, exists := s.roles[roleARN]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", roleARN)
	}

	// 3. 检查信任策略
	trusted := false
	callerAccount := "arn:aws:iam::" + caller.Account + ":root"
	for _, principal := range role.TrustPolicy.AllowedPrincipals {
		if principal == callerAccount || principal == caller.ARN {
			trusted = true
			break
		}
	}
	if !trusted {
		return nil, fmt.Errorf("caller %s is not trusted by role %s", caller.ARN, roleARN)
	}

	// 4. 检查 ExternalID（防止混淆代理攻击）
	if role.TrustPolicy.RequireExternalID {
		if externalID == "" {
			return nil, fmt.Errorf("ExternalID is required for role %s", roleARN)
		}
		if externalID != role.TrustPolicy.ExternalID {
			return nil, fmt.Errorf("invalid ExternalID for role %s", roleARN)
		}
	}

	// 5. 验证会话时长
	if duration > role.MaxSessionDur {
		duration = role.MaxSessionDur
	}
	if duration == 0 {
		duration = 1 * time.Hour
	}

	// 6. 生成临时凭证
	accessKeyID := generateAccessKeyID()
	creds := &TemporaryCredentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: s.generateSecretKey(roleARN, sessionName),
		SessionToken:    s.generateSessionToken(accessKeyID, roleARN),
		Expiration:      time.Now().Add(duration),
		RoleARN:         roleARN,
		SessionName:     sessionName,
	}

	// 7. 注册临时会话
	s.sessions[accessKeyID] = creds
	s.accounts[accessKeyID] = &CallerIdentity{
		Account: strings.Split(roleARN, ":")[4],
		ARN:     fmt.Sprintf("%s/%s", roleARN, sessionName),
		UserID:  fmt.Sprintf("%s:%s", accessKeyID, sessionName),
	}

	return creds, nil
}

// GetCallerIdentity 获取调用者身份
func (s *STSService) GetCallerIdentity(accessKeyID string) (*CallerIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identity, exists := s.accounts[accessKeyID]
	if !exists {
		return nil, fmt.Errorf("unknown caller: %s", accessKeyID)
	}

	// 检查临时凭证是否过期
	if creds, ok := s.sessions[accessKeyID]; ok {
		if creds.IsExpired() {
			return nil, fmt.Errorf("temporary credentials expired at %s", creds.Expiration.Format(time.RFC3339))
		}
	}

	return identity, nil
}

// ============================================================
// 凭证缓存与自动刷新
// ============================================================

// CredentialsCache 凭证缓存（模拟 SDK 的 aws.CredentialsCache）
type CredentialsCache struct {
	mu          sync.Mutex
	stsService  *STSService
	callerKeyID string
	roleARN     string
	sessionName string
	externalID  string
	duration    time.Duration
	current     *TemporaryCredentials
	refreshCount int
}

// NewCredentialsCache 创建凭证缓存
func NewCredentialsCache(stsService *STSService, callerKeyID, roleARN, sessionName, externalID string, duration time.Duration) *CredentialsCache {
	return &CredentialsCache{
		stsService:  stsService,
		callerKeyID: callerKeyID,
		roleARN:     roleARN,
		sessionName: sessionName,
		externalID:  externalID,
		duration:    duration,
	}
}

// Retrieve 获取凭证（自动刷新）
func (c *CredentialsCache) Retrieve() (*TemporaryCredentials, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果凭证有效且未即将过期，直接返回
	if c.current != nil && !c.current.IsExpired() && !c.current.IsExpiringSoon() {
		return c.current, nil
	}

	// 刷新凭证
	creds, err := c.stsService.AssumeRole(c.callerKeyID, c.roleARN, c.sessionName, c.duration, c.externalID)
	if err != nil {
		return nil, fmt.Errorf("refresh credentials failed: %w", err)
	}

	c.current = creds
	c.refreshCount++
	return creds, nil
}

// ============================================================
// 演示函数
// ============================================================

func main() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("STS AssumeRole 临时凭证（纯 Go 模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	stsService := NewSTSService()

	// --- 1. 基本 AssumeRole ---
	demoBasicAssumeRole(stsService)

	// --- 2. GetCallerIdentity ---
	demoGetCallerIdentity(stsService)

	// --- 3. ExternalID 防护 ---
	demoExternalID(stsService)

	// --- 4. 跨账号访问 ---
	demoCrossAccountAccess(stsService)

	// --- 5. Token 过期与刷新 ---
	demoTokenRefresh(stsService)

	// --- 6. 凭证缓存 ---
	demoCredentialsCache(stsService)
}

// demoBasicAssumeRole 基本 AssumeRole 流程
func demoBasicAssumeRole(stsService *STSService) {
	fmt.Println("\n--- 1. 基本 AssumeRole ---")

	callerKeyID := "AKIAIOSFODNN7EXAMPLE"
	roleARN := "arn:aws:iam::222222222222:role/S3ReadOnlyRole"

	fmt.Printf("  调用者: %s\n", callerKeyID)
	fmt.Printf("  目标角色: %s\n", roleARN)

	creds, err := stsService.AssumeRole(callerKeyID, roleARN, "my-session", 1*time.Hour, "")
	if err != nil {
		fmt.Printf("  ❌ AssumeRole 失败: %v\n", err)
		return
	}

	fmt.Println("\n  临时凭证:")
	fmt.Printf("    AccessKeyId:     %s\n", creds.AccessKeyID)
	fmt.Printf("    SecretAccessKey: %s...（已截断）\n", creds.SecretAccessKey[:20])
	fmt.Printf("    SessionToken:    %s...（已截断）\n", creds.SessionToken[:30])
	fmt.Printf("    Expiration:      %s\n", creds.Expiration.Format(time.RFC3339))
	fmt.Printf("    过期: %v | 即将过期: %v\n", creds.IsExpired(), creds.IsExpiringSoon())

	fmt.Println("\n  ⚠️  使用临时凭证时，必须同时携带 AccessKeyId + SecretAccessKey + SessionToken")
}

// demoGetCallerIdentity GetCallerIdentity 调试
func demoGetCallerIdentity(stsService *STSService) {
	fmt.Println("\n--- 2. GetCallerIdentity ---")

	// 原始身份
	identity, _ := stsService.GetCallerIdentity("AKIAIOSFODNN7EXAMPLE")
	fmt.Println("  原始身份:")
	fmt.Printf("    Account: %s\n", identity.Account)
	fmt.Printf("    ARN:     %s\n", identity.ARN)
	fmt.Printf("    UserID:  %s\n", identity.UserID)

	// AssumeRole 后的身份
	creds, _ := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE",
		"arn:aws:iam::222222222222:role/S3ReadOnlyRole", "debug-session", 1*time.Hour, "")

	identity2, _ := stsService.GetCallerIdentity(creds.AccessKeyID)
	fmt.Println("\n  AssumeRole 后身份:")
	fmt.Printf("    Account: %s（目标账号）\n", identity2.Account)
	fmt.Printf("    ARN:     %s\n", identity2.ARN)
	fmt.Printf("    UserID:  %s\n", identity2.UserID)
	fmt.Println("  ✅ GetCallerIdentity 常用于调试，确认当前使用的身份")
}

// demoExternalID ExternalID 防止混淆代理攻击
func demoExternalID(stsService *STSService) {
	fmt.Println("\n--- 3. ExternalID 防护（防止混淆代理攻击） ---")

	roleARN := "arn:aws:iam::222222222222:role/AdminRole"

	// 不提供 ExternalID → 失败
	_, err := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE", roleARN, "test", 1*time.Hour, "")
	fmt.Printf("  无 ExternalID: %v\n", err)

	// 错误的 ExternalID → 失败
	_, err = stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE", roleARN, "test", 1*time.Hour, "wrong-id")
	fmt.Printf("  错误 ExternalID: %v\n", err)

	// 正确的 ExternalID → 成功
	creds, err := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE", roleARN, "admin-session", 1*time.Hour, "partner-external-id-12345")
	if err != nil {
		fmt.Printf("  正确 ExternalID: 失败 %v\n", err)
	} else {
		fmt.Printf("  正确 ExternalID: 成功 ✅ (AccessKeyId=%s)\n", creds.AccessKeyID)
	}

	fmt.Println("\n  混淆代理攻击场景:")
	fmt.Println("    攻击者 A 知道你的服务会 AssumeRole 到角色 X")
	fmt.Println("    A 伪造请求让你的服务代替 A 去 AssumeRole")
	fmt.Println("    ExternalID 是只有合法调用者知道的秘密值，防止此类攻击")
}

// demoCrossAccountAccess 跨账号访问
func demoCrossAccountAccess(stsService *STSService) {
	fmt.Println("\n--- 4. 跨账号访问 ---")

	fmt.Println("  场景: 账号 A (111111111111) 的开发者访问账号 C (333333333333) 的 IoT 资源")

	// 步骤 1: 使用账号 A 的凭证
	fmt.Println("\n  步骤 1: 使用账号 A 的长期凭证")
	identity, _ := stsService.GetCallerIdentity("AKIAIOSFODNN7EXAMPLE")
	fmt.Printf("    当前身份: %s (账号 %s)\n", identity.ARN, identity.Account)

	// 步骤 2: AssumeRole 到账号 C 的 IoT 角色
	fmt.Println("\n  步骤 2: AssumeRole 到账号 C 的 IoT 角色")
	creds, err := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE",
		"arn:aws:iam::333333333333:role/IoTDeviceRole", "iot-session", 1*time.Hour, "")
	if err != nil {
		fmt.Printf("    ❌ 失败: %v\n", err)
		return
	}
	fmt.Printf("    获取临时凭证: %s (过期: %s)\n", creds.AccessKeyID, creds.Expiration.Format("15:04:05"))

	// 步骤 3: 使用临时凭证访问账号 C 的资源
	fmt.Println("\n  步骤 3: 使用临时凭证访问账号 C 的 IoT 资源")
	identity2, _ := stsService.GetCallerIdentity(creds.AccessKeyID)
	fmt.Printf("    当前身份: %s (账号 %s)\n", identity2.ARN, identity2.Account)
	fmt.Println("    ✅ 成功以账号 C 的 IoT 角色身份访问资源")
}

// demoTokenRefresh Token 过期与刷新
func demoTokenRefresh(stsService *STSService) {
	fmt.Println("\n--- 5. Token 过期与刷新 ---")

	// 创建一个短有效期的凭证（演示用）
	creds, _ := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE",
		"arn:aws:iam::222222222222:role/S3ReadOnlyRole", "short-session", 10*time.Second, "")

	fmt.Printf("  创建短期凭证: 有效期 10s\n")
	fmt.Printf("    过期时间: %s\n", creds.Expiration.Format("15:04:05"))
	fmt.Printf("    已过期: %v\n", creds.IsExpired())

	// 模拟使用凭证
	fmt.Println("\n  模拟使用凭证:")
	for i := 1; i <= 3; i++ {
		identity, err := stsService.GetCallerIdentity(creds.AccessKeyID)
		if err != nil {
			fmt.Printf("    [%ds] ❌ 凭证已过期: %v\n", i*5, err)
			fmt.Println("    → 需要重新 AssumeRole 获取新凭证")

			// 刷新凭证
			newCreds, _ := stsService.AssumeRole("AKIAIOSFODNN7EXAMPLE",
				"arn:aws:iam::222222222222:role/S3ReadOnlyRole", "refreshed-session", 1*time.Hour, "")
			fmt.Printf("    → 刷新成功: 新 AccessKeyId=%s\n", newCreds.AccessKeyID)
			break
		}
		fmt.Printf("    [%ds] ✅ 凭证有效，身份: %s\n", i*5, identity.ARN)
		time.Sleep(5 * time.Second)
	}
}

// demoCredentialsCache 凭证缓存（自动刷新）
func demoCredentialsCache(stsService *STSService) {
	fmt.Println("\n--- 6. 凭证缓存（自动刷新） ---")

	cache := NewCredentialsCache(stsService,
		"AKIAIOSFODNN7EXAMPLE",
		"arn:aws:iam::222222222222:role/S3ReadOnlyRole",
		"cached-session",
		"",
		1*time.Hour,
	)

	fmt.Println("  模拟多次获取凭证（缓存自动管理）:")
	for i := 1; i <= 3; i++ {
		creds, err := cache.Retrieve()
		if err != nil {
			fmt.Printf("    [%d] ❌ 获取失败: %v\n", i, err)
			continue
		}
		fmt.Printf("    [%d] AccessKeyId=%s (刷新次数: %d)\n", i, creds.AccessKeyID, cache.refreshCount)
	}

	fmt.Println("\n  CredentialsCache 工作原理:")
	fmt.Println("    1. 首次调用 → AssumeRole 获取凭证并缓存")
	fmt.Println("    2. 后续调用 → 返回缓存的凭证（避免重复请求 STS）")
	fmt.Println("    3. 凭证即将过期（< 5min）→ 自动刷新")
	fmt.Println("    4. AWS SDK 的 aws.NewCredentialsCache() 实现了相同逻辑")
}
