package config

import (
	"os"
	"strconv"
	"time"
)

// Config アプリケーションの全設定を保持する構造体
type Config struct {
	DB     DBConfig
	JWT    JWTConfig
	Server ServerConfig
}

// DBConfig データベース設定を保持する構造体
type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig JWT設定を保持する構造体
type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

// ServerConfig HTTPサーバー設定を保持する構造体
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// Load 環境変数から設定を読み込む
func Load() *Config {
	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		expiryHours = 24
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		maxOpenConns = 25
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		maxIdleConns = 5
	}

	return &Config{
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "3306"),
			User:            getEnv("DB_USER", "nocode"),
			Password:        getEnv("DB_PASSWORD", "nocodepassword"),
			Name:            getEnv("DB_NAME", "nocode-app"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: 5 * time.Minute,
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "default-secret-key-change-in-production"),
			ExpiryHours: expiryHours,
		},
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			AllowedOrigins:  parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		},
	}
}

// DSN MySQLのデータソース名を返す
func (c *DBConfig) DSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// getEnv 環境変数を取得し、未設定の場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseOrigins カンマ区切りのオリジン文字列をスライスに変換する
func parseOrigins(origins string) []string {
	if origins == "" {
		return []string{"http://localhost:3000"}
	}
	// カンマで分割
	result := make([]string, 0)
	start := 0
	for i := 0; i <= len(origins); i++ {
		if i == len(origins) || origins[i] == ',' {
			if start < i {
				result = append(result, origins[start:i])
			}
			start = i + 1
		}
	}
	return result
}
