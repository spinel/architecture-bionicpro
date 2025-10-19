package auth

import (
	"bionicpro-backend/internal/config"
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// DemoService предоставляет упрощённый сервис аутентификации для демонстрации
type DemoService struct {
	config *config.AuthConfig
	logger *logrus.Logger
}

// NewDemoService создает новый демо-сервис аутентификации
func NewDemoService(cfg *config.AuthConfig) *DemoService {
	return &DemoService{
		config: cfg,
	}
}

// ValidateToken проверяет валидность токена (упрощённая версия для демонстрации)
func (s *DemoService) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	// Убираем префикс "Bearer " если он есть
	token = strings.TrimPrefix(token, "Bearer ")

	// Парсим токен
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("ошибка валидации токена: %w", err)
	}

	// Создаём UserInfo из claims
	userInfo := &UserInfo{
		Sub:               claims.Sub,
		Email:             claims.Email,
		PreferredUsername: claims.Sub,
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{
			Roles: []string{"prothetic_user"},
		},
	}

	return userInfo, nil
}

// parseToken парсит JWT токен
func (s *DemoService) parseToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный алгоритм подписи: %v", token.Header["alg"])
		}
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
func (s *DemoService) HasRole(userInfo *UserInfo, role string) bool {
	for _, userRole := range userInfo.RealmAccess.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// IsAdmin проверяет, является ли пользователь администратором
func (s *DemoService) IsAdmin(userInfo *UserInfo) bool {
	return s.HasRole(userInfo, "administrator")
}

// IsProtheticUser проверяет, является ли пользователь пользователем протеза
func (s *DemoService) IsProtheticUser(userInfo *UserInfo) bool {
	return s.HasRole(userInfo, "prothetic_user")
}
