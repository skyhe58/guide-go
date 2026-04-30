// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 综合练习：命令行 TODO 工具
// 整合知识点：结构体、切片、map、指针、错误处理、函数、闭包、字符串处理、控制流、iota
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ========== 常量与枚举（iota） ==========

// Priority 任务优先级
type Priority int

const (
	PriorityLow    Priority = iota // 0
	PriorityMedium                 // 1
	PriorityHigh                   // 2
)

// String 实现 fmt.Stringer 接口
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "低"
	case PriorityMedium:
		return "中"
	case PriorityHigh:
		return "高"
	default:
		return "未知"
	}
}

// Status 任务状态
type Status int

const (
	StatusPending Status = iota
	StatusDone
)

func (s Status) String() string {
	if s == StatusDone {
		return "✅"
	}
	return "⬜"
}

// ========== 错误定义 ==========

var (
	ErrNotFound     = errors.New("任务不存在")
	ErrInvalidInput = errors.New("无效输入")
	ErrEmptyTitle   = errors.New("任务标题不能为空")
)

// ========== 结构体定义 ==========

// Task 任务（结构体 + 方法）
type Task struct {
	ID        int
	Title     string
	Priority  Priority
	Status    Status
	CreatedAt time.Time
	Tags      []string // 切片
}

// String 格式化输出
func (t *Task) String() string {
	tags := ""
	if len(t.Tags) > 0 {
		tags = " [" + strings.Join(t.Tags, ", ") + "]"
	}
	return fmt.Sprintf("%s #%d [优先级:%s] %s%s (%s)",
		t.Status, t.ID, t.Priority, t.Title, tags,
		t.CreatedAt.Format("01-02 15:04"))
}

// MarkDone 标记完成（指针接收者：修改原值）
func (t *Task) MarkDone() {
	t.Status = StatusDone
}

// HasTag 检查是否包含标签
func (t *Task) HasTag(tag string) bool {
	for _, tg := range t.Tags {
		if strings.EqualFold(tg, tag) {
			return true
		}
	}
	return false
}

// TodoList TODO 列表管理器
type TodoList struct {
	tasks  []*Task          // 切片存储任务（指针切片避免拷贝）
	nextID int              // 自增 ID
	index  map[int]*Task    // map 索引，O(1) 查找
}

// NewTodoList 构造函数
func NewTodoList() *TodoList {
	return &TodoList{
		tasks:  make([]*Task, 0),
		nextID: 1,
		index:  make(map[int]*Task),
	}
}

// Add 添加任务（错误处理）
func (tl *TodoList) Add(title string, priority Priority, tags []string) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}

	task := &Task{
		ID:        tl.nextID,
		Title:     title,
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		Tags:      tags,
	}

	tl.tasks = append(tl.tasks, task)
	tl.index[task.ID] = task
	tl.nextID++
	return task, nil
}

// Get 获取任务（comma ok 模式）
func (tl *TodoList) Get(id int) (*Task, error) {
	task, ok := tl.index[id]
	if !ok {
		return nil, fmt.Errorf("ID=%d: %w", id, ErrNotFound)
	}
	return task, nil
}

// Done 完成任务
func (tl *TodoList) Done(id int) error {
	task, err := tl.Get(id)
	if err != nil {
		return err
	}
	task.MarkDone()
	return nil
}

// Delete 删除任务（切片删除操作）
func (tl *TodoList) Delete(id int) error {
	if _, ok := tl.index[id]; !ok {
		return fmt.Errorf("ID=%d: %w", id, ErrNotFound)
	}

	// 从切片中删除
	for i, t := range tl.tasks {
		if t.ID == id {
			tl.tasks = append(tl.tasks[:i], tl.tasks[i+1:]...)
			break
		}
	}
	delete(tl.index, id)
	return nil
}

// List 列出任务（闭包过滤器）
func (tl *TodoList) List(filter func(*Task) bool) []*Task {
	var result []*Task
	for _, t := range tl.tasks {
		if filter == nil || filter(t) {
			result = append(result, t)
		}
	}
	return result
}

// Stats 统计信息（多返回值）
func (tl *TodoList) Stats() (total, done, pending int) {
	total = len(tl.tasks)
	for _, t := range tl.tasks {
		if t.Status == StatusDone {
			done++
		}
	}
	pending = total - done
	return
}

// Search 搜索任务（字符串处理）
func (tl *TodoList) Search(keyword string) []*Task {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	return tl.List(func(t *Task) bool {
		return strings.Contains(strings.ToLower(t.Title), keyword)
	})
}

// ========== 命令行交互 ==========

func main() {
	fmt.Println("========================================")
	fmt.Println("  📝 Go TODO CLI — 命令行任务管理工具")
	fmt.Println("  综合练习：整合 Go 基础语法所有知识点")
	fmt.Println("========================================")

	todoList := NewTodoList()

	// 添加示例数据
	addSampleData(todoList)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n--- 命令菜单 ---")
		fmt.Println("  add    - 添加任务")
		fmt.Println("  list   - 列出所有任务")
		fmt.Println("  done   - 完成任务")
		fmt.Println("  delete - 删除任务")
		fmt.Println("  search - 搜索任务")
		fmt.Println("  stats  - 统计信息")
		fmt.Println("  filter - 按条件筛选")
		fmt.Println("  quit   - 退出")
		fmt.Print("\n请输入命令: ")

		if !scanner.Scan() {
			break
		}
		cmd := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch cmd {
		case "add":
			handleAdd(scanner, todoList)
		case "list":
			handleList(todoList)
		case "done":
			handleDone(scanner, todoList)
		case "delete":
			handleDelete(scanner, todoList)
		case "search":
			handleSearch(scanner, todoList)
		case "stats":
			handleStats(todoList)
		case "filter":
			handleFilter(scanner, todoList)
		case "quit", "q", "exit":
			fmt.Println("👋 再见！")
			return
		case "":
			continue
		default:
			fmt.Printf("❌ 未知命令: %q\n", cmd)
		}
	}
}

func addSampleData(tl *TodoList) {
	tl.Add("学习 Go 基础语法", PriorityHigh, []string{"学习", "Go"})
	tl.Add("完成切片练习", PriorityMedium, []string{"练习", "Go"})
	tl.Add("阅读 Effective Go", PriorityLow, []string{"阅读"})
	fmt.Println("📌 已添加 3 条示例任务")
}

func handleAdd(scanner *bufio.Scanner, tl *TodoList) {
	fmt.Print("任务标题: ")
	if !scanner.Scan() {
		return
	}
	title := scanner.Text()

	fmt.Print("优先级 (0=低, 1=中, 2=高): ")
	if !scanner.Scan() {
		return
	}
	pNum, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || pNum < 0 || pNum > 2 {
		pNum = 1 // 默认中优先级
	}

	fmt.Print("标签 (逗号分隔，可留空): ")
	if !scanner.Scan() {
		return
	}
	var tags []string
	tagStr := strings.TrimSpace(scanner.Text())
	if tagStr != "" {
		for _, t := range strings.Split(tagStr, ",") {
			if tag := strings.TrimSpace(t); tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	task, err := tl.Add(title, Priority(pNum), tags)
	if err != nil {
		// 错误处理：errors.Is 判断
		if errors.Is(err, ErrEmptyTitle) {
			fmt.Println("❌ 标题不能为空")
		} else {
			fmt.Printf("❌ 添加失败: %v\n", err)
		}
		return
	}
	fmt.Printf("✅ 已添加: %s\n", task)
}

func handleList(tl *TodoList) {
	tasks := tl.List(nil)
	if len(tasks) == 0 {
		fmt.Println("📭 暂无任务")
		return
	}
	fmt.Printf("\n📋 任务列表 (%d 条):\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("  %s\n", t)
	}
}

func handleDone(scanner *bufio.Scanner, tl *TodoList) {
	fmt.Print("任务 ID: ")
	if !scanner.Scan() {
		return
	}
	id, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil {
		fmt.Println("❌ 请输入有效的数字 ID")
		return
	}
	if err := tl.Done(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			fmt.Printf("❌ 任务 #%d 不存在\n", id)
		} else {
			fmt.Printf("❌ 操作失败: %v\n", err)
		}
		return
	}
	fmt.Printf("✅ 任务 #%d 已完成\n", id)
}

func handleDelete(scanner *bufio.Scanner, tl *TodoList) {
	fmt.Print("任务 ID: ")
	if !scanner.Scan() {
		return
	}
	id, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil {
		fmt.Println("❌ 请输入有效的数字 ID")
		return
	}
	if err := tl.Delete(id); err != nil {
		fmt.Printf("❌ 删除失败: %v\n", err)
		return
	}
	fmt.Printf("🗑️ 任务 #%d 已删除\n", id)
}

func handleSearch(scanner *bufio.Scanner, tl *TodoList) {
	fmt.Print("搜索关键词: ")
	if !scanner.Scan() {
		return
	}
	keyword := scanner.Text()
	results := tl.Search(keyword)
	if len(results) == 0 {
		fmt.Printf("🔍 未找到包含 %q 的任务\n", keyword)
		return
	}
	fmt.Printf("🔍 找到 %d 条匹配:\n", len(results))
	for _, t := range results {
		fmt.Printf("  %s\n", t)
	}
}

func handleStats(tl *TodoList) {
	total, done, pending := tl.Stats()
	fmt.Println("\n📊 统计信息:")
	fmt.Printf("  总计: %d 条\n", total)
	fmt.Printf("  已完成: %d 条\n", done)
	fmt.Printf("  待完成: %d 条\n", pending)
	if total > 0 {
		pct := float64(done) / float64(total) * 100
		// 进度条
		bar := strings.Repeat("█", done) + strings.Repeat("░", pending)
		fmt.Printf("  进度: [%s] %.0f%%\n", bar, pct)
	}
}

func handleFilter(scanner *bufio.Scanner, tl *TodoList) {
	fmt.Println("筛选条件:")
	fmt.Println("  1 - 待完成任务")
	fmt.Println("  2 - 已完成任务")
	fmt.Println("  3 - 高优先级任务")
	fmt.Print("选择: ")
	if !scanner.Scan() {
		return
	}

	var filter func(*Task) bool
	var label string

	// switch 无条件匹配
	choice := strings.TrimSpace(scanner.Text())
	switch choice {
	case "1":
		filter = func(t *Task) bool { return t.Status == StatusPending }
		label = "待完成"
	case "2":
		filter = func(t *Task) bool { return t.Status == StatusDone }
		label = "已完成"
	case "3":
		filter = func(t *Task) bool { return t.Priority == PriorityHigh }
		label = "高优先级"
	default:
		fmt.Println("❌ 无效选择")
		return
	}

	results := tl.List(filter)
	if len(results) == 0 {
		fmt.Printf("📭 没有%s的任务\n", label)
		return
	}
	fmt.Printf("\n📋 %s任务 (%d 条):\n", label, len(results))
	for _, t := range results {
		fmt.Printf("  %s\n", t)
	}
}
