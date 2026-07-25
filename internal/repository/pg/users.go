package pg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/lib/pq"
	"roadmap/internal/domain"
	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	UserTable = "users"

	UserColID             = "id"
	UserColLogin          = "login"
	UserColPasswordHash   = "password_hash"
	UserColRole           = "role"
	UserColIsBlocked      = "is_blocked"
	UserColTelegramChatID = "telegram_chat_id"
	UserColCreatedAt      = "created_at"
	UserColUpdatedAt      = "updated_at"
)

// UserColumns возвращает поля таблицы users для SELECT.
func UserColumns() []string {
	return []string{
		UserColID,
		UserColLogin,
		UserColPasswordHash,
		UserColRole,
		UserColIsBlocked,
		UserColTelegramChatID,
		UserColCreatedAt,
		UserColUpdatedAt,
	}
}

// UserSelect — список колонок (опционально с alias).
func UserSelect(alias string) string {
	return selectList(alias, UserColumns())
}

// UserRepo — доступ к таблице users.
type UserRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewUserRepo создаёт репозиторий пользователей.
func NewUserRepo(db *infrapg.DB, log *slog.Logger) *UserRepo {
	return &UserRepo{db: db, log: log}
}

// Exists проверяет наличие пользователя по id.
func (r *UserRepo) Exists(ctx context.Context, userID string) (bool, error) {
	q := "SELECT 1 FROM " + UserTable + " WHERE " + UserColID + " = $1"
	var one int
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Count возвращает число пользователей.
func (r *UserRepo) Count(ctx context.Context) (int, error) {
	q := "SELECT COUNT(*) FROM " + UserTable
	var n int
	err := r.db.QueryRowContext(ctx, q).Scan(&n)
	return n, err
}

// GetByID возвращает пользователя по id.
func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	q := "SELECT " + UserSelect("") + " FROM " + UserTable + " WHERE " + UserColID + " = $1"
	return r.scanOne(r.db.QueryRowContext(ctx, q, userID))
}

// GetByLogin возвращает пользователя по login.
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	q := "SELECT " + UserSelect("") + " FROM " + UserTable + " WHERE " + UserColLogin + " = $1"
	return r.scanOne(r.db.QueryRowContext(ctx, q, login))
}

// GetByTelegramChatID возвращает пользователя по chat_id бота.
func (r *UserRepo) GetByTelegramChatID(ctx context.Context, chatID int64) (*domain.User, error) {
	q := "SELECT " + UserSelect("") + " FROM " + UserTable + " WHERE " + UserColTelegramChatID + " = $1"
	return r.scanOne(r.db.QueryRowContext(ctx, q, chatID))
}

// List возвращает всех пользователей (админка), без password_hash в ответе — хеш всё равно в домене.
func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	q := "SELECT " + UserSelect("") + " FROM " + UserTable + " ORDER BY " + UserColCreatedAt
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.User, 0)
	for rows.Next() {
		u, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

// ListLearnersWithTelegram — обычные юзеры с привязанным TG (для рассылки ОС).
func (r *UserRepo) ListLearnersWithTelegram(ctx context.Context) ([]domain.User, error) {
	q := "SELECT " + UserSelect("") + " FROM " + UserTable +
		" WHERE " + UserColRole + " = $1 AND " + UserColTelegramChatID + " IS NOT NULL" +
		" ORDER BY " + UserColCreatedAt
	rows, err := r.db.QueryContext(ctx, q, domain.RoleUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.User, 0)
	for rows.Next() {
		u, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

// Create создаёт пользователя.
func (r *UserRepo) Create(ctx context.Context, login, passwordHash, role string) (*domain.User, error) {
	now := time.Now().UTC()
	q := "INSERT INTO " + UserTable +
		" (" + UserColLogin + ", " + UserColPasswordHash + ", " + UserColRole + ", " +
		UserColCreatedAt + ", " + UserColUpdatedAt + ")" +
		" VALUES ($1, $2, $3, $4, $5)" +
		" RETURNING " + UserSelect("")

	u, err := r.scanOne(r.db.QueryRowContext(ctx, q, login, passwordHash, role, now, now))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domain.ErrLoginTaken
		}
		return nil, err
	}
	return u, nil
}

// SetTelegramChatID привязывает chat_id к пользователю.
func (r *UserRepo) SetTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	q := "UPDATE " + UserTable + " SET " + UserColTelegramChatID + " = $2, " +
		UserColUpdatedAt + " = $3 WHERE " + UserColID + " = $1"
	res, err := r.db.ExecContext(ctx, q, userID, chatID, time.Now().UTC())
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetBlocked выставляет флаг блокировки главной.
func (r *UserRepo) SetBlocked(ctx context.Context, userID string, blocked bool) error {
	q := "UPDATE " + UserTable + " SET " + UserColIsBlocked + " = $2, " +
		UserColUpdatedAt + " = $3 WHERE " + UserColID + " = $1"
	_, err := r.db.ExecContext(ctx, q, userID, blocked, time.Now().UTC())
	return err
}

func (r *UserRepo) scanOne(row *sql.Row) (*domain.User, error) {
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}

func (r *UserRepo) scanRow(rows *sql.Rows) (*domain.User, error) {
	return scanUser(rows)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(s scannable) (*domain.User, error) {
	var u domain.User
	var chat sql.NullInt64
	err := s.Scan(
		&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.IsBlocked, &chat, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if chat.Valid {
		v := chat.Int64
		u.TelegramChatID = &v
	}
	return &u, nil
}
