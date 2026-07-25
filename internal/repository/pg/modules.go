package pg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/lib/pq"
	"roadmap/internal/domain"
	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	ModuleTable = "modules"

	ModuleColID          = "id"
	ModuleColSprintID    = "sprint_id"
	ModuleColSlug        = "slug"
	ModuleColTitle       = "title"
	ModuleColDescription = "description"
	ModuleColPosition    = "position"
	ModuleColIsPublished = "is_published"
)

// ModuleColumns возвращает поля таблицы modules для SELECT.
func ModuleColumns() []string {
	return []string{
		ModuleColID,
		ModuleColSprintID,
		ModuleColSlug,
		ModuleColTitle,
		ModuleColDescription,
		ModuleColPosition,
		ModuleColIsPublished,
	}
}

// ModuleSelect — список колонок (опционально с alias).
func ModuleSelect(alias string) string {
	return selectList(alias, ModuleColumns())
}

// ModuleRepo — доступ к таблице modules.
type ModuleRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewModuleRepo создаёт репозиторий модулей.
func NewModuleRepo(db *infrapg.DB, log *slog.Logger) *ModuleRepo {
	return &ModuleRepo{db: db, log: log}
}

// ListPublishedBySprint возвращает опубликованные модули спринта.
func (r *ModuleRepo) ListPublishedBySprint(ctx context.Context, sprintID string) ([]domain.Module, error) {
	q := "SELECT " + ModuleSelect("") + " FROM " + ModuleTable +
		" WHERE " + ModuleColSprintID + " = $1 AND " + ModuleColIsPublished +
		" ORDER BY " + ModuleColPosition + " ASC"

	rows, err := r.db.QueryContext(ctx, q, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Module
	for rows.Next() {
		var m domain.Module
		if err := rows.Scan(&m.ID, &m.SprintID, &m.Slug, &m.Title, &m.Description, &m.Position, &m.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetPublishedBySlug возвращает опубликованный модуль по slug.
func (r *ModuleRepo) GetPublishedBySlug(ctx context.Context, slug string) (*domain.Module, error) {
	q := "SELECT " + ModuleSelect("") + " FROM " + ModuleTable +
		" WHERE " + ModuleColSlug + " = $1 AND " + ModuleColIsPublished

	var m domain.Module
	err := r.db.QueryRowContext(ctx, q, slug).Scan(
		&m.ID, &m.SprintID, &m.Slug, &m.Title, &m.Description, &m.Position, &m.IsPublished,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListAllBySprint возвращает все модули спринта (админка).
func (r *ModuleRepo) ListAllBySprint(ctx context.Context, sprintID string) ([]domain.Module, error) {
	q := "SELECT " + ModuleSelect("") + " FROM " + ModuleTable +
		" WHERE " + ModuleColSprintID + " = $1 ORDER BY " + ModuleColPosition + " ASC"
	rows, err := r.db.QueryContext(ctx, q, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Module
	for rows.Next() {
		var m domain.Module
		if err := rows.Scan(&m.ID, &m.SprintID, &m.Slug, &m.Title, &m.Description, &m.Position, &m.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// NextPosition — следующий position модуля внутри спринта.
func (r *ModuleRepo) NextPosition(ctx context.Context, sprintID string) (int, error) {
	q := "SELECT COALESCE(MAX(" + ModuleColPosition + "), 0) + 1 FROM " + ModuleTable +
		" WHERE " + ModuleColSprintID + " = $1"
	var n int
	err := r.db.QueryRowContext(ctx, q, sprintID).Scan(&n)
	return n, err
}

// Create создаёт модуль.
func (r *ModuleRepo) Create(ctx context.Context, m domain.Module) (*domain.Module, error) {
	q := "INSERT INTO " + ModuleTable +
		" (" + ModuleColSprintID + ", " + ModuleColSlug + ", " + ModuleColTitle + ", " +
		ModuleColDescription + ", " + ModuleColPosition + ", " + ModuleColIsPublished + ")" +
		" VALUES ($1,$2,$3,$4,$5,$6) RETURNING " + ModuleSelect("")
	return r.scanOne(r.db.QueryRowContext(ctx, q, m.SprintID, m.Slug, m.Title, m.Description, m.Position, m.IsPublished))
}

// Update обновляет модуль.
func (r *ModuleRepo) Update(ctx context.Context, m domain.Module) (*domain.Module, error) {
	q := "UPDATE " + ModuleTable + " SET " +
		ModuleColSlug + "=$1, " + ModuleColTitle + "=$2, " + ModuleColDescription + "=$3, " +
		ModuleColPosition + "=$4, " + ModuleColIsPublished + "=$5" +
		" WHERE " + ModuleColID + "=$6 RETURNING " + ModuleSelect("")
	return r.scanOne(r.db.QueryRowContext(ctx, q, m.Slug, m.Title, m.Description, m.Position, m.IsPublished, m.ID))
}

// Delete удаляет модуль.
func (r *ModuleRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM "+ModuleTable+" WHERE "+ModuleColID+"=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ModuleRepo) scanOne(row *sql.Row) (*domain.Module, error) {
	var m domain.Module
	err := row.Scan(&m.ID, &m.SprintID, &m.Slug, &m.Title, &m.Description, &m.Position, &m.IsPublished)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domain.ErrConflict
		}
		return nil, err
	}
	return &m, nil
}
