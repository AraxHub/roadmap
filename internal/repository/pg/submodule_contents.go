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
	SubmoduleContentTable = "submodule_contents"

	SubmoduleContentColSubmoduleID = "submodule_id"
	SubmoduleContentColBodyMD      = "body_md"
	SubmoduleContentColUpdatedAt   = "updated_at"
)

// SubmoduleContentColumns возвращает поля таблицы submodule_contents для SELECT.
func SubmoduleContentColumns() []string {
	return []string{
		SubmoduleContentColSubmoduleID,
		SubmoduleContentColBodyMD,
		SubmoduleContentColUpdatedAt,
	}
}

// SubmoduleContentSelect — список колонок (опционально с alias).
func SubmoduleContentSelect(alias string) string {
	return selectList(alias, SubmoduleContentColumns())
}

// SubmoduleContentRepo — доступ к таблице submodule_contents.
type SubmoduleContentRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewSubmoduleContentRepo создаёт репозиторий контента.
func NewSubmoduleContentRepo(db *infrapg.DB, log *slog.Logger) *SubmoduleContentRepo {
	return &SubmoduleContentRepo{db: db, log: log}
}

// GetBySubmoduleID возвращает markdown подмодуля.
func (r *SubmoduleContentRepo) GetBySubmoduleID(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error) {
	q := "SELECT " + SubmoduleContentSelect("") + " FROM " + SubmoduleContentTable +
		" WHERE " + SubmoduleContentColSubmoduleID + " = $1"

	var c domain.SubmoduleContent
	err := r.db.QueryRowContext(ctx, q, submoduleID).Scan(&c.SubmoduleID, &c.BodyMD, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Upsert сохраняет markdown подмодуля.
func (r *SubmoduleContentRepo) Upsert(ctx context.Context, submoduleID, bodyMD string) (*domain.SubmoduleContent, error) {
	now := time.Now().UTC()
	q := "INSERT INTO " + SubmoduleContentTable +
		" (" + SubmoduleContentColSubmoduleID + ", " + SubmoduleContentColBodyMD + ", " + SubmoduleContentColUpdatedAt + ")" +
		" VALUES ($1, $2, $3)" +
		" ON CONFLICT (" + SubmoduleContentColSubmoduleID + ") DO UPDATE SET " +
		SubmoduleContentColBodyMD + " = EXCLUDED." + SubmoduleContentColBodyMD + ", " +
		SubmoduleContentColUpdatedAt + " = EXCLUDED." + SubmoduleContentColUpdatedAt +
		" RETURNING " + SubmoduleContentSelect("")

	var c domain.SubmoduleContent
	err := r.db.QueryRowContext(ctx, q, submoduleID, bodyMD, now).Scan(&c.SubmoduleID, &c.BodyMD, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
