package domain

import "time"

// RefreshToken — сессия refresh-токена (в БД хранится hash).
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// IsActive — не отозван и не истёк.
func (t RefreshToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return t.ExpiresAt.After(now)
}
