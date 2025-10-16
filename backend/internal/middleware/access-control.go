package middleware

import (
	"log"
	"net/http"
)

// AccessLogger логирует попытки доступа к защищённым ресурсам
type AccessLogger struct{}

// NewAccessLogger создаёт новый access logger
func NewAccessLogger() *AccessLogger {
	return &AccessLogger{}
}

// LogAccess логирует попытку доступа к ресурсу
func (al *AccessLogger) LogAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем информацию о пользователе
		userInfo, err := GetUserFromContext(r.Context())
		if err != nil {
			log.Printf("[ACCESS] Anonymous request: %s %s from %s",
				r.Method, r.URL.Path, r.RemoteAddr)
		} else {
			log.Printf("[ACCESS] User: %s (%s) - %s %s from %s",
				userInfo.Username, userInfo.UserID, r.Method, r.URL.Path, r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}

// PreventUserIDManipulation проверяет, что user_id не передаётся в query параметрах
func (al *AccessLogger) PreventUserIDManipulation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем query параметры на наличие попыток манипуляции user_id
		queryParams := r.URL.Query()

		suspiciousParams := []string{"user_id", "userId", "uid", "owner", "ownerId"}
		for _, param := range suspiciousParams {
			if queryParams.Has(param) {
				userInfo, _ := GetUserFromContext(r.Context())
				userID := "anonymous"
				if userInfo != nil {
					userID = userInfo.UserID
				}

				log.Printf("[SECURITY ALERT] User %s attempted to manipulate %s parameter: %s=%s from %s",
					userID, param, param, queryParams.Get(param), r.RemoteAddr)

				http.Error(w, "forbidden: parameter manipulation detected", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter простой rate limiter (в продакшн лучше использовать Redis)
type RateLimiter struct {
	// В реальном приложении здесь будет карта с лимитами
}

// NewRateLimiter создаёт новый rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

// Limit ограничивает количество запросов (упрощённая версия)
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// В продакшн версии здесь будет проверка rate limits
		// Сейчас просто пропускаем запрос
		next.ServeHTTP(w, r)
	})
}
