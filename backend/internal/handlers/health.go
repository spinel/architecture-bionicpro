package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"bionicpro-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// HealthHandler обработчик для проверки состояния системы
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler создаёт новый экземпляр HealthHandler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// Health проверяет состояние системы
// @Summary Проверить состояние системы
// @Description Возвращает информацию о состоянии API и подключении к базе данных
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// Проверяем подключение к базе данных
	dbStatus := "connected"
	if err := h.db.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Database:  dbStatus,
	}

	// Если база данных недоступна, возвращаем ошибку
	if dbStatus == "disconnected" {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Ready проверяет готовность системы
// @Summary Проверить готовность системы
// @Description Проверяет, готова ли система к обработке запросов
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// Проверяем подключение к базе данных
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false,
			Error:   "система не готова: база данных недоступна",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}

	// Проверяем наличие необходимых таблиц
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_name = 'user_analytics_warehouse'
	)`

	var tableExists bool
	if err := h.db.QueryRow(query).Scan(&tableExists); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false,
			Error:   "система не готова: ошибка проверки таблиц",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}

	if !tableExists {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false,
			Error:   "система не готова: отсутствуют необходимые таблицы",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}

	response := models.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Database:  "connected",
	}

	c.JSON(http.StatusOK, response)
}
