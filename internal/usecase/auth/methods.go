package auth

import (
	"context"
	"errors"
	"time"

	"roadmap/internal/domain"
	pkgauth "roadmap/internal/pkg/auth"
)

// Login проверяет login/password и выдаёт пару токенов.
func (u *UseCase) Login(ctx context.Context, login, password string) (*TokenPair, error) {
	user, err := u.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !pkgauth.CheckPassword(user.PasswordHash, password) {
		return nil, domain.ErrInvalidCredentials
	}
	return u.issuePair(ctx, user)
}

// Refresh ротирует refresh-токен и выдаёт новую пару.
func (u *UseCase) Refresh(ctx context.Context, rawRefresh string) (*TokenPair, error) {
	if rawRefresh == "" {
		return nil, domain.ErrInvalidToken
	}
	hash := pkgauth.HashToken(rawRefresh)
	stored, err := u.refresh.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	if !stored.IsActive(time.Now().UTC()) {
		return nil, domain.ErrInvalidToken
	}
	user, err := u.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if err := u.refresh.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}
	return u.issuePair(ctx, user)
}

// Logout отзывает refresh по сырому значению из cookie.
func (u *UseCase) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	return u.refresh.RevokeByHash(ctx, pkgauth.HashToken(rawRefresh))
}

// CreateUser создаёт ЛК (admin). Пароль генерируется и возвращается один раз.
func (u *UseCase) CreateUser(ctx context.Context, login, role string) (*CreatedUser, error) {
	if role != domain.RoleUser && role != domain.RoleAdmin {
		role = domain.RoleUser
	}
	if login == "" {
		return nil, domain.ErrInvalidCredentials
	}
	plain, err := pkgauth.GeneratePassword(16)
	if err != nil {
		return nil, err
	}
	hash, err := pkgauth.HashPassword(plain)
	if err != nil {
		return nil, err
	}
	user, err := u.users.Create(ctx, login, hash, role)
	if err != nil {
		return nil, err
	}
	return &CreatedUser{
		ID:       user.ID,
		Login:    user.Login,
		Password: plain,
		Role:     user.Role,
	}, nil
}

// BootstrapAdmin создаёт первого админа, если users пуст и заданы credentials.
func (u *UseCase) BootstrapAdmin(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return nil
	}
	n, err := u.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := pkgauth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = u.users.Create(ctx, login, hash, domain.RoleAdmin)
	return err
}

// Me возвращает текущего пользователя.
func (u *UseCase) Me(ctx context.Context, userID string) (*domain.User, error) {
	return u.users.GetByID(ctx, userID)
}

func (u *UseCase) issuePair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	access, accessExp, err := u.tokens.IssueAccess(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	raw, hash, err := pkgauth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshExp := time.Now().UTC().Add(u.tokens.RefreshTTL())
	if _, err := u.refresh.Create(ctx, user.ID, hash, refreshExp); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     raw,
		RefreshExpiresAt: refreshExp,
		UserID:           user.ID,
		Login:            user.Login,
		Role:             user.Role,
	}, nil
}
