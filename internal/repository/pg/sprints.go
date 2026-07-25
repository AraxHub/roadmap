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
	SprintTable = "sprints"

	SprintColID          = "id"
	SprintColSlug        = "slug"
	SprintColTitle       = "title"
	SprintColDescription = "description"
	SprintColPosition    = "position"
	SprintColIsPublished = "is_published"
)

// SprintColumns возвращает поля таблицы sprints для SELECT.
func SprintColumns() []string {
	return []string{
		SprintColID,
		SprintColSlug,
		SprintColTitle,
		SprintColDescription,
		SprintColPosition,
		SprintColIsPublished,
	}
}

// SprintSelect — список колонок (опционально с alias).
func SprintSelect(alias string) string {
	return selectList(alias, SprintColumns())
}

// SprintRepo — доступ к таблице sprints.
type SprintRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewSprintRepo создаёт репозиторий спринтов.
func NewSprintRepo(db *infrapg.DB, log *slog.Logger) *SprintRepo {
	return &SprintRepo{db: db, log: log}
}

// ListPublished возвращает опубликованные спринты по position ASC.
func (r *SprintRepo) ListPublished(ctx context.Context) ([]domain.Sprint, error) {
	q := "SELECT " + SprintSelect("") + " FROM " + SprintTable +
		" WHERE " + SprintColIsPublished + " ORDER BY " + SprintColPosition + " ASC"

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Sprint
	for rows.Next() {
		var s domain.Sprint
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.Description, &s.Position, &s.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListAll возвращает все спринты (для админки).
func (r *SprintRepo) ListAll(ctx context.Context) ([]domain.Sprint, error) {
	q := "SELECT " + SprintSelect("") + " FROM " + SprintTable + " ORDER BY " + SprintColPosition + " ASC"
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Sprint
	for rows.Next() {
		var s domain.Sprint
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.Description, &s.Position, &s.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// NextPosition — следующий position в конец списка.
func (r *SprintRepo) NextPosition(ctx context.Context) (int, error) {
	q := "SELECT COALESCE(MAX(" + SprintColPosition + "), 0) + 1 FROM " + SprintTable
	var n int
	err := r.db.QueryRowContext(ctx, q).Scan(&n)
	return n, err
}

// Create создаёт спринт.
func (r *SprintRepo) Create(ctx context.Context, s domain.Sprint) (*domain.Sprint, error) {
	q := "INSERT INTO " + SprintTable +
		" (" + SprintColSlug + ", " + SprintColTitle + ", " + SprintColDescription + ", " +
		SprintColPosition + ", " + SprintColIsPublished + ")" +
		" VALUES ($1, $2, $3, $4, $5) RETURNING " + SprintSelect("")
	return r.scanOne(r.db.QueryRowContext(ctx, q, s.Slug, s.Title, s.Description, s.Position, s.IsPublished))
}

// Update обновляет спринт.
func (r *SprintRepo) Update(ctx context.Context, s domain.Sprint) (*domain.Sprint, error) {
	q := "UPDATE " + SprintTable + " SET " +
		SprintColSlug + "=$1, " + SprintColTitle + "=$2, " + SprintColDescription + "=$3, " +
		SprintColPosition + "=$4, " + SprintColIsPublished + "=$5" +
		" WHERE " + SprintColID + "=$6 RETURNING " + SprintSelect("")
	out, err := r.scanOne(r.db.QueryRowContext(ctx, q, s.Slug, s.Title, s.Description, s.Position, s.IsPublished, s.ID))
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete удаляет спринт.
func (r *SprintRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM "+SprintTable+" WHERE "+SprintColID+"=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SprintRepo) scanOne(row *sql.Row) (*domain.Sprint, error) {
	var s domain.Sprint
	err := row.Scan(&s.ID, &s.Slug, &s.Title, &s.Description, &s.Position, &s.IsPublished)
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
	return &s, nil
}
