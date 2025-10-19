package service

import (
	"fmt"
	"time"

	"bionicpro-backend/internal/models"
	"bionicpro-backend/internal/repository"
)

// AnalyticsService интерфейс для работы с аналитическими данными
type AnalyticsService interface {
	GetUserReport(userID string, startDate, endDate time.Time) (*models.ReportResponse, error)
	GetDailySummary(startDate, endDate time.Time) ([]models.DailySummary, error)
	ValidateDateRange(startDate, endDate time.Time) error
}

// analyticsService реализация AnalyticsService
type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

// NewAnalyticsService создаёт новый экземпляр AnalyticsService
func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// GetUserReport получает отчёт пользователя
func (s *analyticsService) GetUserReport(userID string, startDate, endDate time.Time) (*models.ReportResponse, error) {
	// Валидация входных данных
	if err := s.ValidateDateRange(startDate, endDate); err != nil {
		return &models.ReportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Проверяем существование пользователя
	exists, err := s.analyticsRepo.CheckUserExists(userID)
	if err != nil {
		return &models.ReportResponse{
			Success: false,
			Error:   fmt.Sprintf("ошибка проверки пользователя: %v", err),
		}, nil
	}

	if !exists {
		return &models.ReportResponse{
			Success: false,
			Error:   fmt.Sprintf("пользователь %s не найден", userID),
		}, nil
	}

	// Получаем отчёты пользователя
	reports, err := s.analyticsRepo.GetUserReports(userID, startDate, endDate)
	if err != nil {
		return &models.ReportResponse{
			Success: false,
			Error:   fmt.Sprintf("ошибка получения отчётов: %v", err),
		}, nil
	}

	// Получаем сводку по пользователю
	summary, err := s.analyticsRepo.GetUserReportSummary(userID, startDate, endDate)
	if err != nil {
		// Если сводка не найдена, продолжаем без неё
		summary = nil
	}

	// Формируем метаинформацию
	meta := models.ReportMeta{
		GeneratedAt: time.Now(),
		UserID:      userID,
		DateRange:   fmt.Sprintf("%s - %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		RecordCount: len(reports),
	}

	// Формируем ответ
	response := &models.ReportResponse{
		Success: true,
		Data:    reports,
		Meta:    meta,
	}

	if summary != nil {
		response.Summary = *summary
	}

	return response, nil
}

// GetDailySummary получает сводку по дням
func (s *analyticsService) GetDailySummary(startDate, endDate time.Time) ([]models.DailySummary, error) {
	// Валидация входных данных
	if err := s.ValidateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Получаем сводку по дням
	summaries, err := s.analyticsRepo.GetDailySummary(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сводки по дням: %w", err)
	}

	return summaries, nil
}

// ValidateDateRange валидирует диапазон дат
func (s *analyticsService) ValidateDateRange(startDate, endDate time.Time) error {
	// Проверяем, что начальная дата не позже конечной
	if startDate.After(endDate) {
		return fmt.Errorf("начальная дата не может быть позже конечной")
	}

	// Проверяем, что даты не в будущем
	now := time.Now()
	if startDate.After(now) {
		return fmt.Errorf("начальная дата не может быть в будущем")
	}

	// Проверяем максимальный диапазон (например, 1 год)
	maxRange := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxRange {
		return fmt.Errorf("максимальный диапазон дат: 1 год")
	}

	return nil
}
