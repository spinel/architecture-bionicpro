package auth

import (
	"context"
)

// AuthService интерфейс для сервиса аутентификации
type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
	HasRole(userInfo *UserInfo, role string) bool
	IsAdmin(userInfo *UserInfo) bool
	IsProtheticUser(userInfo *UserInfo) bool
}
