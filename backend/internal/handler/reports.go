package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/spinel/architecture-bionicpro/backend/internal/middleware"
	"github.com/spinel/architecture-bionicpro/backend/internal/models"
	"github.com/spinel/architecture-bionicpro/backend/internal/storage"
)

// ReportsHandler обрабатывает запросы к API отчётов
type ReportsHandler struct {
	storage *storage.ClickHouseStorage
}

// NewReportsHandler создаёт новый handler для отчётов
func NewReportsHandler(storage *storage.ClickHouseStorage) *ReportsHandler {
	return &ReportsHandler{
		storage: storage,
	}
}

// GetReports возвращает список отчётов для текущего пользователя
// GET /api/reports?limit=10&offset=0
//
// БЕЗОПАСНОСТЬ:
// - Доступ только к своим отчётам (user_id из JWT)
// - Нельзя передать user_id в query параметрах
// - Фильтрация на уровне SQL запроса
func (h *ReportsHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о пользователе из контекста (извлечена из JWT)
	userInfo, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// БЕЗОПАСНОСТЬ: Используем ТОЛЬКО user_id из JWT токена
	// Игнорируем любые попытки передать user_id через параметры
	authenticatedUserID := userInfo.UserID

	// Парсим параметры пагинации
	limit := h.getIntParam(r, "limit", 10)
	offset := h.getIntParam(r, "offset", 0)

	// Валидация параметров
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем отчёты из OLAP базы
	// ВАЖНО: Используем authenticatedUserID из JWT, а не из query params
	reports, err := h.storage.GetReportsByUser(r.Context(), authenticatedUserID, limit, offset)
	if err != nil {
		h.sendError(w, "failed to fetch reports", http.StatusInternalServerError)
		return
	}

	// Получаем общее количество отчётов
	totalCount, err := h.storage.CountReportsByUser(r.Context(), authenticatedUserID)
	if err != nil {
		h.sendError(w, "failed to count reports", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := models.ReportResponse{
		Reports:    reports,
		TotalCount: totalCount,
		Page:       offset/limit + 1,
		PageSize:   limit,
	}

	// Если отчётов нет, возвращаем пустой массив
	if response.Reports == nil {
		response.Reports = []models.Report{}
	}

	h.sendJSON(w, response, http.StatusOK)
}

// GetReportByID возвращает конкретный отчёт по ID
// GET /api/reports/{id}
//
// БЕЗОПАСНОСТЬ:
// - Проверка принадлежности отчёта пользователю
// - Возвращает 403 Forbidden если отчёт принадлежит другому пользователю
// - Возвращает 404 Not Found если отчёт не существует
func (h *ReportsHandler) GetReportByID(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о пользователе из контекста (извлечена из JWT)
	userInfo, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		h.sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// БЕЗОПАСНОСТЬ: Используем ТОЛЬКО user_id из JWT токена
	authenticatedUserID := userInfo.UserID

	// Извлекаем ID отчёта из URL
	vars := mux.Vars(r)
	reportID := vars["id"]

	if reportID == "" {
		h.sendError(w, "report ID is required", http.StatusBadRequest)
		return
	}

	// Получаем отчёт из OLAP базы
	report, err := h.storage.GetReportByID(r.Context(), reportID)
	if err != nil {
		// Не раскрываем, существует ли отчёт - возвращаем 404
		h.sendError(w, "report not found", http.StatusNotFound)
		return
	}

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Отчёт принадлежит текущему пользователю?
	if report.UserID != authenticatedUserID {
		// Логируем попытку доступа к чужому отчёту
		h.logSecurityEvent(
			"ACCESS_DENIED",
			authenticatedUserID,
			userInfo.Username,
			reportID,
			report.UserID,
			r.RemoteAddr,
		)

		// Возвращаем 403 Forbidden
		h.sendError(w, "access denied: you can only access your own reports", http.StatusForbidden)
		return
	}

	h.sendJSON(w, report, http.StatusOK)
}

// HealthCheck проверяет здоровье сервиса
// GET /api/health
func (h *ReportsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, map[string]string{
		"status":  "healthy",
		"service": "reports-api",
	}, http.StatusOK)
}

// getIntParam извлекает целочисленный параметр из query string
func (h *ReportsHandler) getIntParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// sendJSON отправляет JSON ответ
func (h *ReportsHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// sendError отправляет JSON ошибку
func (h *ReportsHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// logSecurityEvent логирует события безопасности
func (h *ReportsHandler) logSecurityEvent(
	eventType string,
	userID string,
	username string,
	reportID string,
	reportOwnerID string,
	remoteAddr string,
) {
	log.Printf("[SECURITY] %s: User %s (%s) attempted to access report %s (owner: %s) from %s",
		eventType, username, userID, reportID, reportOwnerID, remoteAddr)
}
