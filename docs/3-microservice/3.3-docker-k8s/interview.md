---
title: "Docker 与 K8s 面试指南"
module: "docker-k8s"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - 面试
  - Docker
  - Kubernetes
  - 容器化
  - 云原生
relatedEntries:
  - "/3-microservice/3.3-docker-k8s/01-docker-basics"
  - "/3-microservice/3.3-docker-k8s/05-k8s-architecture"
  - "/3-microservice/3.3-docker-k8s/07-k8s-go-deploy"
prerequisites:
  - "/3-microservice/3.3-docker-k8s/"
estimatedTime: "60min"
---

# Docker 与 K8s 面试指南

## 面试知识图谱

```mermaid
graph TB
    subgraph "Docker 面试重点"
        D1[容器 vs 虚拟机]
        D2[镜像分层原理]
        D3[多阶段构建]
        D4[网络模式]
        D5[数据持久化]
        D1 --> D2
        D2 --> D3
        D4 --> D5
    end
    
    subgraph "K8s 面试重点"
        K1[架构与组件]
        K2[Pod 生命周期]
        K3[Deployment 滚动更新]
        K4[Service 类型]
        K5[健康检查]
        K6[优雅停机]
        K7[HPA 扩缩容]
        K1 --> K2
        K2 --> K3
        K3 --> K4
        K5 --> K6
        K6 --> K7
    end
    
    subgraph "Go + K8s 面试重点"
        G1[Go Dockerfile 优化]
        G2[signal.NotifyContext]
        G3[/healthz /readyz]
        G4[Prometheus 指标]
        G1 --> G2
        G2 --> G3
        G3 --> G4
    end
    
    D3 --> G1
    K5 --> G3
    K6 --> G2
```

## Docker 面试题

### Q1: Docker 容器和虚拟机有什么区别？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 虚拟化层级不同
2. 性能和资源对比
3. 隔离机制不同

**标准答案**：

Docker 容器是操作系统级虚拟化，共享宿主机内核，通过 Namespace 隔离进程、网络、文件系统，通过 Cgroup 限制资源。虚拟机是硬件级虚拟化，通过 Hypervisor 运行完整的 Guest OS。容器启动秒级、MB 级资源占用、接近原生性能；虚拟机启动分钟级、GB 级资源占用、有明显性能损耗。Go 应用容器化优势尤其明显——编译为静态二进制文件，基于 scratch 空镜像，最终镜像可以小到 10MB 以内。

**深入追问**：

- Namespace 有哪些类型？
- Cgroup v1 和 v2 的区别？
- 容器逃逸如何防范？

### Q2: 如何将 Go 应用的 Docker 镜像优化到 10MB 以内？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

核心是多阶段构建：第一阶段用 `golang:1.22-alpine` 编译，设置 `CGO_ENABLED=0` 确保纯静态链接，`-ldflags="-s -w"` 去掉调试信息；第二阶段用 `scratch` 空镜像，只复制二进制文件和必要的 CA 证书。最终镜像通常 6-10MB。可选 UPX 压缩进一步减小，但会增加启动时间。

**深入追问**：

- CGO_ENABLED=0 不设置会怎样？
- scratch 和 distroless 的区别？
- 如何处理需要 CGO 的场景？

### Q3: Docker 镜像的分层存储原理？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

Docker 镜像基于 Union FS（如 overlay2）实现分层存储。每条 Dockerfile 指令创建一个只读层，运行容器时在顶部添加可写层。层是只读的，可被多个镜像共享，节省存储空间。这也是为什么 Dockerfile 中要将不常变化的指令放在前面——利用构建缓存加速构建。Go 项目的最佳实践是先 COPY go.mod/go.sum 并 go mod download，再 COPY 源码编译。

### Q4: Docker 的网络模式有哪些？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

四种主要网络模式：bridge（默认，虚拟网桥，适合单机多容器）、host（共享宿主机网络栈，最高性能）、none（无网络，完全隔离）、overlay（跨主机通信，用于 Swarm/多主机）。日常开发最常用自定义 bridge 网络，它支持 DNS 解析（可通过容器名访问），Docker Compose 默认创建自定义 bridge。

## Kubernetes 面试题

### Q5: 简述 K8s 的架构和核心组件

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

K8s 采用 Master-Node 架构。控制平面：API Server（集群入口，所有操作的网关）、etcd（分布式 KV 存储，保存集群状态）、Scheduler（Pod 调度）、Controller Manager（控制循环，确保实际状态=期望状态）。工作节点：kubelet（管理 Pod 生命周期）、kube-proxy（网络代理，Service 负载均衡）、Container Runtime（containerd，运行容器）。K8s 采用声明式管理，用户定义期望状态，控制器持续调谐。

**深入追问**：

- etcd 存储了哪些数据？
- API Server 的认证授权流程？
- Scheduler 的调度策略？

### Q6: Pod 的创建流程是怎样的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

kubectl 提交 YAML → API Server 验证并存储到 etcd → Controller Manager 创建 ReplicaSet → ReplicaSet Controller 创建 Pod（Pending）→ Scheduler Watch 到未调度 Pod，选择 Node 并绑定 → kubelet Watch 到分配的 Pod，调用 Container Runtime 拉取镜像、创建容器 → kubelet 上报 Pod 状态为 Running。

### Q7: Deployment、ReplicaSet、Pod 三者的关系？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

三层层级关系：Deployment → ReplicaSet → Pod。Deployment 管理更新策略（滚动更新、回滚），ReplicaSet 维持副本数，Pod 运行容器。滚动更新时，Deployment 创建新 ReplicaSet，逐步增加新副本、减少旧副本，实现零停机更新。

### Q8: K8s Service 的类型及适用场景？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

四种类型：ClusterIP（默认，集群内部访问）、NodePort（每个 Node 开放端口，开发测试用）、LoadBalancer（云厂商负载均衡器，生产外部访问）、ExternalName（DNS CNAME，访问外部服务）。生产环境推荐 ClusterIP + Ingress 组合。

### Q9: livenessProbe 和 readinessProbe 的区别？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

livenessProbe 检测应用是否存活，失败后重启 Pod；readinessProbe 检测应用是否就绪，失败后从 Service Endpoints 移除（不接收流量），但不重启。Go 应用中 `/healthz` 返回 200 表示存活，`/readyz` 检查依赖（DB、Redis）后返回 200 表示就绪。注意：livenessProbe 不应检查外部依赖，否则依赖故障会导致所有 Pod 被重启。

## Go + K8s 面试题

### Q10: Go 服务如何实现优雅停机？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

使用 Go 1.16+ 的 `signal.NotifyContext` 捕获 SIGTERM 信号，收到信号后调用 `http.Server.Shutdown()` 优雅关闭 HTTP 服务器（停止接受新连接，等待进行中请求完成），然后依次关闭数据库、Redis 等连接。K8s 中需注意 `terminationGracePeriodSeconds`（默认 30 秒）要大于应用关闭超时时间。

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
defer stop()
// ... 启动服务器
<-ctx.Done()
server.Shutdown(context.WithTimeout(context.Background(), 15*time.Second))
```

### Q11: Go 应用在 K8s 中如何实现健康检查？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

实现两个 HTTP 端点：`/healthz`（存活检查，返回 200 即可）和 `/readyz`（就绪检查，需验证数据库、Redis 等依赖可用）。在 Deployment YAML 中配置 livenessProbe 和 readinessProbe 指向这两个端点。启动阶段可配合 startupProbe 避免慢启动被误杀。

### Q12: 如何设计一个 Go 微服务的 K8s 部署方案？

**难度**：⭐⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

完整方案包含：
1. **Dockerfile**：多阶段构建，scratch 基础镜像，≤ 10MB
2. **Deployment**：多副本、滚动更新、资源限制、健康检查
3. **Service**：ClusterIP 内部通信
4. **Ingress**：域名路由、TLS 终止
5. **ConfigMap/Secret**：配置和敏感数据管理
6. **HPA**：基于 CPU/自定义指标自动扩缩容
7. **应用层**：signal.NotifyContext 优雅停机、/healthz /readyz 健康检查、Prometheus 指标暴露
8. **Helm**：模板化管理，支持多环境部署

## 面试题难度分布

| 难度 | 题目 | 适用岗位 |
|------|------|---------|
| ⭐⭐ | Q1 容器 vs 虚拟机、Q4 网络模式、Q11 健康检查 | 初级/中级 |
| ⭐⭐⭐ | Q2 镜像优化、Q3 分层原理、Q5-Q10 K8s 核心 | 中级/高级 |
| ⭐⭐⭐⭐ | Q12 完整部署方案设计 | 高级/架构师 |

## 参考资料

- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Docker 官方文档](https://docs.docker.com/)
- [Go signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
- [Helm 官方文档](https://helm.sh/docs/)
