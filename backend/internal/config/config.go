package config

import (
	"os"
	"strconv"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
	Server   ServerConfig
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AuthConfig конфигурация аутентификации
type AuthConfig struct {
	KeycloakURL      string
	KeycloakRealm    string
	KeycloakClientID string
	JWTSecret        string
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port string
	Host string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		port = 5432
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     port,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "bionicpro"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			KeycloakURL:      getEnv("KEYCLOAK_URL", "http://localhost:8080"),
			KeycloakRealm:    getEnv("KEYCLOAK_REALM", "reports-realm"),
			KeycloakClientID: getEnv("KEYCLOAK_CLIENT_ID", "reports-api"),
			JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
	}, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
