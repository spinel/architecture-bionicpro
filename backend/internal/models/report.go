package models

import "time"

// Report представляет отчёт об использовании протеза
type Report struct {
	ReportID    string    `json:"report_id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DeviceID    string    `json:"device_id"`
	DeviceType  string    `json:"device_type"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Metrics     Metrics   `json:"metrics"`
	CreatedAt   time.Time `json:"created_at"`
}

// Metrics содержит агрегированные метрики использования
type Metrics struct {
	TotalUsageHours    float64 `json:"total_usage_hours"`
	AverageDailyHours  float64 `json:"average_daily_hours"`
	BatteryChanges     int     `json:"battery_changes"`
	MaintenanceAlerts  int     `json:"maintenance_alerts"`
	MovementActivities int     `json:"movement_activities"`
	ErrorCount         int     `json:"error_count"`
	ComfortScore       float64 `json:"comfort_score"`
	EfficiencyScore    float64 `json:"efficiency_score"`
}

// ReportFilter содержит параметры фильтрации отчётов
type ReportFilter struct {
	UserID    string    `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// ReportResponse содержит ответ API с отчётами
type ReportResponse struct {
	Reports    []Report `json:"reports"`
	TotalCount int      `json:"total_count"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

// UserClaims содержит информацию из JWT токена
type UserClaims struct {
	UserID   string   `json:"sub"`
	Username string   `json:"preferred_username"`
	Email    string   `json:"email"`
	Roles    []string `json:"realm_access.roles"`
}
