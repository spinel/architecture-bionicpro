package models

import (
	"time"
)

// UserReport представляет отчёт пользователя
type UserReport struct {
	UserID          string    `json:"user_id" db:"user_id"`
	DateCreated     time.Time `json:"date_created" db:"date_created"`
	TotalSessions   int       `json:"total_sessions" db:"total_sessions"`
	AvgAccuracy     float64   `json:"avg_accuracy" db:"avg_accuracy"`
	AvgBatteryLevel float64   `json:"avg_battery_level" db:"avg_battery_level"`
	AvgTemperature  float64   `json:"avg_temperature" db:"avg_temperature"`
	FirstActivity   time.Time `json:"first_activity" db:"first_activity"`
	LastActivity    time.Time `json:"last_activity" db:"last_activity"`
	TotalMovements  int       `json:"total_movements" db:"total_movements"`
	AccuracyStatus  string    `json:"accuracy_status" db:"accuracy_status"`
	BatteryStatus   string    `json:"battery_status" db:"battery_status"`
	SessionDuration float64   `json:"session_duration_hours" db:"session_duration_hours"`
}

// DailySummary представляет сводку по дням
type DailySummary struct {
	DateCreated            time.Time `json:"date_created" db:"date_created"`
	ActiveUsers            int       `json:"active_users" db:"active_users"`
	AvgSessionsPerUser     float64   `json:"avg_sessions_per_user" db:"avg_sessions_per_user"`
	AvgAccuracyOverall     float64   `json:"avg_accuracy_overall" db:"avg_accuracy_overall"`
	AvgBatteryOverall      float64   `json:"avg_battery_overall" db:"avg_battery_overall"`
	TotalMovementsAllUsers int       `json:"total_movements_all_users" db:"total_movements_all_users"`
}

// ReportRequest представляет запрос на получение отчёта
type ReportRequest struct {
	UserID    string    `json:"user_id" form:"user_id" binding:"required"`
	StartDate time.Time `json:"start_date" form:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" form:"end_date" binding:"required"`
}

// ReportResponse представляет ответ с отчётом
type ReportResponse struct {
	Success bool         `json:"success"`
	Data    []UserReport `json:"data,omitempty"`
	Summary DailySummary `json:"summary,omitempty"`
	Error   string       `json:"error,omitempty"`
	Meta    ReportMeta   `json:"meta"`
}

// ReportMeta содержит метаинформацию об отчёте
type ReportMeta struct {
	GeneratedAt time.Time `json:"generated_at"`
	UserID      string    `json:"user_id"`
	DateRange   string    `json:"date_range"`
	RecordCount int       `json:"record_count"`
}

// HealthResponse представляет ответ о состоянии API
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
}

// ErrorResponse представляет ответ об ошибке
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}
