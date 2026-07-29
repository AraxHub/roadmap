package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"roadmap/internal/domain"
	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	SubmoduleContentTable = "submodule_contents"

	SubmoduleContentColSubmoduleID = "submodule_id"
	SubmoduleContentColBlocks      = "blocks"
	SubmoduleContentColUpdatedAt   = "updated_at"
)

// SubmoduleContentColumns возвращает поля таблицы submodule_contents для SELECT.
func SubmoduleContentColumns() []string {
	return []string{
		SubmoduleContentColSubmoduleID,
		SubmoduleContentColBlocks,
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

func scanSubmoduleContent(submoduleID string, raw []byte, updatedAt time.Time) (*domain.SubmoduleContent, error) {
	blocks := make([]domain.ContentBlock, 0)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &blocks); err != nil {
			return nil, err
		}
	}
	if blocks == nil {
		blocks = []domain.ContentBlock{}
	}
	return &domain.SubmoduleContent{
		SubmoduleID: submoduleID,
		Blocks:      blocks,
		UpdatedAt:   updatedAt,
	}, nil
}

// GetBySubmoduleID возвращает блоки подмодуля.
func (r *SubmoduleContentRepo) GetBySubmoduleID(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error) {
	q := "SELECT " + SubmoduleContentSelect("") + " FROM " + SubmoduleContentTable +
		" WHERE " + SubmoduleContentColSubmoduleID + " = $1"

	var id string
	var raw []byte
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, q, submoduleID).Scan(&id, &raw, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return scanSubmoduleContent(id, raw, updatedAt)
}

// Upsert сохраняет блоки подмодуля.
func (r *SubmoduleContentRepo) Upsert(ctx context.Context, submoduleID string, blocks []domain.ContentBlock) (*domain.SubmoduleContent, error) {
	if blocks == nil {
		blocks = []domain.ContentBlock{}
	}
	raw, err := json.Marshal(blocks)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	q := "INSERT INTO " + SubmoduleContentTable +
		" (" + SubmoduleContentColSubmoduleID + ", " + SubmoduleContentColBlocks + ", " + SubmoduleContentColUpdatedAt + ")" +
		" VALUES ($1, $2::jsonb, $3)" +
		" ON CONFLICT (" + SubmoduleContentColSubmoduleID + ") DO UPDATE SET " +
		SubmoduleContentColBlocks + " = EXCLUDED." + SubmoduleContentColBlocks + ", " +
		SubmoduleContentColUpdatedAt + " = EXCLUDED." + SubmoduleContentColUpdatedAt +
		" RETURNING " + SubmoduleContentSelect("")

	var id string
	var outRaw []byte
	var updatedAt time.Time
	err = r.db.QueryRowContext(ctx, q, submoduleID, raw, now).Scan(&id, &outRaw, &updatedAt)
	if err != nil {
		return nil, err
	}
	return scanSubmoduleContent(id, outRaw, updatedAt)
}
