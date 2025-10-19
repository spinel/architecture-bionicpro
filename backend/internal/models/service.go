package models

import (
	"time"
)

// Report представляет отчёт в системе
type Report struct {
	ID          int       `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Status      string    `json:"status" db:"status"`
	FileURL     string    `json:"file_url" db:"file_url"`
}
