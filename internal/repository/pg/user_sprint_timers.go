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
	UserSprintTimerTable = "user_sprint_timers"

	USTColUserID      = "user_id"
	USTColSprintID    = "sprint_id"
	USTColStartedAt   = "started_at"
	USTColDeadlineAt  = "deadline_at"
	USTColCompletedAt = "completed_at"
)

// UserSprintTimerColumns — поля SELECT.
func UserSprintTimerColumns() []string {
	return []string{
		USTColUserID,
		USTColSprintID,
		USTColStartedAt,
		USTColDeadlineAt,
		USTColCompletedAt,
	}
}

// UserSprintTimerSelect — список колонок.
func UserSprintTimerSelect(alias string) string {
	return selectList(alias, UserSprintTimerColumns())
}

// UserSprintTimerRepo — персональные таймеры спринтов.
type UserSprintTimerRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewUserSprintTimerRepo создаёт репозиторий.
func NewUserSprintTimerRepo(db *infrapg.DB, log *slog.Logger) *UserSprintTimerRepo {
	return &UserSprintTimerRepo{db: db, log: log}
}

// ListByUser возвращает все таймеры пользователя.
func (r *UserSprintTimerRepo) ListByUser(ctx context.Context, userID string) ([]domain.UserSprintTimer, error) {
	q := "SELECT " + UserSprintTimerSelect("") + " FROM " + UserSprintTimerTable +
		" WHERE " + USTColUserID + " = $1"
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.UserSprintTimer, 0)
	for rows.Next() {
		t, err := scanTimer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

// Get возвращает таймер пары user/sprint.
func (r *UserSprintTimerRepo) Get(ctx context.Context, userID, sprintID string) (*domain.UserSprintTimer, error) {
	q := "SELECT " + UserSprintTimerSelect("") + " FROM " + UserSprintTimerTable +
		" WHERE " + USTColUserID + " = $1 AND " + USTColSprintID + " = $2"
	t, err := scanTimer(r.db.QueryRowContext(ctx, q, userID, sprintID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

// Start создаёт таймер, если его ещё нет (идемпотентно).
func (r *UserSprintTimerRepo) Start(ctx context.Context, userID, sprintID string, startedAt, deadlineAt time.Time) error {
	q := "INSERT INTO " + UserSprintTimerTable +
		" (" + USTColUserID + ", " + USTColSprintID + ", " + USTColStartedAt + ", " + USTColDeadlineAt + ")" +
		" VALUES ($1, $2, $3, $4)" +
		" ON CONFLICT (" + USTColUserID + ", " + USTColSprintID + ") DO NOTHING"
	_, err := r.db.ExecContext(ctx, q, userID, sprintID, startedAt, deadlineAt)
	return err
}

// Complete фиксирует завершение спринта (если ещё не закрыт).
func (r *UserSprintTimerRepo) Complete(ctx context.Context, userID, sprintID string, at time.Time) error {
	q := "UPDATE " + UserSprintTimerTable +
		" SET " + USTColCompletedAt + " = $3" +
		" WHERE " + USTColUserID + " = $1 AND " + USTColSprintID + " = $2 AND " + USTColCompletedAt + " IS NULL"
	_, err := r.db.ExecContext(ctx, q, userID, sprintID, at)
	return err
}

type timerScanner interface {
	Scan(dest ...any) error
}

func scanTimer(s timerScanner) (*domain.UserSprintTimer, error) {
	var t domain.UserSprintTimer
	var completed sql.NullTime
	err := s.Scan(&t.UserID, &t.SprintID, &t.StartedAt, &t.DeadlineAt, &completed)
	if err != nil {
		return nil, err
	}
	if completed.Valid {
		v := completed.Time.UTC()
		t.CompletedAt = &v
	}
	t.StartedAt = t.StartedAt.UTC()
	t.DeadlineAt = t.DeadlineAt.UTC()
	return &t, nil
}
