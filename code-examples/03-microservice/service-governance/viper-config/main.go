// Viper 配置管理 — 完整示例
// 演示：YAML 配置读取 / 环境变量覆盖 / 结构体映射 / 配置热更新 / 默认值 / 嵌套配置
// Go 1.22+ | 验证日期 2025-01-01
//
// 纯 Go 实现，无需 Docker
//
// 运行方式：
//   go run ./viper-config/
//
// 环境变量覆盖示例：
//   APP_SERVER_PORT=9090 APP_DATABASE_HOST=db.example.com go run ./viper-config/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ============================================================
// 配置结构体定义
// ============================================================

// AppConfig 应用总配置
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Features FeatureConfig  `mapstructure:"features"`
	App      AppInfo        `mapstructure:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN 生成数据库连接字符串
func (d DatabaseConfig) DSN() string {
	switch d.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.User, d.Password, d.Name)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			d.User, d.Password, d.Host, d.Port, d.Name)
	default:
		return ""
	}
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 生成 Redis 连接地址
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// FeatureConfig 功能开关配置
type FeatureConfig struct {
	EnableCache     bool `mapstructure:"enable_cache"`
	EnableRateLimit bool `mapstructure:"enable_rate_limit"`
	RateLimitQPS    int  `mapstructure:"rate_limit_qps"`
	EnableNewUI     bool `mapstructure:"enable_new_ui"`
	MaintenanceMode bool `mapstructure:"maintenance_mode"`
}

// AppInfo 应用信息
type AppInfo struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

// ============================================================
// 配置管理器（线程安全的配置访问）
// ============================================================

// ConfigManager 线程安全的配置管理器
type ConfigManager struct {
	mu     sync.RWMutex
	config *AppConfig
	v      *viper.Viper
}

// NewConfigManager 创建配置管理器
func NewConfigManager(v *viper.Viper) (*ConfigManager, error) {
	cm := &ConfigManager{v: v}
	if err := cm.reload(); err != nil {
		return nil, err
	}
	return cm, nil
}

// reload 重新加载配置
func (cm *ConfigManager) reload() error {
	var config AppConfig
	if err := cm.v.Unmarshal(&config); err != nil {
		return fmt.Errorf("配置解析失败: %w", err)
	}
	cm.mu.Lock()
	cm.config = &config
	cm.mu.Unlock()
	return nil
}

// Get 获取当前配置（线程安全）
func (cm *ConfigManager) Get() AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return *cm.config
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Viper 配置管理 — 完整示例")
	fmt.Println("=" + strings.Repeat("=", 59))

	// --- 1. 基本配置读取 ---
	demoBasicConfig()

	// --- 2. 环境变量覆盖 ---
	demoEnvOverride()

	// --- 3. 结构体映射 ---
	demoStructMapping()

	// --- 4. 默认值与嵌套配置 ---
	demoDefaultsAndNested()

	// --- 5. 配置热更新 ---
	demoWatchConfig()
}

// demoBasicConfig 演示基本配置读取
func demoBasicConfig() {
	fmt.Println("\n--- 1. 基本配置读取 ---")

	v := viper.New()

	// 设置配置文件路径（相对于运行目录）
	// 查找 viper-config 子目录下的 config.yaml
	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("  ❌ 读取配置文件失败: %v\n", err)
		fmt.Println("  请确保在 service-governance 目录下运行")
		return
	}
	fmt.Printf("  ✅ 配置文件: %s\n", v.ConfigFileUsed())

	// 读取各种类型的配置值
	fmt.Printf("  server.host = %s\n", v.GetString("server.host"))
	fmt.Printf("  server.port = %d\n", v.GetInt("server.port"))
	fmt.Printf("  database.driver = %s\n", v.GetString("database.driver"))
	fmt.Printf("  features.enable_cache = %v\n", v.GetBool("features.enable_cache"))
	fmt.Printf("  features.rate_limit_qps = %d\n", v.GetInt("features.rate_limit_qps"))

	// 检查 key 是否存在
	fmt.Printf("  server.port 存在: %v\n", v.IsSet("server.port"))
	fmt.Printf("  server.unknown 存在: %v\n", v.IsSet("server.unknown"))

	// 获取所有 key
	allKeys := v.AllKeys()
	fmt.Printf("  配置项总数: %d\n", len(allKeys))
}

// demoEnvOverride 演示环境变量覆盖
func demoEnvOverride() {
	fmt.Println("\n--- 2. 环境变量覆盖 ---")

	v := viper.New()

	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.ReadInConfig()

	// 配置环境变量前缀和替换规则
	// 环境变量 APP_SERVER_PORT 对应配置 server.port
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取原始值
	fmt.Printf("  配置文件中 server.port = %d\n", v.GetInt("server.port"))
	fmt.Printf("  配置文件中 database.host = %s\n", v.GetString("database.host"))

	// 模拟设置环境变量
	os.Setenv("APP_SERVER_PORT", "9090")
	os.Setenv("APP_DATABASE_HOST", "db.production.com")
	defer os.Unsetenv("APP_SERVER_PORT")
	defer os.Unsetenv("APP_DATABASE_HOST")

	// 环境变量覆盖后的值
	fmt.Printf("  环境变量覆盖后 server.port = %d\n", v.GetInt("server.port"))
	fmt.Printf("  环境变量覆盖后 database.host = %s\n", v.GetString("database.host"))
	fmt.Println("  ✅ 环境变量优先级高于配置文件（12-Factor App 原则）")

	// 演示配置优先级
	fmt.Println("\n  Viper 配置优先级（从高到低）:")
	fmt.Println("    1. viper.Set() 显式设置")
	fmt.Println("    2. 命令行参数 (pflag)")
	fmt.Println("    3. 环境变量")
	fmt.Println("    4. 配置文件")
	fmt.Println("    5. 远程配置源 (etcd/Consul)")
	fmt.Println("    6. viper.SetDefault() 默认值")

	// 演示 Set 最高优先级
	v.Set("server.port", 3000)
	fmt.Printf("\n  viper.Set() 后 server.port = %d（最高优先级）\n", v.GetInt("server.port"))
}

// demoStructMapping 演示结构体映射
func demoStructMapping() {
	fmt.Println("\n--- 3. 结构体映射 ---")

	v := viper.New()

	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.ReadInConfig()

	// 映射到结构体
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		fmt.Printf("  ❌ 结构体映射失败: %v\n", err)
		return
	}

	fmt.Println("  ✅ 配置已映射到 AppConfig 结构体")
	fmt.Printf("  服务器: %s:%d (模式: %s)\n",
		config.Server.Host, config.Server.Port, config.Server.Mode)
	fmt.Printf("  数据库: %s (DSN: %s)\n",
		config.Database.Driver, config.Database.DSN())
	fmt.Printf("  Redis: %s\n", config.Redis.Addr())
	fmt.Printf("  日志: level=%s, format=%s\n",
		config.Logging.Level, config.Logging.Format)
	fmt.Printf("  功能开关: cache=%v, rate_limit=%v (QPS=%d), new_ui=%v\n",
		config.Features.EnableCache, config.Features.EnableRateLimit,
		config.Features.RateLimitQPS, config.Features.EnableNewUI)
	fmt.Printf("  应用: %s v%s (%s)\n",
		config.App.Name, config.App.Version, config.App.Env)

	// 使用 ConfigManager 实现线程安全访问
	cm, err := NewConfigManager(v)
	if err != nil {
		fmt.Printf("  ❌ 创建 ConfigManager 失败: %v\n", err)
		return
	}
	cfg := cm.Get()
	fmt.Printf("\n  ConfigManager 线程安全读取: server.port=%d\n", cfg.Server.Port)
}

// demoDefaultsAndNested 演示默认值和嵌套配置
func demoDefaultsAndNested() {
	fmt.Println("\n--- 4. 默认值与嵌套配置 ---")

	v := viper.New()

	// 设置默认值（优先级最低）
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("app.name", "default-service")
	v.SetDefault("app.version", "0.0.1")

	// 不读取配置文件，仅使用默认值
	fmt.Println("  仅使用默认值（未读取配置文件）:")
	fmt.Printf("    server.host = %s\n", v.GetString("server.host"))
	fmt.Printf("    server.port = %d\n", v.GetInt("server.port"))
	fmt.Printf("    server.mode = %s\n", v.GetString("server.mode"))
	fmt.Printf("    app.name = %s\n", v.GetString("app.name"))

	// 读取配置文件覆盖默认值
	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err == nil {
		fmt.Println("\n  读取配置文件后（覆盖默认值）:")
		fmt.Printf("    server.mode = %s（配置文件值覆盖默认值 release）\n",
			v.GetString("server.mode"))
		fmt.Printf("    app.name = %s（配置文件值覆盖默认值 default-service）\n",
			v.GetString("app.name"))
	}

	// 嵌套配置的子树提取
	fmt.Println("\n  嵌套配置子树提取:")
	dbSub := v.Sub("database")
	if dbSub != nil {
		fmt.Printf("    database 子树: host=%s, port=%d, driver=%s\n",
			dbSub.GetString("host"), dbSub.GetInt("port"), dbSub.GetString("driver"))
	}
}

// demoWatchConfig 演示配置热更新
func demoWatchConfig() {
	fmt.Println("\n--- 5. 配置热更新（WatchConfig） ---")

	v := viper.New()

	configPath := findConfigPath()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("  ❌ 读取配置文件失败: %v\n", err)
		return
	}

	// 创建线程安全的配置管理器
	cm, err := NewConfigManager(v)
	if err != nil {
		fmt.Printf("  ❌ 创建 ConfigManager 失败: %v\n", err)
		return
	}

	// 设置配置变更回调
	changeDetected := make(chan struct{}, 1)
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("  [热更新] 检测到配置文件变更: %s\n", e.Name)
		if err := cm.reload(); err != nil {
			fmt.Printf("  [热更新] 重新加载失败: %v\n", err)
			return
		}
		cfg := cm.Get()
		fmt.Printf("  [热更新] 新配置: server.port=%d, app.env=%s\n",
			cfg.Server.Port, cfg.App.Env)
		select {
		case changeDetected <- struct{}{}:
		default:
		}
	})

	// 启动文件监听
	v.WatchConfig()
	fmt.Println("  ✅ 已启动配置文件监听")
	fmt.Println("  提示: 修改 viper-config/config.yaml 文件可触发热更新回调")
	fmt.Printf("  当前配置: server.port=%d\n", cm.Get().Server.Port)

	// 等待一小段时间展示热更新能力（实际项目中会持续运行）
	fmt.Println("  等待 2 秒检测配置变更...")
	select {
	case <-changeDetected:
		fmt.Println("  ✅ 配置热更新成功！")
	case <-time.After(2 * time.Second):
		fmt.Println("  （未检测到配置变更，这是正常的——手动修改 config.yaml 可触发）")
	}

	fmt.Println("\n✅ Viper 配置管理演示完成")
}

// findConfigPath 查找配置文件路径
func findConfigPath() string {
	// 尝试多个可能的路径
	candidates := []string{
		"./viper-config",
		".",
	}
	for _, p := range candidates {
		if _, err := os.Stat(filepath.Join(p, "config.yaml")); err == nil {
			return p
		}
	}
	return "./viper-config"
}
