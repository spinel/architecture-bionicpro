package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DemoAccessControlMiddleware упрощённый middleware для демонстрации контроля доступа
type DemoAccessControlMiddleware struct {
	logger *logrus.Logger
}

// NewDemoAccessControlMiddleware создаёт новый демо-middleware контроля доступа
func NewDemoAccessControlMiddleware(logger *logrus.Logger) *DemoAccessControlMiddleware {
	return &DemoAccessControlMiddleware{
		logger: logger,
	}
}

// RequireOwnDataDemo демонстрирует контроль доступа без аутентификации
func (m *DemoAccessControlMiddleware) RequireOwnDataDemo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из заголовка X-User-ID (для демонстрации)
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Заголовок X-User-ID обязателен для демонстрации",
			})
			c.Abort()
			return
		}

		// Получаем запрашиваемый user_id
		requestedUserID := c.Param("user_id")
		if requestedUserID == "" {
			requestedUserID = c.Query("user_id")
		}
		if requestedUserID == "" {
			requestedUserID = userID
		}

		// Проверяем, что пользователь запрашивает только свои данные
		if requestedUserID != userID {
			m.logger.WithFields(logrus.Fields{
				"authenticated_user": userID,
				"requested_user":     requestedUserID,
			}).Warn("Попытка доступа к данным другого пользователя")

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Доступ запрещён: вы можете просматривать только свои отчёты",
			})
			c.Abort()
			return
		}

		// Сохраняем ID пользователя в контексте
		c.Set("requested_user_id", requestedUserID)
		c.Next()
	}
}

// RequireOwnDataFromQueryDemo проверяет доступ к данным из query параметров (демо)
func (m *DemoAccessControlMiddleware) RequireOwnDataFromQueryDemo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из заголовка X-User-ID (для демонстрации)
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Заголовок X-User-ID обязателен для демонстрации",
			})
			c.Abort()
			return
		}

		// Получаем запрашиваемый user_id из query
		requestedUserID := c.Query("user_id")
		if requestedUserID == "" {
			requestedUserID = userID
		}

		// Проверяем, что пользователь запрашивает только свои данные
		if requestedUserID != userID {
			m.logger.WithFields(logrus.Fields{
				"authenticated_user": userID,
				"requested_user":     requestedUserID,
			}).Warn("Попытка доступа к данным другого пользователя через query параметр")

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Доступ запрещён: вы можете просматривать только свои отчёты",
			})
			c.Abort()
			return
		}

		// Сохраняем ID пользователя в контексте
		c.Set("requested_user_id", requestedUserID)
		c.Next()
	}
}
