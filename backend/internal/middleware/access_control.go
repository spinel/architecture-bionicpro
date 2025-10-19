package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AccessControlMiddleware middleware для контроля доступа к отчётам
type AccessControlMiddleware struct {
	logger *logrus.Logger
}

// NewAccessControlMiddleware создаёт новый middleware контроля доступа
func NewAccessControlMiddleware(logger *logrus.Logger) *AccessControlMiddleware {
	return &AccessControlMiddleware{
		logger: logger,
	}
}

// RequireOwnData возвращает middleware, который проверяет доступ только к собственным данным
func (m *AccessControlMiddleware) RequireOwnData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из контекста (устанавливается в RequireAuth)
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Error("ID пользователя не найден в контексте")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Неверный тип ID пользователя")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		// Проверяем, что пользователь запрашивает данные только о себе
		requestedUserID := c.Param("user_id")
		if requestedUserID == "" {
			// Если user_id не указан в параметрах, используем ID из токена
			c.Set("requested_user_id", userIDStr)
		} else {
			// Проверяем, что запрашиваемый пользователь совпадает с аутентифицированным
			if requestedUserID != userIDStr {
				m.logger.WithFields(logrus.Fields{
					"authenticated_user": userIDStr,
					"requested_user":     requestedUserID,
				}).Warn("Попытка доступа к данным другого пользователя")

				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Доступ запрещён: вы можете просматривать только свои отчёты",
				})
				c.Abort()
				return
			}
			c.Set("requested_user_id", requestedUserID)
		}

		c.Next()
	}
}

// RequireOwnDataFromQuery проверяет доступ к данным из query параметров
func (m *AccessControlMiddleware) RequireOwnDataFromQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из контекста
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Error("ID пользователя не найден в контексте")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Неверный тип ID пользователя")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		// Получаем user_id из query параметров
		requestedUserID := c.Query("user_id")
		if requestedUserID == "" {
			// Если user_id не указан, используем ID из токена
			c.Set("requested_user_id", userIDStr)
		} else {
			// Проверяем, что запрашиваемый пользователь совпадает с аутентифицированным
			if requestedUserID != userIDStr {
				m.logger.WithFields(logrus.Fields{
					"authenticated_user": userIDStr,
					"requested_user":     requestedUserID,
				}).Warn("Попытка доступа к данным другого пользователя через query параметр")

				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Доступ запрещён: вы можете просматривать только свои отчёты",
				})
				c.Abort()
				return
			}
			c.Set("requested_user_id", requestedUserID)
		}

		c.Next()
	}
}

// RequireAdminOrOwnData возвращает middleware, который разрешает доступ администраторам или к собственным данным
func (m *AccessControlMiddleware) RequireAdminOrOwnData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем информацию о пользователе
		userInfo, exists := c.Get("user")
		if !exists {
			m.logger.Error("Информация о пользователе не найдена")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Error("ID пользователя не найден в контексте")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Неверный тип ID пользователя")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Ошибка аутентификации",
			})
			c.Abort()
			return
		}

		// Проверяем, является ли пользователь администратором
		// (здесь можно добавить проверку ролей из userInfo)
		isAdmin := m.isAdmin(userInfo)

		if isAdmin {
			// Администратор может просматривать любые данные
			requestedUserID := c.Param("user_id")
			if requestedUserID == "" {
				requestedUserID = c.Query("user_id")
			}
			if requestedUserID == "" {
				requestedUserID = userIDStr
			}
			c.Set("requested_user_id", requestedUserID)
			c.Next()
			return
		}

		// Обычный пользователь может просматривать только свои данные
		requestedUserID := c.Param("user_id")
		if requestedUserID == "" {
			requestedUserID = c.Query("user_id")
		}
		if requestedUserID == "" {
			requestedUserID = userIDStr
		}

		if requestedUserID != userIDStr {
			m.logger.WithFields(logrus.Fields{
				"authenticated_user": userIDStr,
				"requested_user":     requestedUserID,
			}).Warn("Попытка доступа к данным другого пользователя")

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Доступ запрещён: вы можете просматривать только свои отчёты",
			})
			c.Abort()
			return
		}

		c.Set("requested_user_id", requestedUserID)
		c.Next()
	}
}

// isAdmin проверяет, является ли пользователь администратором
func (m *AccessControlMiddleware) isAdmin(userInfo interface{}) bool {
	// Здесь можно добавить логику проверки ролей
	// Пока возвращаем false для всех пользователей
	// В реальном приложении здесь была бы проверка ролей из JWT токена
	return false
}

// GetRequestedUserID извлекает ID запрашиваемого пользователя из контекста
func GetRequestedUserID(c *gin.Context) (string, bool) {
	requestedUserID, exists := c.Get("requested_user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := requestedUserID.(string)
	return userIDStr, ok
}
