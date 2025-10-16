package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	Server     ServerConfig
	Keycloak   KeycloakConfig
	ClickHouse ClickHouseConfig
	CORS       CORSConfig
}

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Port string
	Host string
}

// KeycloakConfig содержит настройки Keycloak
type KeycloakConfig struct {
	URL      string
	Realm    string
	ClientID string
}

// ClickHouseConfig содержит настройки ClickHouse
type ClickHouseConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// CORSConfig содержит настройки CORS
type CORSConfig struct {
	AllowedOrigins string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем .env если существует
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8000"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Keycloak: KeycloakConfig{
			URL:      getEnv("KEYCLOAK_URL", "http://localhost:8080"),
			Realm:    getEnv("KEYCLOAK_REALM", "reports-realm"),
			ClientID: getEnv("KEYCLOAK_CLIENT_ID", "reports-backend"),
		},
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     getEnv("CLICKHOUSE_PORT", "9000"),
			Database: getEnv("CLICKHOUSE_DATABASE", "reports_db"),
			User:     getEnv("CLICKHOUSE_USER", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Keycloak.URL == "" {
		return fmt.Errorf("KEYCLOAK_URL is required")
	}
	if c.Keycloak.Realm == "" {
		return fmt.Errorf("KEYCLOAK_REALM is required")
	}
	if c.ClickHouse.Host == "" {
		return fmt.Errorf("CLICKHOUSE_HOST is required")
	}
	return nil
}

// GetAddress возвращает адрес сервера в формате host:port
func (s *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
