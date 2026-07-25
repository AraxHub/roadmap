package domain

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User — пользователь платформы.
type User struct {
	ID             string
	Login          string
	PasswordHash   string
	Role           string
	IsBlocked      bool
	TelegramChatID *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsAdmin проверяет роль администратора.
func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// HasTelegram — привязан ли чат бота.
func (u User) HasTelegram() bool {
	return u.TelegramChatID != nil && *u.TelegramChatID != 0
}
