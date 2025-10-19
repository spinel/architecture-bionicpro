package service

import (
	"bionicpro-backend/internal/models"
	"bionicpro-backend/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ReportService предоставляет сервис для работы с отчетами
type ReportService struct {
	reportRepo repository.ReportRepository
	logger     *logrus.Logger
}

// ReportData содержит данные для генерации отчета
type ReportData struct {
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// NewReportService создает новый сервис отчетов
func NewReportService(reportRepo repository.ReportRepository, logger *logrus.Logger) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		logger:     logger,
	}
}

// GetReports возвращает список отчетов для пользователя
func (s *ReportService) GetReports(ctx context.Context, userID string) ([]models.Report, error) {
	s.logger.WithField("user_id", userID).Info("Получение списка отчетов")

	reports, err := s.reportRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения отчетов")
		return nil, fmt.Errorf("ошибка получения отчетов: %w", err)
	}

	return reports, nil
}

// GetReport возвращает конкретный отчет
func (s *ReportService) GetReport(ctx context.Context, reportID int, userID string) (*models.Report, error) {
	s.logger.WithFields(logrus.Fields{
		"report_id": reportID,
		"user_id":   userID,
	}).Info("Получение отчета")

	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка получения отчета")
		return nil, fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		s.logger.WithFields(logrus.Fields{
			"report_id": reportID,
			"user_id":   userID,
		}).Warn("Попытка доступа к чужому отчету")
		return nil, fmt.Errorf("отчет не найден")
	}

	return report, nil
}

// GenerateReport создает новый отчет
func (s *ReportService) GenerateReport(ctx context.Context, data ReportData) (*models.Report, error) {
	s.logger.WithField("user_id", data.UserID).Info("Генерация отчета")

	// Создаем запись отчета в базе данных
	report := &models.Report{
		UserID:      data.UserID,
		Title:       data.Title,
		Description: data.Description,
		CreatedAt:   time.Now(),
		Status:      "generating",
	}

	// Сохраняем отчет в базе данных
	reportID, err := s.reportRepo.Create(ctx, report)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка создания отчета")
		return nil, fmt.Errorf("ошибка создания отчета: %w", err)
	}

	report.ID = reportID

	// В реальном приложении здесь была бы асинхронная генерация отчета
	// Для демонстрации просто обновляем статус
	go s.generateReportAsync(ctx, reportID, data)

	return report, nil
}

// generateReportAsync асинхронно генерирует отчет
func (s *ReportService) generateReportAsync(ctx context.Context, reportID int, data ReportData) {
	// Имитируем время генерации отчета
	time.Sleep(5 * time.Second)

	// Обновляем статус отчета
	err := s.reportRepo.UpdateStatus(ctx, reportID, "completed")
	if err != nil {
		s.logger.WithError(err).Error("Ошибка обновления статуса отчета")
		return
	}

	// Генерируем URL файла отчета
	fileURL := fmt.Sprintf("/api/v1/reports/download/%d", reportID)
	err = s.reportRepo.UpdateFileURL(ctx, reportID, fileURL)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка обновления URL файла")
		return
	}

	s.logger.WithField("report_id", reportID).Info("Отчет успешно сгенерирован")
}

// DownloadReport возвращает данные для скачивания отчета
func (s *ReportService) DownloadReport(ctx context.Context, reportID int, userID string) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"report_id": reportID,
		"user_id":   userID,
	}).Info("Скачивание отчета")

	// Получаем отчет
	report, err := s.GetReport(ctx, reportID, userID)
	if err != nil {
		return nil, err
	}

	// Проверяем статус отчета
	if report.Status != "completed" {
		return nil, fmt.Errorf("отчет еще не готов")
	}

	// В реальном приложении здесь была бы генерация PDF файла
	// Для демонстрации возвращаем простой текстовый отчет
	reportContent := s.generateReportContent(report, ReportData{
		UserID:      userID,
		Title:       report.Title,
		Description: report.Description,
		StartDate:   time.Now().AddDate(0, 0, -30), // 30 дней назад
		EndDate:     time.Now(),
	})

	return []byte(reportContent), nil
}

// generateReportContent генерирует содержимое отчета
func (s *ReportService) generateReportContent(report *models.Report, data ReportData) string {
	content := fmt.Sprintf(`
ОТЧЕТ О РАБОТЕ ПРОТЕЗА
======================

ID отчета: %d
Пользователь: %s
Дата создания: %s
Период: %s - %s

ОПИСАНИЕ
--------
%s

СТАТИСТИКА ИСПОЛЬЗОВАНИЯ
------------------------
- Общее время использования: 8 часов 30 минут
- Количество движений: 1,247
- Средняя точность: 94.2%%
- Количество сессий: 12

ТЕХНИЧЕСКОЕ СОСТОЯНИЕ
--------------------
- Уровень заряда батареи: 87%%
- Температура: 23°C
- Статус датчиков: Норма
- Последняя калибровка: %s

РЕКОМЕНДАЦИИ
------------
1. Рекомендуется провести калибровку датчиков
2. Уровень заряда батареи в норме
3. Производительность в пределах нормы

---
Сгенерировано системой BionicPRO
Дата: %s
`,
		report.ID,
		report.UserID,
		report.CreatedAt.Format("2006-01-02 15:04:05"),
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		report.Description,
		time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return content
}
