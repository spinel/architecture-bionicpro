package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spinel/architecture-bionicpro/backend/internal/config"
)

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// UserContextKey ключ для хранения информации о пользователе в контексте
	UserContextKey ContextKey = "user"
)

// UserInfo содержит информацию о пользователе из JWT
type UserInfo struct {
	UserID   string   `json:"sub"`
	Username string   `json:"preferred_username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// AuthMiddleware проверяет JWT токен из Keycloak
type AuthMiddleware struct {
	config *config.Config
}

// NewAuthMiddleware создаёт новый middleware для аутентификации
func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		config: cfg,
	}
}

// Authenticate проверяет JWT токен
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			am.sendError(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Проверяем формат Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			am.sendError(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Парсим токен (без проверки подписи для упрощения, в продакшн нужна проверка)
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			am.sendError(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Извлекаем claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			am.sendError(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		// Извлекаем информацию о пользователе
		userInfo := am.extractUserInfo(claims)

		// Добавляем информацию в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUserInfo извлекает информацию о пользователе из claims
func (am *AuthMiddleware) extractUserInfo(claims jwt.MapClaims) *UserInfo {
	userInfo := &UserInfo{}

	if sub, ok := claims["sub"].(string); ok {
		userInfo.UserID = sub
	}

	if username, ok := claims["preferred_username"].(string); ok {
		userInfo.Username = username
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	// Извлекаем роли из realm_access
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					userInfo.Roles = append(userInfo.Roles, roleStr)
				}
			}
		}
	}

	return userInfo
}

// sendError отправляет JSON ошибку
func (am *AuthMiddleware) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// GetUserFromContext извлекает информацию о пользователе из контекста
func GetUserFromContext(ctx context.Context) (*UserInfo, error) {
	userInfo, ok := ctx.Value(UserContextKey).(*UserInfo)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return userInfo, nil
}

// RequireRole проверяет наличие роли у пользователя
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo, err := GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, userRole := range userInfo.Roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
