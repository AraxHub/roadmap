package pg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"roadmap/internal/domain"
	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	RefreshTokenTable = "refresh_tokens"

	RefreshTokenColID        = "id"
	RefreshTokenColUserID    = "user_id"
	RefreshTokenColTokenHash = "token_hash"
	RefreshTokenColExpiresAt = "expires_at"
	RefreshTokenColRevokedAt = "revoked_at"
	RefreshTokenColCreatedAt = "created_at"
)

// RefreshTokenColumns возвращает поля таблицы refresh_tokens для SELECT.
func RefreshTokenColumns() []string {
	return []string{
		RefreshTokenColID,
		RefreshTokenColUserID,
		RefreshTokenColTokenHash,
		RefreshTokenColExpiresAt,
		RefreshTokenColRevokedAt,
		RefreshTokenColCreatedAt,
	}
}

// RefreshTokenSelect — список колонок (опционально с alias).
func RefreshTokenSelect(alias string) string {
	return selectList(alias, RefreshTokenColumns())
}

// RefreshTokenRepo — доступ к таблице refresh_tokens.
type RefreshTokenRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewRefreshTokenRepo создаёт репозиторий refresh-токенов.
func NewRefreshTokenRepo(db *infrapg.DB, log *slog.Logger) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db, log: log}
}

// Create сохраняет hash refresh-токена.
func (r *RefreshTokenRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error) {
	now := time.Now().UTC()
	q := "INSERT INTO " + RefreshTokenTable +
		" (" + RefreshTokenColUserID + ", " + RefreshTokenColTokenHash + ", " +
		RefreshTokenColExpiresAt + ", " + RefreshTokenColCreatedAt + ")" +
		" VALUES ($1, $2, $3, $4)" +
		" RETURNING " + RefreshTokenSelect("")

	return r.scanOne(r.db.QueryRowContext(ctx, q, userID, tokenHash, expiresAt, now))
}

// GetByHash возвращает токен по hash.
func (r *RefreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	q := "SELECT " + RefreshTokenSelect("") + " FROM " + RefreshTokenTable +
		" WHERE " + RefreshTokenColTokenHash + " = $1"
	return r.scanOne(r.db.QueryRowContext(ctx, q, tokenHash))
}

// Revoke помечает токен отозванным.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	q := "UPDATE " + RefreshTokenTable +
		" SET " + RefreshTokenColRevokedAt + " = $1" +
		" WHERE " + RefreshTokenColID + " = $2 AND " + RefreshTokenColRevokedAt + " IS NULL"
	_, err := r.db.ExecContext(ctx, q, now, id)
	return err
}

// RevokeByHash отзывает токен по hash.
func (r *RefreshTokenRepo) RevokeByHash(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()
	q := "UPDATE " + RefreshTokenTable +
		" SET " + RefreshTokenColRevokedAt + " = $1" +
		" WHERE " + RefreshTokenColTokenHash + " = $2 AND " + RefreshTokenColRevokedAt + " IS NULL"
	_, err := r.db.ExecContext(ctx, q, now, tokenHash)
	return err
}

func (r *RefreshTokenRepo) scanOne(row *sql.Row) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	var revoked sql.NullTime
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &revoked, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if revoked.Valid {
		t.RevokedAt = &revoked.Time
	}
	return &t, nil
}
