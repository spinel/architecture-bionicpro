package handlers

import (
	"net/http"
	"strconv"
	"time"

	"bionicpro-backend/internal/models"
	"bionicpro-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler обработчик для аналитических данных
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

// NewAnalyticsHandler создаёт новый экземпляр AnalyticsHandler
func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetUserReport получает отчёт пользователя
// @Summary Получить отчёт пользователя
// @Description Возвращает аналитический отчёт пользователя за указанный период
// @Tags reports
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param user_id query string false "ID пользователя (опционально, по умолчанию - текущий пользователь)"
// @Param start_date query string true "Начальная дата (YYYY-MM-DD)"
// @Param end_date query string true "Конечная дата (YYYY-MM-DD)"
// @Success 200 {object} models.ReportResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reports [get]
func (h *AnalyticsHandler) GetUserReport(c *gin.Context) {
	// Получаем ID пользователя из контекста (устанавливается middleware)
	userID, exists := c.Get("requested_user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения ID пользователя",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Неверный тип ID пользователя",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "параметр start_date обязателен",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "параметр end_date обязателен",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Парсим даты
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "неверный формат start_date, используйте YYYY-MM-DD",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "неверный формат end_date, используйте YYYY-MM-DD",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Получаем отчёт
	report, err := h.analyticsService.GetUserReport(userIDStr, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "внутренняя ошибка сервера",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Возвращаем результат
	if report.Success {
		c.JSON(http.StatusOK, report)
	} else {
		c.JSON(http.StatusNotFound, report)
	}
}

// GetDailySummary получает сводку по дням
// @Summary Получить сводку по дням
// @Description Возвращает сводку аналитических данных по дням за указанный период
// @Tags reports
// @Accept json
// @Produce json
// @Param start_date query string true "Начальная дата (YYYY-MM-DD)"
// @Param end_date query string true "Конечная дата (YYYY-MM-DD)"
// @Success 200 {array} models.DailySummary
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reports/summary [get]
func (h *AnalyticsHandler) GetDailySummary(c *gin.Context) {
	// Получаем параметры запроса
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "параметр start_date обязателен",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "параметр end_date обязателен",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Парсим даты
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "неверный формат start_date, используйте YYYY-MM-DD",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "неверный формат end_date, используйте YYYY-MM-DD",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Получаем сводку
	summaries, err := h.analyticsService.GetDailySummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "внутренняя ошибка сервера",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, summaries)
}

// GetUserReportByID получает отчёт пользователя по ID (для совместимости)
// @Summary Получить отчёт пользователя по ID
// @Description Возвращает отчёт пользователя по ID с дополнительными параметрами
// @Tags reports
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Param user_id path string true "ID пользователя"
// @Param days query int false "Количество дней назад (по умолчанию 7)"
// @Success 200 {object} models.ReportResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reports/user/{user_id} [get]
func (h *AnalyticsHandler) GetUserReportByID(c *gin.Context) {
	// Получаем ID пользователя из контекста (устанавливается middleware)
	userID, exists := c.Get("requested_user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения ID пользователя",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Неверный тип ID пользователя",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Получаем количество дней (по умолчанию 7)
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	// Вычисляем даты
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Получаем отчёт
	report, err := h.analyticsService.GetUserReport(userIDStr, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "внутренняя ошибка сервера",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Возвращаем результат
	if report.Success {
		c.JSON(http.StatusOK, report)
	} else {
		c.JSON(http.StatusNotFound, report)
	}
}
