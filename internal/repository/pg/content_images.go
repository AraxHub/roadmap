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
	ContentImageTable = "content_images"

	ContentImageColID          = "id"
	ContentImageColSubmoduleID = "submodule_id"
	ContentImageColMimeType    = "mime_type"
	ContentImageColBytes       = "bytes"
	ContentImageColByteSize    = "byte_size"
	ContentImageColCreatedAt   = "created_at"
)

// ContentImageColumns — колонки для SELECT метаданных + bytes.
func ContentImageColumns() []string {
	return []string{
		ContentImageColID,
		ContentImageColSubmoduleID,
		ContentImageColMimeType,
		ContentImageColBytes,
		ContentImageColByteSize,
		ContentImageColCreatedAt,
	}
}

func ContentImageSelect(alias string) string {
	return selectList(alias, ContentImageColumns())
}

// ContentImageRepo — картинки контента в Postgres.
type ContentImageRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewContentImageRepo создаёт репозиторий картинок.
func NewContentImageRepo(db *infrapg.DB, log *slog.Logger) *ContentImageRepo {
	return &ContentImageRepo{db: db, log: log}
}

// Create сохраняет картинку, id генерирует БД.
func (r *ContentImageRepo) Create(ctx context.Context, submoduleID, mimeType string, data []byte) (*domain.ContentImage, error) {
	now := time.Now().UTC()
	q := "INSERT INTO " + ContentImageTable +
		" (" + ContentImageColSubmoduleID + ", " + ContentImageColMimeType + ", " +
		ContentImageColBytes + ", " + ContentImageColByteSize + ", " + ContentImageColCreatedAt + ")" +
		" VALUES ($1, $2, $3, $4, $5)" +
		" RETURNING " + ContentImageSelect("")

	var img domain.ContentImage
	err := r.db.QueryRowContext(ctx, q, submoduleID, mimeType, data, len(data), now).Scan(
		&img.ID, &img.SubmoduleID, &img.MimeType, &img.Bytes, &img.ByteSize, &img.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// GetByID возвращает картинку целиком.
func (r *ContentImageRepo) GetByID(ctx context.Context, id string) (*domain.ContentImage, error) {
	q := "SELECT " + ContentImageSelect("") + " FROM " + ContentImageTable +
		" WHERE " + ContentImageColID + " = $1"

	var img domain.ContentImage
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&img.ID, &img.SubmoduleID, &img.MimeType, &img.Bytes, &img.ByteSize, &img.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// ExistsForSubmodule проверяет, что image принадлежит submodule.
func (r *ContentImageRepo) ExistsForSubmodule(ctx context.Context, imageID, submoduleID string) (bool, error) {
	q := "SELECT 1 FROM " + ContentImageTable +
		" WHERE " + ContentImageColID + " = $1 AND " + ContentImageColSubmoduleID + " = $2"
	var one int
	err := r.db.QueryRowContext(ctx, q, imageID, submoduleID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Delete удаляет картинку.
func (r *ContentImageRepo) Delete(ctx context.Context, id string) error {
	q := "DELETE FROM " + ContentImageTable + " WHERE " + ContentImageColID + " = $1"
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
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
