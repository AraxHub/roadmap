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
	FeedbackRequestTable = "feedback_requests"

	FRColID          = "id"
	FRColUserID      = "user_id"
	FRColRequestedAt = "requested_at"
	FRColTgMessageID = "tg_message_id"
	FRColStatus      = "status"
	FRColAnsweredAt  = "answered_at"
	FRColAnswerText  = "answer_text"
)

// FeedbackRequestColumns — поля SELECT.
func FeedbackRequestColumns() []string {
	return []string{
		FRColID,
		FRColUserID,
		FRColRequestedAt,
		FRColTgMessageID,
		FRColStatus,
		FRColAnsweredAt,
		FRColAnswerText,
	}
}

// FeedbackRequestSelect — список колонок.
func FeedbackRequestSelect(alias string) string {
	return selectList(alias, FeedbackRequestColumns())
}

// FeedbackRequestRepo — просьбы об ОС.
type FeedbackRequestRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewFeedbackRequestRepo создаёт репозиторий.
func NewFeedbackRequestRepo(db *infrapg.DB, log *slog.Logger) *FeedbackRequestRepo {
	return &FeedbackRequestRepo{db: db, log: log}
}

// HasPending — есть ли незакрытая просьба.
func (r *FeedbackRequestRepo) HasPending(ctx context.Context, userID string) (bool, error) {
	q := "SELECT 1 FROM " + FeedbackRequestTable +
		" WHERE " + FRColUserID + " = $1 AND " + FRColStatus + " = $2 LIMIT 1"
	var one int
	err := r.db.QueryRowContext(ctx, q, userID, domain.FeedbackPending).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Create создаёт pending-просьбу.
func (r *FeedbackRequestRepo) Create(ctx context.Context, userID string, tgMessageID *int64) (*domain.FeedbackRequest, error) {
	q := "INSERT INTO " + FeedbackRequestTable +
		" (" + FRColUserID + ", " + FRColRequestedAt + ", " + FRColTgMessageID + ", " + FRColStatus + ")" +
		" VALUES ($1, $2, $3, $4) RETURNING " + FeedbackRequestSelect("")
	return scanFeedback(r.db.QueryRowContext(ctx, q, userID, time.Now().UTC(), tgMessageID, domain.FeedbackPending))
}

// AnswerPending сохраняет ответ и закрывает последнюю pending-просьбу.
func (r *FeedbackRequestRepo) AnswerPending(ctx context.Context, userID, text string) (*domain.FeedbackRequest, error) {
	now := time.Now().UTC()
	q := "UPDATE " + FeedbackRequestTable +
		" SET " + FRColStatus + " = $3, " + FRColAnsweredAt + " = $4, " + FRColAnswerText + " = $5" +
		" WHERE " + FRColID + " = (" +
		" SELECT " + FRColID + " FROM " + FeedbackRequestTable +
		" WHERE " + FRColUserID + " = $1 AND " + FRColStatus + " = $2" +
		" ORDER BY " + FRColRequestedAt + " DESC LIMIT 1" +
		") RETURNING " + FeedbackRequestSelect("")
	fr, err := scanFeedback(r.db.QueryRowContext(ctx, q, userID, domain.FeedbackPending, domain.FeedbackAnswered, now, text))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return fr, err
}

// ListByUser — история ОС пользователя.
func (r *FeedbackRequestRepo) ListByUser(ctx context.Context, userID string) ([]domain.FeedbackRequest, error) {
	q := "SELECT " + FeedbackRequestSelect("") + " FROM " + FeedbackRequestTable +
		" WHERE " + FRColUserID + " = $1 ORDER BY " + FRColRequestedAt + " DESC"
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.FeedbackRequest, 0)
	for rows.Next() {
		fr, err := scanFeedback(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *fr)
	}
	return out, rows.Err()
}

type feedbackScanner interface {
	Scan(dest ...any) error
}

func scanFeedback(s feedbackScanner) (*domain.FeedbackRequest, error) {
	var fr domain.FeedbackRequest
	var msgID sql.NullInt64
	var answered sql.NullTime
	err := s.Scan(&fr.ID, &fr.UserID, &fr.RequestedAt, &msgID, &fr.Status, &answered, &fr.AnswerText)
	if err != nil {
		return nil, err
	}
	if msgID.Valid {
		v := msgID.Int64
		fr.TgMessageID = &v
	}
	if answered.Valid {
		v := answered.Time.UTC()
		fr.AnsweredAt = &v
	}
	fr.RequestedAt = fr.RequestedAt.UTC()
	return &fr, nil
}
