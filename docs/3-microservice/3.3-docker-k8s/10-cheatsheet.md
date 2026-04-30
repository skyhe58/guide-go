---
title: "Docker 与 K8s 常用命令速查表"
module: "docker-k8s"
difficulty: "beginner"
interviewFrequency: "medium"
tags:
  - Docker
  - Kubernetes
  - 速查表
  - 命令行
codeExample: "03-microservice/docker-k8s/"
relatedEntries:
  - "/3-microservice/3.3-docker-k8s/01-docker-basics"
  - "/3-microservice/3.3-docker-k8s/05-k8s-architecture"
prerequisites:
  - "/3-microservice/3.3-docker-k8s/01-docker-basics"
estimatedTime: "随时查阅"
---

# Docker 与 K8s 常用命令速查表

## Docker 命令

### 镜像管理

```bash
# 构建镜像
docker build -t my-go-app:v1.0.0 .
docker build -t my-go-app:v1.0.0 -f Dockerfile.prod .

# 查看本地镜像
docker images
docker image ls

# 拉取/推送镜像
docker pull golang:1.22-alpine
docker push my-registry/my-go-app:v1.0.0

# 给镜像打标签
docker tag my-go-app:v1.0.0 my-registry/my-go-app:v1.0.0

# 删除镜像
docker rmi my-go-app:v1.0.0
docker image prune          # 删除悬空镜像
docker image prune -a       # 删除所有未使用的镜像

# 查看镜像分层
docker history my-go-app:v1.0.0

# 查看镜像详情
docker inspect my-go-app:v1.0.0
```

### 容器管理

```bash
# 运行容器
docker run -d --name my-app -p 8080:8080 my-go-app:v1.0.0
docker run -it --rm golang:1.22-alpine sh    # 交互式运行

# 查看运行中的容器
docker ps
docker ps -a                # 包含已停止的容器

# 停止/启动/重启容器
docker stop my-app
docker start my-app
docker restart my-app

# 删除容器
docker rm my-app
docker rm -f my-app         # 强制删除运行中的容器
docker container prune       # 删除所有已停止的容器

# 查看容器日志
docker logs my-app
docker logs -f my-app        # 实时跟踪日志
docker logs --tail 100 my-app # 最后 100 行

# 进入容器
docker exec -it my-app sh
docker exec -it my-app /bin/bash

# 查看容器资源使用
docker stats
docker stats my-app

# 复制文件
docker cp my-app:/app/config.yaml ./config.yaml
docker cp ./config.yaml my-app:/app/config.yaml
```

### Docker Compose

```bash
# 启动所有服务
docker compose up -d
docker compose -f docker-compose.yml up -d

# 停止所有服务
docker compose down
docker compose down -v       # 同时删除数据卷

# 查看服务状态
docker compose ps

# 查看服务日志
docker compose logs
docker compose logs -f app   # 跟踪特定服务日志

# 重新构建并启动
docker compose up -d --build

# 扩缩容
docker compose up -d --scale app=3

# 执行命令
docker compose exec app sh
```

### 网络与数据卷

```bash
# 网络管理
docker network ls
docker network create my-net
docker network inspect my-net
docker network connect my-net my-app
docker network disconnect my-net my-app

# 数据卷管理
docker volume ls
docker volume create my-data
docker volume inspect my-data
docker volume rm my-data
docker volume prune          # 删除未使用的数据卷
```

### 系统清理

```bash
# 查看磁盘使用
docker system df

# 全面清理（慎用）
docker system prune          # 删除停止的容器、悬空镜像、未使用的网络
docker system prune -a       # 额外删除所有未使用的镜像
docker system prune --volumes # 额外删除未使用的数据卷
```

## Kubernetes 命令

### 集群信息

```bash
# 查看集群信息
kubectl cluster-info
kubectl version

# 查看节点
kubectl get nodes
kubectl describe node <node-name>
kubectl top nodes            # 节点资源使用（需 Metrics Server）
```

### 资源查看

```bash
# 查看资源（通用格式）
kubectl get <resource>
kubectl get <resource> -o wide          # 更多信息
kubectl get <resource> -o yaml          # YAML 格式
kubectl get <resource> -o json          # JSON 格式
kubectl get <resource> -n <namespace>   # 指定命名空间
kubectl get <resource> --all-namespaces # 所有命名空间

# 常用资源查看
kubectl get pods
kubectl get deployments
kubectl get services
kubectl get configmaps
kubectl get secrets
kubectl get ingress
kubectl get hpa
kubectl get events --sort-by='.lastTimestamp'

# 查看资源详情
kubectl describe pod <pod-name>
kubectl describe deployment <deploy-name>
kubectl describe service <svc-name>
```

### 资源创建与管理

```bash
# 声明式管理（推荐）
kubectl apply -f deployment.yaml
kubectl apply -f k8s/                   # 应用目录下所有 YAML
kubectl apply -f https://example.com/manifest.yaml

# 删除资源
kubectl delete -f deployment.yaml
kubectl delete pod <pod-name>
kubectl delete deployment <deploy-name>

# 编辑资源
kubectl edit deployment <deploy-name>

# 扩缩容
kubectl scale deployment <deploy-name> --replicas=5

# 滚动更新
kubectl set image deployment/<deploy-name> app=my-go-app:v2.0.0

# 查看滚动更新状态
kubectl rollout status deployment/<deploy-name>

# 回滚
kubectl rollout undo deployment/<deploy-name>
kubectl rollout undo deployment/<deploy-name> --to-revision=2

# 查看更新历史
kubectl rollout history deployment/<deploy-name>
```

### Pod 调试

```bash
# 查看 Pod 日志
kubectl logs <pod-name>
kubectl logs -f <pod-name>              # 实时跟踪
kubectl logs <pod-name> -c <container>  # 多容器 Pod 指定容器
kubectl logs <pod-name> --previous      # 上一个容器的日志

# 进入 Pod
kubectl exec -it <pod-name> -- sh
kubectl exec -it <pod-name> -c <container> -- sh

# 端口转发（本地调试）
kubectl port-forward <pod-name> 8080:8080
kubectl port-forward svc/<svc-name> 8080:80

# 查看 Pod 资源使用
kubectl top pods
kubectl top pods --sort-by=memory

# 复制文件
kubectl cp <pod-name>:/app/logs ./logs
kubectl cp ./config.yaml <pod-name>:/app/config.yaml
```

### ConfigMap 和 Secret

```bash
# 创建 ConfigMap
kubectl create configmap app-config --from-literal=APP_ENV=production
kubectl create configmap app-config --from-file=config.yaml

# 创建 Secret
kubectl create secret generic app-secret --from-literal=DB_PASSWORD=mypassword
kubectl create secret generic tls-secret --cert=tls.crt --key=tls.key

# 查看 Secret（Base64 解码）
kubectl get secret app-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
```

### Namespace

```bash
# 查看命名空间
kubectl get namespaces

# 创建命名空间
kubectl create namespace staging

# 设置默认命名空间
kubectl config set-context --current --namespace=staging

# 查看当前上下文
kubectl config current-context
kubectl config get-contexts
```

### Helm 命令

```bash
# Chart 管理
helm create my-chart         # 创建新 Chart
helm template my-release ./my-chart  # 渲染模板（不安装）
helm lint ./my-chart         # 检查 Chart 语法

# 安装/升级/卸载
helm install my-release ./my-chart
helm install my-release ./my-chart -f values-prod.yaml
helm upgrade my-release ./my-chart --set image.tag=v2.0.0
helm uninstall my-release

# 查看/回滚
helm list
helm history my-release
helm rollback my-release 1

# 仓库管理
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm search repo bitnami/postgresql
```

## 常用组合命令

```bash
# 查看 Pod 并按重启次数排序
kubectl get pods --sort-by='.status.containerStatuses[0].restartCount'

# 查看所有非 Running 状态的 Pod
kubectl get pods --field-selector=status.phase!=Running

# 强制删除卡住的 Pod
kubectl delete pod <pod-name> --grace-period=0 --force

# 查看某个 Deployment 的所有 Pod
kubectl get pods -l app=my-go-app

# 批量重启 Deployment（触发滚动更新）
kubectl rollout restart deployment/<deploy-name>

# 查看 Service 的 Endpoints
kubectl get endpoints <svc-name>

# 在集群内部测试 Service 连通性
kubectl run test --rm -it --image=busybox -- wget -qO- http://<svc-name>:<port>
```

## 参考资料

- [Docker CLI 参考](https://docs.docker.com/reference/cli/docker/)
- [kubectl 参考](https://kubernetes.io/docs/reference/kubectl/)
- [Helm CLI 参考](https://helm.sh/docs/helm/)
- [kubectl 速查表（官方）](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
