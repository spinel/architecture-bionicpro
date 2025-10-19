package auth

import (
	"bionicpro-backend/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// Service предоставляет сервис аутентификации
type Service struct {
	config *config.AuthConfig
	logger *logrus.Logger
}

// UserInfo содержит информацию о пользователе из Keycloak
type UserInfo struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// TokenClaims содержит claims из JWT токена
type TokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// NewService создает новый сервис аутентификации
func NewService(cfg *config.AuthConfig) *Service {
	return &Service{
		config: cfg,
	}
}

// ValidateToken проверяет валидность токена через Keycloak
func (s *Service) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	// Убираем префикс "Bearer " если он есть
	token = strings.TrimPrefix(token, "Bearer ")

	// Проверяем токен через Keycloak
	userInfo, err := s.getUserInfoFromKeycloak(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("ошибка валидации токена: %w", err)
	}

	return userInfo, nil
}

// getUserInfoFromKeycloak получает информацию о пользователе из Keycloak
func (s *Service) getUserInfoFromKeycloak(ctx context.Context, token string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", s.config.KeycloakURL, s.config.KeycloakRealm)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &userInfo, nil
}

// ParseToken парсит JWT токен (для внутреннего использования)
func (s *Service) ParseToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// В реальном приложении здесь должна быть проверка подписи
		// Для демонстрации возвращаем секретный ключ
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("неверный токен")
}

// HasRole проверяет, есть ли у пользователя указанная роль
func (s *Service) HasRole(userInfo *UserInfo, role string) bool {
	for _, userRole := range userInfo.RealmAccess.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// IsAdmin проверяет, является ли пользователь администратором
func (s *Service) IsAdmin(userInfo *UserInfo) bool {
	return s.HasRole(userInfo, "administrator")
}

// IsProtheticUser проверяет, является ли пользователь пользователем протеза
func (s *Service) IsProtheticUser(userInfo *UserInfo) bool {
	return s.HasRole(userInfo, "prothetic_user")
}
