package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// UserRepository интерфейс для работы с пользователями в базе данных
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

// User представляет пользователя в базе данных
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// userRepository реализация репозитория пользователей
type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID возвращает пользователя по ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	return &user, nil
}

// Create создает нового пользователя
func (r *userRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return nil
}

// Update обновляет пользователя
func (r *userRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET email = $2, first_name = $3, last_name = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("пользователь не найден")
	}

	return nil
}
