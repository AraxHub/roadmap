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
	SubmoduleTable = "submodules"

	SubmoduleColID          = "id"
	SubmoduleColModuleID    = "module_id"
	SubmoduleColSlug        = "slug"
	SubmoduleColTitle       = "title"
	SubmoduleColPosition    = "position"
	SubmoduleColIsPublished = "is_published"
)

// SubmoduleColumns возвращает поля таблицы submodules для SELECT.
func SubmoduleColumns() []string {
	return []string{
		SubmoduleColID,
		SubmoduleColModuleID,
		SubmoduleColSlug,
		SubmoduleColTitle,
		SubmoduleColPosition,
		SubmoduleColIsPublished,
	}
}

// SubmoduleSelect — список колонок (опционально с alias).
func SubmoduleSelect(alias string) string {
	return selectList(alias, SubmoduleColumns())
}

// SubmoduleRepo — доступ к таблице submodules.
type SubmoduleRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewSubmoduleRepo создаёт репозиторий подмодулей.
func NewSubmoduleRepo(db *infrapg.DB, log *slog.Logger) *SubmoduleRepo {
	return &SubmoduleRepo{db: db, log: log}
}

// ListPublishedByModule возвращает опубликованные подмодули модуля.
func (r *SubmoduleRepo) ListPublishedByModule(ctx context.Context, moduleID string) ([]domain.Submodule, error) {
	q := "SELECT " + SubmoduleSelect("") + " FROM " + SubmoduleTable +
		" WHERE " + SubmoduleColModuleID + " = $1 AND " + SubmoduleColIsPublished +
		" ORDER BY " + SubmoduleColPosition + " ASC"

	rows, err := r.db.QueryContext(ctx, q, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Submodule
	for rows.Next() {
		var sm domain.Submodule
		if err := rows.Scan(&sm.ID, &sm.ModuleID, &sm.Slug, &sm.Title, &sm.Position, &sm.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, sm)
	}
	return out, rows.Err()
}

// GetPublishedBySlugs возвращает подмодуль и его модуль по slug'ам.
func (r *SubmoduleRepo) GetPublishedBySlugs(ctx context.Context, moduleSlug, submoduleSlug string) (*domain.Submodule, *domain.Module, error) {
	q := "SELECT " +
		SubmoduleSelect("sm") + ", " + ModuleSelect("m") +
		" FROM " + SubmoduleTable + " sm" +
		" JOIN " + ModuleTable + " m ON m." + ModuleColID + " = sm." + SubmoduleColModuleID +
		" WHERE m." + ModuleColSlug + " = $1 AND sm." + SubmoduleColSlug + " = $2" +
		" AND sm." + SubmoduleColIsPublished + " AND m." + ModuleColIsPublished

	var sm domain.Submodule
	var m domain.Module
	err := r.db.QueryRowContext(ctx, q, moduleSlug, submoduleSlug).Scan(
		&sm.ID, &sm.ModuleID, &sm.Slug, &sm.Title, &sm.Position, &sm.IsPublished,
		&m.ID, &m.SprintID, &m.Slug, &m.Title, &m.Description, &m.Position, &m.IsPublished,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	return &sm, &m, nil
}

// ListPublishedChain возвращает опубликованные подмодули в глобальном порядке.
func (r *SubmoduleRepo) ListPublishedChain(ctx context.Context) ([]domain.ChainItem, error) {
	q := "SELECT " +
		"sm." + SubmoduleColID + ", sm." + SubmoduleColSlug + ", sm." + SubmoduleColTitle + ", " +
		"m." + ModuleColID + ", m." + ModuleColSlug + ", m." + ModuleColTitle + ", " +
		"s." + SprintColID + ", s." + SprintColSlug + ", s." + SprintColTitle +
		" FROM " + SubmoduleTable + " sm" +
		" JOIN " + ModuleTable + " m ON m." + ModuleColID + " = sm." + SubmoduleColModuleID +
		" JOIN " + SprintTable + " s ON s." + SprintColID + " = m." + ModuleColSprintID +
		" WHERE sm." + SubmoduleColIsPublished +
		" AND m." + ModuleColIsPublished +
		" AND s." + SprintColIsPublished +
		" ORDER BY s." + SprintColPosition + " ASC, m." + ModuleColPosition + " ASC, sm." + SubmoduleColPosition + " ASC"

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ChainItem
	for rows.Next() {
		var item domain.ChainItem
		if err := rows.Scan(
			&item.SubmoduleID, &item.SubmoduleSlug, &item.SubmoduleTitle,
			&item.ModuleID, &item.ModuleSlug, &item.ModuleTitle,
			&item.SprintID, &item.SprintSlug, &item.SprintTitle,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListAllByModule возвращает все подмодули модуля (админка).
func (r *SubmoduleRepo) ListAllByModule(ctx context.Context, moduleID string) ([]domain.Submodule, error) {
	q := "SELECT " + SubmoduleSelect("") + " FROM " + SubmoduleTable +
		" WHERE " + SubmoduleColModuleID + " = $1 ORDER BY " + SubmoduleColPosition + " ASC"
	rows, err := r.db.QueryContext(ctx, q, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Submodule
	for rows.Next() {
		var sm domain.Submodule
		if err := rows.Scan(&sm.ID, &sm.ModuleID, &sm.Slug, &sm.Title, &sm.Position, &sm.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, sm)
	}
	return out, rows.Err()
}

// NextPosition — следующий position подмодуля внутри модуля.
func (r *SubmoduleRepo) NextPosition(ctx context.Context, moduleID string) (int, error) {
	q := "SELECT COALESCE(MAX(" + SubmoduleColPosition + "), 0) + 1 FROM " + SubmoduleTable +
		" WHERE " + SubmoduleColModuleID + " = $1"
	var n int
	err := r.db.QueryRowContext(ctx, q, moduleID).Scan(&n)
	return n, err
}

// Create создаёт подмодуль.
func (r *SubmoduleRepo) Create(ctx context.Context, sm domain.Submodule) (*domain.Submodule, error) {
	q := "INSERT INTO " + SubmoduleTable +
		" (" + SubmoduleColModuleID + ", " + SubmoduleColSlug + ", " + SubmoduleColTitle + ", " +
		SubmoduleColPosition + ", " + SubmoduleColIsPublished + ")" +
		" VALUES ($1,$2,$3,$4,$5) RETURNING " + SubmoduleSelect("")
	return r.scanSubmodule(r.db.QueryRowContext(ctx, q, sm.ModuleID, sm.Slug, sm.Title, sm.Position, sm.IsPublished))
}

// Update обновляет подмодуль.
func (r *SubmoduleRepo) Update(ctx context.Context, sm domain.Submodule) (*domain.Submodule, error) {
	q := "UPDATE " + SubmoduleTable + " SET " +
		SubmoduleColSlug + "=$1, " + SubmoduleColTitle + "=$2, " +
		SubmoduleColPosition + "=$3, " + SubmoduleColIsPublished + "=$4" +
		" WHERE " + SubmoduleColID + "=$5 RETURNING " + SubmoduleSelect("")
	return r.scanSubmodule(r.db.QueryRowContext(ctx, q, sm.Slug, sm.Title, sm.Position, sm.IsPublished, sm.ID))
}

// Delete удаляет подмодуль.
func (r *SubmoduleRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM "+SubmoduleTable+" WHERE "+SubmoduleColID+"=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SubmoduleRepo) scanSubmodule(row *sql.Row) (*domain.Submodule, error) {
	var sm domain.Submodule
	err := row.Scan(&sm.ID, &sm.ModuleID, &sm.Slug, &sm.Title, &sm.Position, &sm.IsPublished)
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
	return &sm, nil
}
