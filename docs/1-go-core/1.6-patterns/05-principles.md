---
title: "设计原则"
module: "design-patterns"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - SOLID
  - 设计原则
  - 组合优于继承
  - Accept Interfaces Return Structs
  - 面向接口编程
codeExample: "01-go-core/design-patterns/"
relatedEntries:
  - "/1-go-core/1.6-patterns/04-go-patterns"
  - "/1-go-core/1.2-go-advanced/01-interfaces"
prerequisites:
  - "/1-go-core/1.2-go-advanced/01-interfaces"
estimatedTime: "40min"
---

# 设计原则

## 概念说明

SOLID 原则是面向对象设计的五大基本原则，虽然 Go 不是传统的面向对象语言，但 SOLID 原则在 Go 中同样适用，只是表现形式不同。Go 还有自己独特的设计原则，如"组合优于继承"和"Accept Interfaces, Return Structs"。

## 核心原理

### SOLID 在 Go 中的体现

#### S — 单一职责原则（Single Responsibility Principle）

一个包/结构体/函数只做一件事。Go 标准库是单一职责的典范。

**实际应用：**
- 标准库 `net/http` 中 `Handler` 接口只有一个方法 `ServeHTTP`
- 标准库 `io.Reader` 只负责读，`io.Writer` 只负责写
- Kubernetes 中每个 Controller 只负责一种资源的协调

```go
// ✅ 好的设计：每个接口职责单一
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }

// 需要时通过组合扩展
type ReadWriter interface {
    Reader
    Writer
}
```

#### O — 开闭原则（Open/Closed Principle）

对扩展开放，对修改关闭。Go 通过接口实现。

**实际应用：**
- 标准库 `database/sql` 通过 `driver.Driver` 接口支持不同数据库驱动，无需修改 sql 包代码
- Docker 的存储驱动、日志驱动、网络驱动都通过接口扩展

```go
// 通过接口扩展，无需修改已有代码
type Notifier interface {
    Notify(message string) error
}

type EmailNotifier struct{}
func (e *EmailNotifier) Notify(msg string) error { /* 发送邮件 */ return nil }

type SlackNotifier struct{}
func (s *SlackNotifier) Notify(msg string) error { /* 发送 Slack */ return nil }

// 新增通知方式只需实现 Notifier 接口，无需修改已有代码
type WeChatNotifier struct{}
func (w *WeChatNotifier) Notify(msg string) error { /* 发送微信 */ return nil }
```

#### L — 里氏替换原则（Liskov Substitution Principle）

实现接口的类型必须能替换接口使用。Go 的隐式接口天然支持。

#### I — 接口隔离原则（Interface Segregation Principle）

不应该强迫客户端依赖它不需要的方法。Go 推崇小接口。

**实际应用：**
- Go 标准库的接口通常只有 1-3 个方法
- `io.Reader`（1 个方法）、`fmt.Stringer`（1 个方法）、`sort.Interface`（3 个方法）
- Kubernetes 中 `Lister` 和 `Watcher` 是分开的接口

```go
// ✅ Go 风格：小接口
type Saver interface {
    Save(data []byte) error
}

type Loader interface {
    Load(key string) ([]byte, error)
}

// 需要时组合
type Storage interface {
    Saver
    Loader
}

// ❌ 反模式：大而全的接口
type Repository interface {
    Save(data []byte) error
    Load(key string) ([]byte, error)
    Delete(key string) error
    List() ([]string, error)
    Count() int
    // ... 更多方法
}
```

#### D — 依赖倒置原则（Dependency Inversion Principle）

高层模块不应依赖低层模块，两者都应依赖抽象。

**实际应用：**
- Kubernetes 中 Controller 依赖 `Lister` 接口而非具体的 API 客户端
- Go 标准库 `io.Copy` 依赖 `Reader` 和 `Writer` 接口

```go
// ✅ 依赖接口
type UserService struct {
    repo UserRepository // 接口，不是具体实现
}

type UserRepository interface {
    FindByID(id int) (*User, error)
    Save(user *User) error
}

// 可以注入不同实现
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 组合优于继承

Go 没有类继承，通过结构体嵌入（embedding）实现组合，这是 Go 最核心的设计哲学之一。

**实际应用：**
- 标准库 `bufio.ReadWriter` 嵌入了 `*Reader` 和 `*Writer`
- Gin 的 `Context` 嵌入了 `http.Request`
- Kubernetes 中 `ObjectMeta` 被嵌入到所有资源类型中

```go
// 组合而非继承
type Animal struct {
    Name string
}

func (a *Animal) Eat() { fmt.Printf("%s 正在吃东西\n", a.Name) }

type Dog struct {
    Animal // 嵌入，获得 Eat() 方法
    Breed  string
}

func (d *Dog) Bark() { fmt.Printf("%s 汪汪叫\n", d.Name) }
```

### Accept Interfaces, Return Structs

这是 Go 社区最重要的设计准则之一：函数参数接受接口，返回具体类型。

```go
// ✅ Go 惯例
func NewServer(logger Logger) *Server {  // 参数是接口，返回是具体类型
    return &Server{logger: logger}
}

// ❌ 反模式
func NewServer(logger *ZapLogger) ServerInterface {  // 参数是具体类型，返回是接口
    return &Server{logger: logger}
}
```

**原因：**
1. 接口由消费者定义，而非提供者定义（与 Java 相反）
2. 返回具体类型让调用者可以访问所有方法，需要时再抽象为接口
3. 避免不必要的接口抽象层

## 常见面试题

### Q1: Go 中如何体现 SOLID 原则？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 逐一说明 SOLID 在 Go 中的体现
2. 重点强调接口隔离（小接口）和依赖倒置（Accept Interfaces）
3. 举标准库的例子

**标准答案**：

Go 通过语言特性天然支持 SOLID：单一职责通过小包和小接口实现（如 `io.Reader` 只有一个方法）；开闭原则通过接口扩展（如 `database/sql` 的驱动机制）；里氏替换通过隐式接口实现；接口隔离是 Go 的核心哲学——推崇 1-3 个方法的小接口；依赖倒置通过"Accept Interfaces, Return Structs"实现。

**深入追问**：

- Go 的隐式接口和 Java 的显式接口各有什么优缺点？
- "Accept Interfaces, Return Structs" 有没有例外情况？

### Q2: 为什么 Go 推崇"组合优于继承"？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. Go 没有类继承，这是语言设计的刻意选择
2. 继承的问题：脆弱基类、菱形继承、紧耦合
3. 组合的优势：灵活、松耦合、易测试

**标准答案**：

Go 的设计者认为继承带来的问题（脆弱基类、菱形继承、深层继承链导致的紧耦合）大于它的便利。Go 通过结构体嵌入实现代码复用，通过接口实现多态。组合的优势是松耦合、易于测试（可以替换组合的组件）、避免了继承层次过深的问题。标准库中 `bufio.ReadWriter` 嵌入 `Reader` 和 `Writer` 就是组合的典范。

**深入追问**：

- 结构体嵌入和继承的本质区别是什么？（嵌入是 has-a，继承是 is-a）
- 嵌入时方法提升的规则是什么？

## 常见陷阱

1. **过度抽象**：不要为了接口而接口，只有在需要多态或测试 Mock 时才定义接口
2. **大接口**：Go 推崇小接口（1-3 个方法），大接口违反接口隔离原则
3. **嵌入导致的方法冲突**：多个嵌入类型有同名方法时，需要显式指定调用哪个

## 参考资料

- [Go Proverbs](https://go-proverbs.github.io/)
- [Effective Go - Embedding](https://go.dev/doc/effective_go#embedding)
- [Jack Lindamood - Accept Interfaces, Return Structs](https://bryanftan.medium.com/accept-interfaces-return-structs-in-go-d4cab29a301b)
- [Dave Cheney - SOLID Go Design](https://dave.cheney.net/2016/08/20/solid-go-design)
