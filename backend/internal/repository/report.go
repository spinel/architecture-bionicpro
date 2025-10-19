package repository

import (
	"bionicpro-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
)

// ReportRepository интерфейс для работы с отчетами в базе данных
type ReportRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]models.Report, error)
	GetByID(ctx context.Context, id int) (*models.Report, error)
	Create(ctx context.Context, report *models.Report) (int, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateFileURL(ctx context.Context, id int, fileURL string) error
}

// reportRepository реализация репозитория отчетов
type reportRepository struct {
	db *sql.DB
}

// NewReportRepository создает новый репозиторий отчетов
func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepository{db: db}
}

// GetByUserID возвращает все отчеты пользователя
func (r *reportRepository) GetByUserID(ctx context.Context, userID string) ([]models.Report, error) {
	query := `
		SELECT id, user_id, title, description, created_at, status, file_url
		FROM reports 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var report models.Report
		var fileURL sql.NullString

		err := rows.Scan(
			&report.ID,
			&report.UserID,
			&report.Title,
			&report.Description,
			&report.CreatedAt,
			&report.Status,
			&fileURL,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}

		if fileURL.Valid {
			report.FileURL = fileURL.String
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return reports, nil
}

// GetByID возвращает отчет по ID
func (r *reportRepository) GetByID(ctx context.Context, id int) (*models.Report, error) {
	query := `
		SELECT id, user_id, title, description, created_at, status, file_url
		FROM reports 
		WHERE id = $1
	`

	var report models.Report
	var fileURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.UserID,
		&report.Title,
		&report.Description,
		&report.CreatedAt,
		&report.Status,
		&fileURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("отчет не найден")
		}
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	if fileURL.Valid {
		report.FileURL = fileURL.String
	}

	return &report, nil
}

// Create создает новый отчет
func (r *reportRepository) Create(ctx context.Context, report *models.Report) (int, error) {
	query := `
		INSERT INTO reports (user_id, title, description, created_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		report.UserID,
		report.Title,
		report.Description,
		report.CreatedAt,
		report.Status,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания отчета: %w", err)
	}

	return id, nil
}

// UpdateStatus обновляет статус отчета
func (r *reportRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE reports SET status = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отчет не найден")
	}

	return nil
}

// UpdateFileURL обновляет URL файла отчета
func (r *reportRepository) UpdateFileURL(ctx context.Context, id int, fileURL string) error {
	query := `UPDATE reports SET file_url = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, fileURL, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления URL файла: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("отчет не найден")
	}

	return nil
}
