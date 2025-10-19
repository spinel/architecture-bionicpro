package middleware

import (
	"bionicpro-backend/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware middleware для аутентификации
type AuthMiddleware struct {
	authService auth.AuthService
	logger      *logrus.Logger
}

// NewAuthMiddleware создает новый middleware аутентификации
func NewAuthMiddleware(authService auth.AuthService, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth возвращает middleware, который требует аутентификации
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Токен авторизации не предоставлен",
			})
			c.Abort()
			return
		}

		// Проверяем формат токена
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный формат токена авторизации",
			})
			c.Abort()
			return
		}

		// Извлекаем токен
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Валидируем токен через Keycloak
		userInfo, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.WithError(err).Error("Ошибка валидации токена")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный или истекший токен",
			})
			c.Abort()
			return
		}

		// Сохраняем информацию о пользователе в контексте
		c.Set("user", userInfo)
		c.Set("user_id", userInfo.Sub)

		// Продолжаем выполнение
		c.Next()
	}
}

// RequireRole возвращает middleware, который требует определенную роль
func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Сначала проверяем аутентификацию
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Получаем информацию о пользователе
		userInfo, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Информация о пользователе не найдена",
			})
			c.Abort()
			return
		}

		// Проверяем роль
		if !m.authService.HasRole(userInfo.(*auth.UserInfo), role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Недостаточно прав доступа",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin возвращает middleware, который требует права администратора
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("administrator")
}

// RequireProtheticUser возвращает middleware, который требует права пользователя протеза
func (m *AuthMiddleware) RequireProtheticUser() gin.HandlerFunc {
	return m.RequireRole("prothetic_user")
}

// GetUserFromContext извлекает информацию о пользователе из контекста
func GetUserFromContext(c *gin.Context) (*auth.UserInfo, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userInfo, ok := user.(*auth.UserInfo)
	return userInfo, ok
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
