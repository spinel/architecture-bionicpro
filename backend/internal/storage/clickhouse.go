package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spinel/architecture-bionicpro/backend/internal/config"
	"github.com/spinel/architecture-bionicpro/backend/internal/models"
)

// ClickHouseStorage управляет подключением к ClickHouse
type ClickHouseStorage struct {
	conn driver.Conn
}

// NewClickHouseStorage создаёт новое подключение к ClickHouse
func NewClickHouseStorage(cfg *config.Config) (*ClickHouseStorage, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.User,
			Password: cfg.ClickHouse.Password,
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true, // Для разработки, в продакшн использовать валидные сертификаты
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      time.Second * 30,
		MaxOpenConns:     5,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Проверяем подключение
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseStorage{conn: conn}, nil
}

// Close закрывает подключение к ClickHouse
func (s *ClickHouseStorage) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// GetReportsByUser получает отчёты для пользователя из OLAP базы
func (s *ClickHouseStorage) GetReportsByUser(ctx context.Context, userID string, limit int, offset int) ([]models.Report, error) {
	query := `
		SELECT 
			report_id,
			user_id,
			username,
			device_id,
			device_type,
			period_start,
			period_end,
			total_usage_hours,
			average_daily_hours,
			battery_changes,
			maintenance_alerts,
			movement_activities,
			error_count,
			comfort_score,
			efficiency_score,
			created_at
		FROM reports_summary
		WHERE user_id = ?
		ORDER BY period_start DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.conn.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []models.Report

	for rows.Next() {
		var report models.Report
		err := rows.Scan(
			&report.ReportID,
			&report.UserID,
			&report.Username,
			&report.DeviceID,
			&report.DeviceType,
			&report.PeriodStart,
			&report.PeriodEnd,
			&report.Metrics.TotalUsageHours,
			&report.Metrics.AverageDailyHours,
			&report.Metrics.BatteryChanges,
			&report.Metrics.MaintenanceAlerts,
			&report.Metrics.MovementActivities,
			&report.Metrics.ErrorCount,
			&report.Metrics.ComfortScore,
			&report.Metrics.EfficiencyScore,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report row: %w", err)
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating report rows: %w", err)
	}

	return reports, nil
}

// GetReportByID получает конкретный отчёт по ID
func (s *ClickHouseStorage) GetReportByID(ctx context.Context, reportID string) (*models.Report, error) {
	query := `
		SELECT 
			report_id,
			user_id,
			username,
			device_id,
			device_type,
			period_start,
			period_end,
			total_usage_hours,
			average_daily_hours,
			battery_changes,
			maintenance_alerts,
			movement_activities,
			error_count,
			comfort_score,
			efficiency_score,
			created_at
		FROM reports_summary
		WHERE report_id = ?
		LIMIT 1
	`

	var report models.Report
	row := s.conn.QueryRow(ctx, query, reportID)

	err := row.Scan(
		&report.ReportID,
		&report.UserID,
		&report.Username,
		&report.DeviceID,
		&report.DeviceType,
		&report.PeriodStart,
		&report.PeriodEnd,
		&report.Metrics.TotalUsageHours,
		&report.Metrics.AverageDailyHours,
		&report.Metrics.BatteryChanges,
		&report.Metrics.MaintenanceAlerts,
		&report.Metrics.MovementActivities,
		&report.Metrics.ErrorCount,
		&report.Metrics.ComfortScore,
		&report.Metrics.EfficiencyScore,
		&report.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	return &report, nil
}

// CountReportsByUser получает количество отчётов для пользователя
func (s *ClickHouseStorage) CountReportsByUser(ctx context.Context, userID string) (int, error) {
	query := `SELECT count() FROM reports_summary WHERE user_id = ?`

	var count int
	row := s.conn.QueryRow(ctx, query, userID)

	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}

	return count, nil
}
