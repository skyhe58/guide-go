---
title: "STS 临时安全凭证"
module: "aws"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - STS
  - AssumeRole
  - 临时凭证
  - 跨账号访问
  - Token 刷新
codeExample: "03-microservice/aws/sts/"
relatedEntries:
  - "/3-microservice/3.4-aws/02-auth"
  - "/3-microservice/3.4-aws/01-sdk-basics"
prerequisites:
  - "/3-microservice/3.4-aws/02-auth"
estimatedTime: "45min"
---

# STS 临时安全凭证

## 概念说明

AWS STS（Security Token Service）是 AWS 的临时安全凭证服务，用于生成有时间限制的临时访问凭证。STS 是 AWS 安全最佳实践的核心组件——生产环境应尽量使用临时凭证而非长期 Access Key。

STS 的核心操作：
- **AssumeRole**：扮演 IAM 角色，获取该角色的临时凭证（最常用）
- **AssumeRoleWithWebIdentity**：使用 OIDC Token 扮演角色（K8s IRSA）
- **GetSessionToken**：为 IAM 用户生成临时凭证（MFA 场景）
- **GetCallerIdentity**：查询当前调用者身份（调试用）

## 核心原理

### AssumeRole 工作流程

```mermaid
sequenceDiagram
    participant APP as Go 应用<br/>（账号 A）
    participant STS as AWS STS
    participant S3 as AWS S3<br/>（账号 B）
    
    Note over APP: 使用账号 A 的凭证
    APP->>STS: 1. AssumeRole(RoleARN=账号B的角色)
    Note over STS: 验证：<br/>1. 调用者身份<br/>2. 角色信任策略<br/>3. 权限策略
    STS-->>APP: 2. 临时凭证<br/>AccessKeyId<br/>SecretAccessKey<br/>SessionToken<br/>Expiration
    
    Note over APP: 使用临时凭证
    APP->>S3: 3. 访问账号 B 的 S3
    S3-->>APP: 4. 返回数据
    
    Note over APP: 凭证即将过期
    APP->>STS: 5. 再次 AssumeRole（刷新）
    STS-->>APP: 6. 新的临时凭证
```

### 临时凭证组成

| 字段 | 说明 | 示例 |
|------|------|------|
| `AccessKeyId` | 临时访问密钥 ID | `ASIAIOSFODNN7EXAMPLE` |
| `SecretAccessKey` | 临时访问密钥 | `wJalrXUtnFEMI/...` |
| `SessionToken` | 会话令牌（必须携带） | `FwoGZXIvYXdzEB...` |
| `Expiration` | 过期时间 | `2025-01-01T12:00:00Z` |

### 信任策略（Trust Policy）

角色的信任策略定义了谁可以 AssumeRole：

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::111111111111:root"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "sts:ExternalId": "my-external-id"
      }
    }
  }]
}
```

### Token 刷新策略

```mermaid
graph TD
    A[获取临时凭证] --> B{凭证是否过期?}
    B -->|未过期| C[使用当前凭证]
    B -->|即将过期<br/>剩余 < 5min| D[异步刷新凭证]
    D --> E[AssumeRole 获取新凭证]
    E --> F[更新凭证缓存]
    F --> C
    B -->|已过期| G[同步刷新凭证]
    G --> E
```

## 标准库方案

Go 标准库不提供 STS 客户端。STS 操作通过 AWS SDK v2 的 `service/sts` 包实现。

## 第三方库方案

### AssumeRole 基本使用

```go
import (
    "github.com/aws/aws-sdk-go-v2/service/sts"
    "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

// 方式一：直接调用 STS API
stsClient := sts.NewFromConfig(cfg)
output, _ := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
    RoleArn:         aws.String("arn:aws:iam::123456789012:role/MyRole"),
    RoleSessionName: aws.String("my-session"),
    DurationSeconds: aws.Int32(3600), // 1 小时
})
// output.Credentials 包含临时凭证

// 方式二：使用 stscreds 自动刷新（推荐）
creds := stscreds.NewAssumeRoleProvider(stsClient, "arn:aws:iam::123456789012:role/MyRole")
cfg.Credentials = aws.NewCredentialsCache(creds)
// SDK 会自动在凭证过期前刷新
```

### GetCallerIdentity（调试）

```go
output, _ := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
fmt.Printf("Account: %s\n", *output.Account)
fmt.Printf("ARN: %s\n", *output.Arn)
fmt.Printf("UserId: %s\n", *output.UserId)
```

## 代码示例

> 💻 完整可运行代码：[code-examples/03-microservice/aws/sts/](https://github.com/your-repo/code-examples/03-microservice/aws/sts/)
> 🏷️ Demo 模式：纯 Go（模拟 STS AssumeRole 流程和 Token 刷新）

## 常见面试题

### Q1: STS AssumeRole 的使用场景有哪些？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 跨账号访问
2. 最小权限原则
3. 临时提权
4. 第三方访问

**标准答案**：

AssumeRole 的核心场景：跨账号访问——账号 A 的应用需要访问账号 B 的资源，通过 AssumeRole 获取账号 B 的临时凭证。最小权限——应用平时使用低权限凭证，需要执行特权操作时临时 AssumeRole 获取高权限。第三方访问——允许外部合作伙伴通过 AssumeRole 访问特定资源，通过 ExternalId 防止混淆代理攻击。K8s Pod 身份——EKS 中通过 IRSA（IAM Roles for Service Accounts）让 Pod 使用 AssumeRoleWithWebIdentity 获取凭证。

**深入追问**：

- 什么是混淆代理问题（Confused Deputy）？
- AssumeRole 的最大会话时长是多少？

## 常见陷阱

1. **忘记携带 SessionToken**：使用临时凭证时，必须同时设置 AccessKeyId、SecretAccessKey 和 SessionToken
2. **会话时长设置不当**：默认 1 小时，最长 12 小时（取决于角色配置），设置过短会频繁刷新
3. **信任策略配置错误**：AssumeRole 失败通常是角色的信任策略未允许调用者
4. **未处理凭证刷新**：临时凭证会过期，应使用 SDK 的 `CredentialsCache` 自动刷新

## 参考资料

- [AWS STS 官方文档](https://docs.aws.amazon.com/STS/latest/APIReference/)
- [IAM 角色最佳实践](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [跨账号访问指南](https://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_cross-account-with-roles.html)
