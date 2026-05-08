---
title: "AWS 认证方式"
module: "aws"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - IAM
  - Access Key
  - STS
  - AssumeRole
  - Web Identity
  - 认证
codeExample: "03-microservice/aws/sts/"
relatedEntries:
  - "/3-microservice/3.4-aws/01-sdk-basics"
  - "/3-microservice/3.4-aws/07-sts"
prerequisites:
  - "/1-go-core/1.1-go-basics/"
estimatedTime: "45min"
---

# AWS 认证方式

## 概念说明

AWS 的认证体系基于 IAM（Identity and Access Management），是 AWS 安全模型的核心。所有对 AWS 服务的 API 调用都需要经过认证和授权。Go 应用通过 AWS SDK 调用 AWS 服务时，必须提供有效的凭证。

AWS 提供多种认证方式，适用于不同的部署场景：本地开发使用 Access Key，EC2/ECS 部署使用 IAM 角色，跨账号访问使用 STS AssumeRole，Kubernetes Pod 使用 Web Identity。

## 核心原理

### IAM 权限模型

```mermaid
graph TB
    subgraph "IAM 实体"
        USER[IAM User<br/>长期凭证]
        ROLE[IAM Role<br/>临时凭证]
        GROUP[IAM Group<br/>用户分组]
    end
    
    subgraph "IAM 策略"
        POLICY[IAM Policy<br/>JSON 权限文档]
        INLINE[内联策略]
        MANAGED[托管策略]
    end
    
    subgraph "AWS 资源"
        S3[S3 Bucket]
        SQS[SQS Queue]
        EC2[EC2 Instance]
    end
    
    USER --> GROUP
    GROUP --> MANAGED
    USER --> INLINE
    ROLE --> MANAGED
    POLICY --> S3 & SQS & EC2
```

### 认证方式对比

| 认证方式 | 凭证类型 | 有效期 | 适用场景 | 安全级别 |
|---------|---------|--------|---------|---------|
| Access Key | 长期密钥 | 永久（需手动轮换） | 本地开发、CI/CD | ⚠️ 中 |
| IAM 角色 | 临时凭证 | 自动轮换 | EC2/ECS/Lambda | ✅ 高 |
| STS AssumeRole | 临时凭证 | 15min-12h | 跨账号访问 | ✅ 高 |
| Web Identity | OIDC Token | 可配置 | K8s Pod（IRSA） | ✅ 高 |
| SSO | 临时凭证 | 可配置 | 企业 SSO 集成 | ✅ 高 |

### STS AssumeRole 流程

```mermaid
sequenceDiagram
    participant APP as Go 应用
    participant STS as AWS STS
    participant S3 as AWS S3
    
    APP->>STS: 1. AssumeRole(RoleARN, SessionName)
    Note over STS: 验证调用者身份<br/>检查信任策略
    STS-->>APP: 2. 临时凭证（AccessKey + SecretKey + SessionToken）
    Note over APP: 凭证有效期: 15min-12h
    APP->>S3: 3. 使用临时凭证调用 S3 API
    S3-->>APP: 4. 返回结果
    
    Note over APP: 凭证过期前自动刷新
    APP->>STS: 5. 再次 AssumeRole（刷新凭证）
    STS-->>APP: 6. 新的临时凭证
```

## 标准库方案

Go 标准库不提供 AWS 认证功能。认证逻辑通过 AWS SDK v2 的 `credentials` 包实现。

## 第三方库方案

### 静态凭证（本地开发）

```go
import "github.com/aws/aws-sdk-go-v2/credentials"

cfg, _ := config.LoadDefaultConfig(context.TODO(),
    config.WithCredentialsProvider(
        credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
    ),
)
```

### 环境变量

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=us-east-1
```

### STS AssumeRole

```go
import (
    "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

stsClient := sts.NewFromConfig(cfg)
creds := stscreds.NewAssumeRoleProvider(stsClient, "arn:aws:iam::123456789012:role/MyRole")
cfg.Credentials = aws.NewCredentialsCache(creds)
```

## 代码示例

> 💻 完整可运行代码：[code-examples/03-microservice/aws/sts/](https://github.com/skyhe58/guide-go/tree/main/code-examples/03-microservice/aws/sts/)
> 🏷️ Demo 模式：纯 Go（模拟 STS AssumeRole 流程）

## 常见面试题

### Q1: AWS IAM 的权限模型是怎样的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. IAM 实体（User/Role/Group）
2. IAM 策略（Policy）结构
3. 权限评估逻辑（显式拒绝 > 显式允许 > 默认拒绝）

**标准答案**：

IAM 权限模型基于策略（Policy）文档，策略是 JSON 格式，包含 Effect（Allow/Deny）、Action（API 操作）、Resource（资源 ARN）、Condition（条件）四个核心元素。策略可以附加到 User、Group 或 Role 上。权限评估遵循"显式拒绝优先"原则：如果任何策略显式拒绝，则拒绝；否则如果有显式允许，则允许；否则默认拒绝。生产环境推荐使用 IAM Role 而非 User，遵循最小权限原则。

**深入追问**：

- 什么是 IAM 策略的 Condition 条件？
- 如何实现跨账号访问？

### Q2: 为什么生产环境推荐使用 IAM 角色而非 Access Key？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 安全性对比
2. 凭证管理成本
3. 最小权限原则

**标准答案**：

IAM 角色使用临时凭证，自动轮换，无需在代码或配置中硬编码密钥，降低泄露风险。Access Key 是长期凭证，一旦泄露需要手动轮换，且可能已被滥用。IAM 角色可以精确控制权限范围，支持条件策略（如限制 IP、时间窗口）。在 EC2/ECS/Lambda 等 AWS 计算服务上，IAM 角色通过实例元数据服务（IMDS）自动注入凭证，应用无需任何凭证配置。

**深入追问**：

- EC2 实例如何获取 IAM 角色的临时凭证？
- STS AssumeRole 的信任策略如何配置？

## 常见陷阱

1. **Access Key 硬编码在代码中**：绝对不要将 Access Key 写在代码里，应使用环境变量或 IAM 角色
2. **权限过大**：不要使用 `AdministratorAccess` 策略，应遵循最小权限原则
3. **忘记设置 SessionToken**：使用 STS 临时凭证时，必须同时设置 AccessKeyId、SecretAccessKey 和 SessionToken
4. **未处理凭证过期**：临时凭证有有效期，SDK 会自动刷新，但自行管理时需要处理过期逻辑

## 参考资料

- [AWS IAM 官方文档](https://docs.aws.amazon.com/IAM/latest/UserGuide/)
- [AWS STS 官方文档](https://docs.aws.amazon.com/STS/latest/APIReference/)
- [AWS SDK for Go v2 Credentials](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/credentials/)
