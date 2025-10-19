package handlers

import (
	"bionicpro-backend/internal/auth"
	"bionicpro-backend/internal/middleware"
	"bionicpro-backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ReportHandler обработчик для работы с отчетами
type ReportHandler struct {
	reportService *service.ReportService
	authService   auth.AuthService
	logger        *logrus.Logger
}

// NewReportHandler создает новый обработчик отчетов
func NewReportHandler(reportService *service.ReportService, authService auth.AuthService, logger *logrus.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		authService:   authService,
		logger:        logger,
	}
}

// GetReports возвращает список отчетов пользователя
func (h *ReportHandler) GetReports(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID пользователя не найден",
		})
		return
	}

	reports, err := h.reportService.GetReports(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Ошибка получения отчетов")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка получения отчетов",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
	})
}

// GetReport возвращает конкретный отчет
func (h *ReportHandler) GetReport(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID пользователя не найден",
		})
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID отчета",
		})
		return
	}

	report, err := h.reportService.GetReport(c.Request.Context(), reportID, userID)
	if err != nil {
		h.logger.WithError(err).Error("Ошибка получения отчета")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Отчет не найден",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// GenerateReportRequest запрос на генерацию отчета
type GenerateReportRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

// GenerateReport создает новый отчет
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID пользователя не найден",
		})
		return
	}

	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные запроса",
		})
		return
	}

	// Парсим даты
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат даты начала",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный формат даты окончания",
		})
		return
	}

	// Создаем данные отчета
	reportData := service.ReportData{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// Генерируем отчет
	report, err := h.reportService.GenerateReport(c.Request.Context(), reportData)
	if err != nil {
		h.logger.WithError(err).Error("Ошибка генерации отчета")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка генерации отчета",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"report":  report,
		"message": "Отчет поставлен в очередь на генерацию",
	})
}

// DownloadReport скачивает отчет
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID пользователя не найден",
		})
		return
	}

	reportIDStr := c.Param("id")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID отчета",
		})
		return
	}

	// Получаем данные отчета
	reportData, err := h.reportService.DownloadReport(c.Request.Context(), reportID, userID)
	if err != nil {
		h.logger.WithError(err).Error("Ошибка скачивания отчета")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ошибка скачивания отчета",
		})
		return
	}

	// Устанавливаем заголовки для скачивания файла
	c.Header("Content-Disposition", "attachment; filename=prothetic-report.txt")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain", reportData)
}
