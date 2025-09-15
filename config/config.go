package config

import (
	"fmt"
	"os"
)

// Config структура конфигурации
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Server
	Port        string
	Environment string
}

// NewConfig создает новую конфигурацию
func NewConfig() (*Config, error) {
	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres123"),
		DBName:     getEnv("DB_NAME", "employee"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Server
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}, nil
}

// ValidateConfig проверяет корректность конфигурации
func (c *Config) ValidateConfig() error {
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD обязателен")
	}
	return nil
}

// GetServerAddress возвращает адрес сервера
func (c *Config) GetServerAddress() string {
	return ":" + c.Port
}

// Database interface methods
func (c *Config) GetDBHost() string     { return c.DBHost }
func (c *Config) GetDBPort() string     { return c.DBPort }
func (c *Config) GetDBUser() string     { return c.DBUser }
func (c *Config) GetDBPassword() string { return c.DBPassword }
func (c *Config) GetDBName() string     { return c.DBName }
func (c *Config) GetDBSSLMode() string  { return c.DBSSLMode }

// getEnv получает переменную окружения с значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
