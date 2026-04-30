import type { DefaultTheme } from 'vitepress'

/**
 * 侧边栏配置 — 按五层学习架构组织
 */

// 第一层：语言核心
const goCoreSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '1.1 Go 基础语法',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.1-go-basics/' },
      { text: '环境搭建', link: '/1-go-core/1.1-go-basics/01-environment' },
      { text: '数据类型', link: '/1-go-core/1.1-go-basics/02-data-types' },
      { text: '变量、常量与作用域', link: '/1-go-core/1.1-go-basics/03-variables' },
      { text: '运算符', link: '/1-go-core/1.1-go-basics/04-operators' },
      { text: '控制流', link: '/1-go-core/1.1-go-basics/05-control-flow' },
      { text: '函数', link: '/1-go-core/1.1-go-basics/06-functions' },
      { text: '错误处理', link: '/1-go-core/1.1-go-basics/07-error-handling' },
      { text: '结构体与方法', link: '/1-go-core/1.1-go-basics/08-struct-method' },
      { text: '数组与切片', link: '/1-go-core/1.1-go-basics/09-slice' },
      { text: 'Map', link: '/1-go-core/1.1-go-basics/10-map' },
      { text: '指针', link: '/1-go-core/1.1-go-basics/11-pointer' },
      { text: '包管理与 Go Module', link: '/1-go-core/1.1-go-basics/12-module' },
      { text: '字符串处理', link: '/1-go-core/1.1-go-basics/13-string' },
      { text: '面试指南', link: '/1-go-core/1.1-go-basics/interview' },
    ],
  },
  {
    text: '1.2 Go 进阶特性',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.2-go-advanced/' },
      { text: '接口', link: '/1-go-core/1.2-go-advanced/01-interfaces' },
      { text: '组合与嵌入', link: '/1-go-core/1.2-go-advanced/02-composition' },
      { text: '反射', link: '/1-go-core/1.2-go-advanced/03-reflection' },
      { text: '泛型', link: '/1-go-core/1.2-go-advanced/04-generics' },
      { text: 'unsafe 包', link: '/1-go-core/1.2-go-advanced/05-unsafe' },
      { text: '代码生成', link: '/1-go-core/1.2-go-advanced/06-codegen' },
      { text: '构建标签', link: '/1-go-core/1.2-go-advanced/07-build-tags' },
      { text: '面试指南', link: '/1-go-core/1.2-go-advanced/interview' },
    ],
  },
  {
    text: '1.3 并发编程',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.3-concurrent/' },
      { text: 'goroutine', link: '/1-go-core/1.3-concurrent/01-goroutine' },
      { text: 'channel', link: '/1-go-core/1.3-concurrent/02-channel' },
      { text: 'sync 包', link: '/1-go-core/1.3-concurrent/03-sync' },
      { text: 'context 包', link: '/1-go-core/1.3-concurrent/04-context' },
      { text: '并发模式', link: '/1-go-core/1.3-concurrent/05-patterns' },
      { text: '原子操作', link: '/1-go-core/1.3-concurrent/06-atomic' },
      { text: '数据竞争检测', link: '/1-go-core/1.3-concurrent/07-race' },
      { text: 'errgroup', link: '/1-go-core/1.3-concurrent/08-errgroup' },
      { text: 'semaphore', link: '/1-go-core/1.3-concurrent/09-semaphore' },
      { text: '面试指南', link: '/1-go-core/1.3-concurrent/interview' },
    ],
  },
  {
    text: '1.4 运行时与性能',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.4-runtime/' },
      { text: 'GMP 调度模型', link: '/1-go-core/1.4-runtime/01-gmp' },
      { text: '垃圾回收', link: '/1-go-core/1.4-runtime/02-gc' },
      { text: '内存管理', link: '/1-go-core/1.4-runtime/03-memory' },
      { text: '栈管理', link: '/1-go-core/1.4-runtime/04-stack' },
      { text: 'pprof', link: '/1-go-core/1.4-runtime/05-pprof' },
      { text: 'trace 工具', link: '/1-go-core/1.4-runtime/06-trace' },
      { text: 'benchmark', link: '/1-go-core/1.4-runtime/07-benchmark' },
      { text: '常见优化技巧', link: '/1-go-core/1.4-runtime/08-optimization' },
      { text: '逃逸分析实战', link: '/1-go-core/1.4-runtime/09-escape' },
      { text: '面试指南', link: '/1-go-core/1.4-runtime/interview' },
    ],
  },
  {
    text: '1.5 测试与工具链',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.5-testing/' },
      { text: 'testing 包', link: '/1-go-core/1.5-testing/01-testing' },
      { text: 'benchmark 测试', link: '/1-go-core/1.5-testing/02-benchmark' },
      { text: 'fuzz testing', link: '/1-go-core/1.5-testing/03-fuzz' },
      { text: '测试覆盖率', link: '/1-go-core/1.5-testing/04-coverage' },
      { text: 'Mock 技术', link: '/1-go-core/1.5-testing/05-mock' },
      { text: '集成测试', link: '/1-go-core/1.5-testing/06-integration' },
      { text: 'HTTP 测试', link: '/1-go-core/1.5-testing/07-httptest' },
      { text: '测试最佳实践', link: '/1-go-core/1.5-testing/08-best-practices' },
      { text: 'go vet', link: '/1-go-core/1.5-testing/09-govet' },
      { text: 'golangci-lint', link: '/1-go-core/1.5-testing/10-golangci-lint' },
      { text: '其他工具', link: '/1-go-core/1.5-testing/11-tools' },
      { text: '面试指南', link: '/1-go-core/1.5-testing/interview' },
    ],
  },
  {
    text: '1.6 设计模式与工程化',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.6-patterns/' },
      { text: '创建型模式', link: '/1-go-core/1.6-patterns/01-creational' },
      { text: '结构型模式', link: '/1-go-core/1.6-patterns/02-structural' },
      { text: '行为型模式', link: '/1-go-core/1.6-patterns/03-behavioral' },
      { text: 'Go 特有模式', link: '/1-go-core/1.6-patterns/04-go-patterns' },
      { text: '设计原则', link: '/1-go-core/1.6-patterns/05-principles' },
      { text: '项目布局', link: '/1-go-core/1.6-patterns/06-project-layout' },
      { text: 'Makefile', link: '/1-go-core/1.6-patterns/07-makefile' },
      { text: 'Wire 依赖注入', link: '/1-go-core/1.6-patterns/08-wire' },
      { text: 'Go 版本管理', link: '/1-go-core/1.6-patterns/09-go-version' },
      { text: '错误处理规范', link: '/1-go-core/1.6-patterns/10-error-convention' },
      { text: '面试指南', link: '/1-go-core/1.6-patterns/interview' },
    ],
  },
  {
    text: '1.7 数据结构与算法',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/1-go-core/1.7-algorithm/' },
      { text: '链表', link: '/1-go-core/1.7-algorithm/01-linked-list' },
      { text: '栈与队列', link: '/1-go-core/1.7-algorithm/02-stack-queue' },
      { text: '哈希表', link: '/1-go-core/1.7-algorithm/03-hash-table' },
      { text: '树与二叉树', link: '/1-go-core/1.7-algorithm/04-tree' },
      { text: '堆', link: '/1-go-core/1.7-algorithm/05-heap' },
      { text: '图', link: '/1-go-core/1.7-algorithm/06-graph' },
      { text: '排序算法', link: '/1-go-core/1.7-algorithm/07-sorting' },
      { text: '二分查找', link: '/1-go-core/1.7-algorithm/08-binary-search' },
      { text: '双指针与滑动窗口', link: '/1-go-core/1.7-algorithm/09-two-pointers' },
      { text: '动态规划', link: '/1-go-core/1.7-algorithm/10-dp' },
      { text: '回溯算法', link: '/1-go-core/1.7-algorithm/11-backtracking' },
      { text: '贪心算法', link: '/1-go-core/1.7-algorithm/12-greedy' },
      { text: '面试指南', link: '/1-go-core/1.7-algorithm/interview' },
    ],
  },
]

// 第二层：Web 开发与数据
const webDataSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '2.1 网络编程与 Web 框架',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.1-web-framework/' },
      { text: 'net/http 标准库', link: '/2-web-data/2.1-web-framework/01-net-http' },
      { text: 'TCP 编程', link: '/2-web-data/2.1-web-framework/02-tcp' },
      { text: 'Gin 框架', link: '/2-web-data/2.1-web-framework/03-gin' },
      { text: 'gRPC', link: '/2-web-data/2.1-web-framework/04-grpc' },
      { text: 'WebSocket', link: '/2-web-data/2.1-web-framework/05-websocket' },
      { text: '框架选型对比', link: '/2-web-data/2.1-web-framework/06-comparison' },
      { text: '面试指南', link: '/2-web-data/2.1-web-framework/interview' },
    ],
  },
  {
    text: '2.2 数据库与 ORM',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.2-database/' },
      { text: 'database/sql 标准库', link: '/2-web-data/2.2-database/01-database-sql' },
      { text: 'GORM', link: '/2-web-data/2.2-database/02-gorm' },
      { text: 'sqlx', link: '/2-web-data/2.2-database/03-sqlx' },
      { text: '数据库迁移', link: '/2-web-data/2.2-database/04-migration' },
      { text: 'MySQL 索引原理', link: '/2-web-data/2.2-database/05-mysql-index' },
      { text: 'MySQL 事务与隔离级别', link: '/2-web-data/2.2-database/06-mysql-transaction' },
      { text: 'MySQL 锁机制', link: '/2-web-data/2.2-database/07-mysql-lock' },
      { text: 'SQL 优化与 EXPLAIN', link: '/2-web-data/2.2-database/08-mysql-optimization' },
      { text: 'PostgreSQL', link: '/2-web-data/2.2-database/09-postgresql' },
      { text: 'ORM 选型对比', link: '/2-web-data/2.2-database/10-comparison' },
      { text: '面试指南', link: '/2-web-data/2.2-database/interview' },
    ],
  },
  {
    text: '2.3 缓存与搜索',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.3-cache-search/' },
      { text: 'Redis 数据结构', link: '/2-web-data/2.3-cache-search/01-redis-data-structures' },
      { text: 'Redis 持久化', link: '/2-web-data/2.3-cache-search/02-redis-persistence' },
      { text: 'Redis 主从与哨兵', link: '/2-web-data/2.3-cache-search/03-redis-replication' },
      { text: 'Redis Cluster', link: '/2-web-data/2.3-cache-search/04-redis-cluster' },
      { text: '缓存穿透/击穿/雪崩', link: '/2-web-data/2.3-cache-search/05-redis-cache-problems' },
      { text: '分布式锁', link: '/2-web-data/2.3-cache-search/06-redis-distributed-lock' },
      { text: 'go-redis 客户端', link: '/2-web-data/2.3-cache-search/07-redis-go-client' },
      { text: 'ES 倒排索引', link: '/2-web-data/2.3-cache-search/08-es-inverted-index' },
      { text: 'ES 映射与分析器', link: '/2-web-data/2.3-cache-search/09-es-mapping' },
      { text: 'ES CRUD', link: '/2-web-data/2.3-cache-search/10-es-crud' },
      { text: 'ES DSL 查询', link: '/2-web-data/2.3-cache-search/11-es-dsl' },
      { text: 'ES 聚合分析', link: '/2-web-data/2.3-cache-search/12-es-aggregation' },
      { text: 'go-elasticsearch', link: '/2-web-data/2.3-cache-search/13-es-go-client' },
      { text: '面试指南', link: '/2-web-data/2.3-cache-search/interview' },
    ],
  },
  {
    text: '2.4 消息队列',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.4-message-queue/' },
      { text: 'Kafka', link: '/2-web-data/2.4-message-queue/01-kafka' },
      { text: 'NATS', link: '/2-web-data/2.4-message-queue/02-nats' },
      { text: 'RabbitMQ', link: '/2-web-data/2.4-message-queue/03-rabbitmq' },
      { text: 'MQTT', link: '/2-web-data/2.4-message-queue/04-mqtt' },
      { text: '消息队列选型对比', link: '/2-web-data/2.4-message-queue/05-comparison' },
      { text: '面试指南', link: '/2-web-data/2.4-message-queue/interview' },
    ],
  },
  {
    text: '2.5 对象存储与文档数据库',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.5-object-storage/' },
      { text: 'MinIO', link: '/2-web-data/2.5-object-storage/01-minio' },
      { text: 'MongoDB', link: '/2-web-data/2.5-object-storage/02-mongodb' },
      { text: '面试指南', link: '/2-web-data/2.5-object-storage/interview' },
    ],
  },
  {
    text: '2.6 认证鉴权',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.6-auth/' },
      { text: 'JWT', link: '/2-web-data/2.6-auth/01-jwt' },
      { text: 'OAuth2', link: '/2-web-data/2.6-auth/02-oauth2' },
      { text: 'Keycloak', link: '/2-web-data/2.6-auth/03-keycloak' },
      { text: 'RBAC', link: '/2-web-data/2.6-auth/04-rbac' },
      { text: 'Gin 鉴权中间件', link: '/2-web-data/2.6-auth/05-gin-middleware' },
      { text: 'Casbin', link: '/2-web-data/2.6-auth/06-casbin' },
      { text: '面试指南', link: '/2-web-data/2.6-auth/interview' },
    ],
  },
  {
    text: '2.7 日志与可观测性',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/2-web-data/2.7-observability/' },
      { text: 'slog 标准库', link: '/2-web-data/2.7-observability/01-slog' },
      { text: 'zerolog', link: '/2-web-data/2.7-observability/02-zerolog' },
      { text: 'zap', link: '/2-web-data/2.7-observability/03-zap' },
      { text: '日志库选型对比', link: '/2-web-data/2.7-observability/04-log-comparison' },
      { text: '日志最佳实践', link: '/2-web-data/2.7-observability/05-log-best-practices' },
      { text: 'Sentry', link: '/2-web-data/2.7-observability/06-sentry' },
      { text: 'OpenTelemetry', link: '/2-web-data/2.7-observability/07-otel' },
      { text: 'Prometheus', link: '/2-web-data/2.7-observability/08-prometheus' },
      { text: 'Grafana', link: '/2-web-data/2.7-observability/09-grafana' },
      { text: '分布式链路追踪', link: '/2-web-data/2.7-observability/10-tracing' },
      { text: '面试指南', link: '/2-web-data/2.7-observability/interview' },
    ],
  },
]

// 第三层：微服务与云原生
const microserviceSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '3.1 微服务架构',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/3-microservice/3.1-microservice/' },
      { text: 'Kratos', link: '/3-microservice/3.1-microservice/01-kratos' },
      { text: 'Go-Zero', link: '/3-microservice/3.1-microservice/02-go-zero' },
      { text: 'Go-Kit', link: '/3-microservice/3.1-microservice/03-go-kit' },
      { text: '框架选型对比', link: '/3-microservice/3.1-microservice/04-comparison' },
      { text: '选型指南', link: '/3-microservice/3.1-microservice/05-selection-guide' },
      { text: '面试指南', link: '/3-microservice/3.1-microservice/interview' },
    ],
  },
  {
    text: '3.2 服务治理',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/3-microservice/3.2-service-governance/' },
      { text: 'etcd', link: '/3-microservice/3.2-service-governance/01-etcd' },
      { text: 'Consul', link: '/3-microservice/3.2-service-governance/02-consul' },
      { text: '注册中心选型对比', link: '/3-microservice/3.2-service-governance/03-registry-comparison' },
      { text: 'Viper', link: '/3-microservice/3.2-service-governance/04-viper' },
      { text: 'etcd 配置中心', link: '/3-microservice/3.2-service-governance/05-etcd-config' },
      { text: '配置管理最佳实践', link: '/3-microservice/3.2-service-governance/06-config-best-practices' },
      { text: '面试指南', link: '/3-microservice/3.2-service-governance/interview' },
    ],
  },
  {
    text: '3.3 容器化与 K8s',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/3-microservice/3.3-docker-k8s/' },
      { text: 'Docker 核心概念', link: '/3-microservice/3.3-docker-k8s/01-docker-basics' },
      { text: 'Go 应用 Dockerfile', link: '/3-microservice/3.3-docker-k8s/02-dockerfile' },
      { text: 'Docker Compose', link: '/3-microservice/3.3-docker-k8s/03-docker-compose' },
      { text: 'Docker 网络与数据卷', link: '/3-microservice/3.3-docker-k8s/04-docker-network' },
      { text: 'K8s 架构', link: '/3-microservice/3.3-docker-k8s/05-k8s-architecture' },
      { text: 'K8s 核心资源', link: '/3-microservice/3.3-docker-k8s/06-k8s-resources' },
      { text: 'Go 应用 K8s 部署', link: '/3-microservice/3.3-docker-k8s/07-k8s-go-deploy' },
      { text: 'HPA 自动扩缩容', link: '/3-microservice/3.3-docker-k8s/08-k8s-hpa' },
      { text: 'Helm', link: '/3-microservice/3.3-docker-k8s/09-helm' },
      { text: '命令速查表', link: '/3-microservice/3.3-docker-k8s/10-cheatsheet' },
      { text: '面试指南', link: '/3-microservice/3.3-docker-k8s/interview' },
    ],
  },
  {
    text: '3.4 云服务集成 AWS',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/3-microservice/3.4-aws/' },
      { text: 'AWS SDK 基础', link: '/3-microservice/3.4-aws/01-sdk-basics' },
      { text: '认证方式', link: '/3-microservice/3.4-aws/02-auth' },
      { text: '本地开发配置', link: '/3-microservice/3.4-aws/03-local-dev' },
      { text: 'S3', link: '/3-microservice/3.4-aws/04-s3' },
      { text: 'SQS', link: '/3-microservice/3.4-aws/05-sqs' },
      { text: 'ECR', link: '/3-microservice/3.4-aws/06-ecr' },
      { text: 'STS', link: '/3-microservice/3.4-aws/07-sts' },
      { text: 'IoT Core', link: '/3-microservice/3.4-aws/08-iot-core' },
      { text: 'KVS', link: '/3-microservice/3.4-aws/09-kvs' },
      { text: 'AWS vs 开源对比', link: '/3-microservice/3.4-aws/10-comparison' },
      { text: '面试指南', link: '/3-microservice/3.4-aws/interview' },
    ],
  },
]

// 第四层：分布式与架构
const distributedSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '4.1 分布式系统',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/4-distributed/4.1-distributed/' },
      { text: 'CAP 与 BASE 理论', link: '/4-distributed/4.1-distributed/01-cap-base' },
      { text: 'Raft 一致性算法', link: '/4-distributed/4.1-distributed/02-raft' },
      { text: '分布式锁', link: '/4-distributed/4.1-distributed/03-distributed-lock' },
      { text: '分布式事务', link: '/4-distributed/4.1-distributed/04-distributed-transaction' },
      { text: '幂等性设计', link: '/4-distributed/4.1-distributed/05-idempotent' },
      { text: '限流算法', link: '/4-distributed/4.1-distributed/06-rate-limiting' },
      { text: '熔断与降级', link: '/4-distributed/4.1-distributed/07-circuit-breaker' },
      { text: '面试指南', link: '/4-distributed/4.1-distributed/interview' },
    ],
  },
  {
    text: '4.2 架构设计场景',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/4-distributed/4.2-architecture/' },
      { text: '秒杀系统', link: '/4-distributed/4.2-architecture/01-seckill' },
      { text: '短链接系统', link: '/4-distributed/4.2-architecture/02-short-url' },
      { text: '订单超时取消', link: '/4-distributed/4.2-architecture/03-order-timeout' },
      { text: '缓存一致性', link: '/4-distributed/4.2-architecture/04-cache-consistency' },
      { text: '接口幂等性', link: '/4-distributed/4.2-architecture/05-idempotent-design' },
      { text: '大文件上传', link: '/4-distributed/4.2-architecture/06-file-upload' },
      { text: '一致性哈希', link: '/4-distributed/4.2-architecture/07-consistent-hash' },
      { text: '面试指南', link: '/4-distributed/4.2-architecture/interview' },
    ],
  },
  {
    text: '4.3 AI 应用',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/4-distributed/4.3-ai/' },
      { text: 'OpenAI API', link: '/4-distributed/4.3-ai/01-openai' },
      { text: 'Prompt Engineering', link: '/4-distributed/4.3-ai/02-prompt' },
      { text: 'RAG', link: '/4-distributed/4.3-ai/03-rag' },
      { text: 'AI Agent', link: '/4-distributed/4.3-ai/04-agent' },
      { text: 'Go LLM 生态', link: '/4-distributed/4.3-ai/05-ecosystem' },
      { text: '面试指南', link: '/4-distributed/4.3-ai/interview' },
    ],
  },
]

// 第五层：运维与部署
const devopsSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '5.1 CI/CD 与 DevOps',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/5-devops/5.1-cicd/' },
      { text: 'GitHub Actions', link: '/5-devops/5.1-cicd/01-github-actions' },
      { text: 'GoReleaser', link: '/5-devops/5.1-cicd/02-goreleaser' },
      { text: 'Makefile', link: '/5-devops/5.1-cicd/03-makefile' },
      { text: '面试指南', link: '/5-devops/5.1-cicd/interview' },
    ],
  },
  {
    text: '5.2 Linux 运维',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/5-devops/5.2-linux/' },
      { text: '常用命令', link: '/5-devops/5.2-linux/01-commands' },
      { text: 'Shell 脚本', link: '/5-devops/5.2-linux/02-shell' },
      { text: '性能排查工具', link: '/5-devops/5.2-linux/03-performance' },
      { text: '日志分析', link: '/5-devops/5.2-linux/04-log-analysis' },
      { text: 'Go 服务部署', link: '/5-devops/5.2-linux/05-go-deploy' },
      { text: '线上问题排查', link: '/5-devops/5.2-linux/06-troubleshooting' },
      { text: 'Go 服务监控', link: '/5-devops/5.2-linux/07-monitoring' },
      { text: '面试指南', link: '/5-devops/5.2-linux/interview' },
    ],
  },
  {
    text: '5.3 Nginx 与反向代理',
    collapsed: true,
    items: [
      { text: '模块概览', link: '/5-devops/5.3-nginx/' },
      { text: 'Nginx 架构', link: '/5-devops/5.3-nginx/01-architecture' },
      { text: '反向代理配置', link: '/5-devops/5.3-nginx/02-reverse-proxy' },
      { text: '负载均衡', link: '/5-devops/5.3-nginx/03-load-balancing' },
      { text: 'HTTPS 配置', link: '/5-devops/5.3-nginx/04-https' },
      { text: '限流与防刷', link: '/5-devops/5.3-nginx/05-rate-limiting' },
      { text: '跨域与 Gzip', link: '/5-devops/5.3-nginx/06-cors-gzip' },
      { text: '日志分析', link: '/5-devops/5.3-nginx/07-log-analysis' },
      { text: '面试指南', link: '/5-devops/5.3-nginx/interview' },
    ],
  },
]

// 面试汇总
const interviewSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '面试汇总',
    items: [
      { text: '面试知识图谱', link: '/interview/knowledge-map' },
      { text: '按公司类型分类', link: '/interview/by-company' },
    ],
  },
]

// 学习路径
const learningPathSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '学习路径',
    items: [
      { text: 'Go 初学者路径', link: '/learning-paths/beginner' },
      { text: 'Go 中级进阶路径', link: '/learning-paths/intermediate' },
      { text: 'Go 高级深入路径', link: '/learning-paths/advanced' },
      { text: '面试突击路径', link: '/learning-paths/interview-sprint' },
      { text: '云原生工程师路径', link: '/learning-paths/cloud-native' },
    ],
  },
]

// 全栈实战项目
const fullstackSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '6.0 GoBlog 全栈实战',
    collapsed: true,
    items: [
      { text: '项目概览', link: '/6-fullstack-project/' },
      { text: '技术选型说明', link: '/6-fullstack-project/01-tech-stack' },
      { text: '项目初始化与环境搭建', link: '/6-fullstack-project/02-project-setup' },
      { text: '用户模块实现', link: '/6-fullstack-project/03-user-module' },
      { text: '文章模块实现', link: '/6-fullstack-project/04-article-module' },
      { text: '标签与评论模块', link: '/6-fullstack-project/05-tag-comment-module' },
      { text: '中间件链实现', link: '/6-fullstack-project/06-middleware' },
      { text: '缓存策略实现', link: '/6-fullstack-project/07-cache-strategy' },
      { text: '测试策略与实现', link: '/6-fullstack-project/08-testing' },
      { text: 'Docker 部署', link: '/6-fullstack-project/09-deployment' },
      { text: 'API 接口参考', link: '/6-fullstack-project/api-reference' },
      { text: '面试指南', link: '/6-fullstack-project/interview' },
    ],
  },
]

/**
 * 导出侧边栏配置
 * 使用路径前缀匹配，不同路径显示不同侧边栏
 */
export const sidebar: DefaultTheme.Sidebar = {
  '/1-go-core/': goCoreSidebar,
  '/2-web-data/': webDataSidebar,
  '/3-microservice/': microserviceSidebar,
  '/4-distributed/': distributedSidebar,
  '/5-devops/': devopsSidebar,
  '/interview/': interviewSidebar,
  '/learning-paths/': learningPathSidebar,
  '/6-fullstack-project/': fullstackSidebar,
  '/guide/': [
    {
      text: '指南',
      items: [
        { text: '快速开始', link: '/guide/getting-started' },
        { text: '使用指南', link: '/guide/how-to-use' },
      ],
    },
  ],
}
