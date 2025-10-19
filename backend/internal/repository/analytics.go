package repository

import (
	"database/sql"
	"fmt"
	"time"

	"bionicpro-backend/internal/models"
)

// AnalyticsRepository интерфейс для работы с аналитическими данными
type AnalyticsRepository interface {
	GetUserReports(userID string, startDate, endDate time.Time) ([]models.UserReport, error)
	GetDailySummary(startDate, endDate time.Time) ([]models.DailySummary, error)
	GetUserReportSummary(userID string, startDate, endDate time.Time) (*models.DailySummary, error)
	CheckUserExists(userID string) (bool, error)
}

// analyticsRepository реализация AnalyticsRepository
type analyticsRepository struct {
	db *sql.DB
}

// NewAnalyticsRepository создаёт новый экземпляр AnalyticsRepository
func NewAnalyticsRepository(db *sql.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetUserReports получает отчёты пользователя за период
func (r *analyticsRepository) GetUserReports(userID string, startDate, endDate time.Time) ([]models.UserReport, error) {
	query := `
		SELECT 
			user_id,
			date_created,
			total_sessions,
			avg_accuracy,
			avg_battery_level,
			avg_temperature,
			first_activity,
			last_activity,
			total_movements,
			CASE 
				WHEN avg_accuracy >= 95 THEN 'Отличная'
				WHEN avg_accuracy >= 90 THEN 'Хорошая'
				WHEN avg_accuracy >= 80 THEN 'Удовлетворительная'
				ELSE 'Требует внимания'
			END as accuracy_status,
			CASE 
				WHEN avg_battery_level >= 80 THEN 'Высокий'
				WHEN avg_battery_level >= 50 THEN 'Средний'
				WHEN avg_battery_level >= 20 THEN 'Низкий'
				ELSE 'Критический'
			END as battery_status,
			EXTRACT(EPOCH FROM (last_activity - first_activity))/3600 as session_duration_hours
		FROM user_analytics_warehouse 
		WHERE user_id = $1 
		AND date_created BETWEEN $2 AND $3
		ORDER BY date_created DESC
	`

	rows, err := r.db.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var reports []models.UserReport
	for rows.Next() {
		var report models.UserReport
		err := rows.Scan(
			&report.UserID,
			&report.DateCreated,
			&report.TotalSessions,
			&report.AvgAccuracy,
			&report.AvgBatteryLevel,
			&report.AvgTemperature,
			&report.FirstActivity,
			&report.LastActivity,
			&report.TotalMovements,
			&report.AccuracyStatus,
			&report.BatteryStatus,
			&report.SessionDuration,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		reports = append(reports, report)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return reports, nil
}

// GetDailySummary получает сводку по дням
func (r *analyticsRepository) GetDailySummary(startDate, endDate time.Time) ([]models.DailySummary, error) {
	query := `
		SELECT 
			date_created,
			COUNT(DISTINCT user_id) as active_users,
			AVG(total_sessions) as avg_sessions_per_user,
			AVG(avg_accuracy) as avg_accuracy_overall,
			AVG(avg_battery_level) as avg_battery_overall,
			SUM(total_movements) as total_movements_all_users
		FROM user_analytics_warehouse 
		WHERE date_created BETWEEN $1 AND $2
		GROUP BY date_created
		ORDER BY date_created DESC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var summaries []models.DailySummary
	for rows.Next() {
		var summary models.DailySummary
		err := rows.Scan(
			&summary.DateCreated,
			&summary.ActiveUsers,
			&summary.AvgSessionsPerUser,
			&summary.AvgAccuracyOverall,
			&summary.AvgBatteryOverall,
			&summary.TotalMovementsAllUsers,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return summaries, nil
}

// GetUserReportSummary получает сводку по пользователю
func (r *analyticsRepository) GetUserReportSummary(userID string, startDate, endDate time.Time) (*models.DailySummary, error) {
	query := `
		SELECT 
			date_created,
			COUNT(DISTINCT user_id) as active_users,
			AVG(total_sessions) as avg_sessions_per_user,
			AVG(avg_accuracy) as avg_accuracy_overall,
			AVG(avg_battery_level) as avg_battery_overall,
			SUM(total_movements) as total_movements_all_users
		FROM user_analytics_warehouse 
		WHERE user_id = $1 AND date_created BETWEEN $2 AND $3
		GROUP BY date_created
		ORDER BY date_created DESC
		LIMIT 1
	`

	row := r.db.QueryRow(query, userID, startDate, endDate)

	var summary models.DailySummary
	err := row.Scan(
		&summary.DateCreated,
		&summary.ActiveUsers,
		&summary.AvgSessionsPerUser,
		&summary.AvgAccuracyOverall,
		&summary.AvgBatteryOverall,
		&summary.TotalMovementsAllUsers,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("сводка для пользователя %s не найдена", userID)
		}
		return nil, fmt.Errorf("ошибка получения сводки: %w", err)
	}

	return &summary, nil
}

// CheckUserExists проверяет существование пользователя
func (r *analyticsRepository) CheckUserExists(userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_analytics_warehouse WHERE user_id = $1)`

	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования пользователя: %w", err)
	}

	return exists, nil
}
