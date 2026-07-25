package auth

import (
	"context"
	"log/slog"
	"time"

	"roadmap/internal/domain"
	pkgauth "roadmap/internal/pkg/auth"
)

// UserStore — пользователи.
type UserStore interface {
	Count(ctx context.Context) (int, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	Create(ctx context.Context, login, passwordHash, role string) (*domain.User, error)
}

// RefreshStore — refresh-токены.
type RefreshStore interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeByHash(ctx context.Context, tokenHash string) error
}

// UseCase — логин, refresh, logout, создание ЛК.
type UseCase struct {
	users    UserStore
	refresh  RefreshStore
	tokens   *pkgauth.TokenService
	log      *slog.Logger
}

// New создаёт auth use case.
func New(users UserStore, refresh RefreshStore, tokens *pkgauth.TokenService, log *slog.Logger) *UseCase {
	return &UseCase{users: users, refresh: refresh, tokens: tokens, log: log}
}

// TokenPair — access + сырой refresh (cookie).
type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
	UserID           string
	Login            string
	Role             string
}

// CreatedUser — ответ после генерации ЛК (пароль один раз).
type CreatedUser struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
