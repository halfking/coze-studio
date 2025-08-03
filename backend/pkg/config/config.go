package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
	RBAC     RBACConfig     `yaml:"rbac"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"` // debug, info, warn, error
}

// RBACConfig RBAC配置
type RBACConfig struct {
	CacheTTL struct {
		UserPermissions  time.Duration `yaml:"user_permissions"`
		RolePermissions  time.Duration `yaml:"role_permissions"`
		WorkspaceMembers time.Duration `yaml:"workspace_members"`
	} `yaml:"cache_ttl"`
	PerformanceThreshold time.Duration `yaml:"performance_threshold"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret         string        `yaml:"secret"`
	ExpirationTime time.Duration `yaml:"expiration_time"`
	RefreshTime    time.Duration `yaml:"refresh_time"`
	Issuer         string        `yaml:"issuer"`
}

// Load 加载配置
func Load() (*Config, error) {
	config := &Config{}

	// 尝试从配置文件加载
	if err := loadFromFile(config); err != nil {
		// 如果文件不存在，使用环境变量和默认值
		loadFromEnv(config)
	}

	// 验证配置
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromFile 从配置文件加载
func loadFromFile(config *Config) error {
	configFile := getEnv("CONFIG_FILE", "configs/config.yaml")
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv 从环境变量加载
func loadFromEnv(config *Config) {
	// 服务器配置
	config.Server.Port = getEnvInt("SERVER_PORT", 8080)
	config.Server.Mode = getEnv("SERVER_MODE", "debug")

	// 数据库配置
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnvInt("DB_PORT", 5432)
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "")
	config.Database.DBName = getEnv("DB_NAME", "cozerights")
	config.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	// Redis配置
	config.Redis.Host = getEnv("REDIS_HOST", "localhost")
	config.Redis.Port = getEnvInt("REDIS_PORT", 6379)
	config.Redis.Password = getEnv("REDIS_PASSWORD", "")
	config.Redis.DB = getEnvInt("REDIS_DB", 0)

	// 日志配置
	config.Log.Level = getEnv("LOG_LEVEL", "info")

	// RBAC配置
	config.RBAC.CacheTTL.UserPermissions = getEnvDuration("RBAC_USER_CACHE_TTL", 5*time.Minute)
	config.RBAC.CacheTTL.RolePermissions = getEnvDuration("RBAC_ROLE_CACHE_TTL", 30*time.Minute)
	config.RBAC.CacheTTL.WorkspaceMembers = getEnvDuration("RBAC_WORKSPACE_CACHE_TTL", 10*time.Minute)
	config.RBAC.PerformanceThreshold = getEnvDuration("RBAC_PERFORMANCE_THRESHOLD", 100*time.Millisecond)

	// JWT配置
	config.JWT.Secret = getEnv("JWT_SECRET", "your-secret-key")
	config.JWT.ExpirationTime = getEnvDuration("JWT_EXPIRATION", 24*time.Hour)
	config.JWT.RefreshTime = getEnvDuration("JWT_REFRESH_TIME", 7*24*time.Hour)
	config.JWT.Issuer = getEnv("JWT_ISSUER", "cozerights")
}

// validate 验证配置
func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if config.JWT.Secret == "" || config.JWT.Secret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数类型的环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration 获取时间间隔类型的环境变量
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			User:    "postgres",
			DBName:  "cozerights",
			SSLMode: "disable",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		Log: LogConfig{
			Level: "info",
		},
		RBAC: RBACConfig{
			CacheTTL: struct {
				UserPermissions  time.Duration `yaml:"user_permissions"`
				RolePermissions  time.Duration `yaml:"role_permissions"`
				WorkspaceMembers time.Duration `yaml:"workspace_members"`
			}{
				UserPermissions:  5 * time.Minute,
				RolePermissions:  30 * time.Minute,
				WorkspaceMembers: 10 * time.Minute,
			},
			PerformanceThreshold: 100 * time.Millisecond,
		},
		JWT: JWTConfig{
			Secret:         "change-me-in-production",
			ExpirationTime: 24 * time.Hour,
			RefreshTime:    7 * 24 * time.Hour,
			Issuer:         "cozerights",
		},
	}
}
