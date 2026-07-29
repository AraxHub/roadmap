package pg

import (
	"context"
	"log/slog"
	"time"

	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	UserSubmoduleProgressTable = "user_submodule_progress"

	UserSubmoduleProgressColUserID      = "user_id"
	UserSubmoduleProgressColSubmoduleID = "submodule_id"
	UserSubmoduleProgressColCompletedAt = "completed_at"
)

// UserSubmoduleProgressColumns возвращает поля таблицы user_submodule_progress для SELECT.
func UserSubmoduleProgressColumns() []string {
	return []string{
		UserSubmoduleProgressColUserID,
		UserSubmoduleProgressColSubmoduleID,
		UserSubmoduleProgressColCompletedAt,
	}
}

// UserSubmoduleProgressSelect — список колонок (опционально с alias).
func UserSubmoduleProgressSelect(alias string) string {
	return selectList(alias, UserSubmoduleProgressColumns())
}

// UserSubmoduleProgressRepo — доступ к таблице user_submodule_progress.
type UserSubmoduleProgressRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewUserSubmoduleProgressRepo создаёт репозиторий прогресса.
func NewUserSubmoduleProgressRepo(db *infrapg.DB, log *slog.Logger) *UserSubmoduleProgressRepo {
	return &UserSubmoduleProgressRepo{db: db, log: log}
}

// ListCompletedSubmoduleIDs возвращает множество завершённых подмодулей пользователя.
func (r *UserSubmoduleProgressRepo) ListCompletedSubmoduleIDs(ctx context.Context, userID string) (map[string]bool, error) {
	q := "SELECT " + UserSubmoduleProgressColSubmoduleID +
		" FROM " + UserSubmoduleProgressTable +
		" WHERE " + UserSubmoduleProgressColUserID + " = $1"

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// ListAllCompletedByUser возвращает завершённые подмодули по всем пользователям.
func (r *UserSubmoduleProgressRepo) ListAllCompletedByUser(ctx context.Context) (map[string]map[string]bool, error) {
	q := "SELECT " + UserSubmoduleProgressColUserID + ", " + UserSubmoduleProgressColSubmoduleID +
		" FROM " + UserSubmoduleProgressTable

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]map[string]bool)
	for rows.Next() {
		var userID, submoduleID string
		if err := rows.Scan(&userID, &submoduleID); err != nil {
			return nil, err
		}
		m := out[userID]
		if m == nil {
			m = make(map[string]bool)
			out[userID] = m
		}
		m[submoduleID] = true
	}
	return out, rows.Err()
}

// Complete фиксирует завершение подмодуля (идемпотентно).
func (r *UserSubmoduleProgressRepo) Complete(ctx context.Context, userID, submoduleID string) error {
	q := "INSERT INTO " + UserSubmoduleProgressTable +
		" (" + UserSubmoduleProgressColUserID + ", " +
		UserSubmoduleProgressColSubmoduleID + ", " +
		UserSubmoduleProgressColCompletedAt + ")" +
		" VALUES ($1, $2, $3)" +
		" ON CONFLICT (" + UserSubmoduleProgressColUserID + ", " +
		UserSubmoduleProgressColSubmoduleID + ") DO NOTHING"

	_, err := r.db.ExecContext(ctx, q, userID, submoduleID, time.Now().UTC())
	return err
}
