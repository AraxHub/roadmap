package pg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	infrapg "roadmap/internal/infrastructure/pg"
)

const (
	FeedbackScheduleTable = "feedback_schedule"

	FSColID          = "id"
	FSColLastRoundAt = "last_round_at"
	FSColMessageText = "message_text"

	DefaultFeedbackMessage = "Привет! Оставь, пожалуйста, обратную связь по обучению — ответь на это сообщение."
)

// FeedbackScheduleRepo — персистентный слот авто-рассылки ОС и текст просьбы.
type FeedbackScheduleRepo struct {
	db  *infrapg.DB
	log *slog.Logger
}

// NewFeedbackScheduleRepo создаёт репозиторий.
func NewFeedbackScheduleRepo(db *infrapg.DB, log *slog.Logger) *FeedbackScheduleRepo {
	return &FeedbackScheduleRepo{db: db, log: log}
}

// GetLastRoundAt — время последнего авто-раунда; nil, если ещё не было.
func (r *FeedbackScheduleRepo) GetLastRoundAt(ctx context.Context) (*time.Time, error) {
	var t sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT `+FSColLastRoundAt+` FROM `+FeedbackScheduleTable+` WHERE `+FSColID+` = 1`,
	).Scan(&t)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, nil
	}
	v := t.Time.UTC()
	return &v, nil
}

// MarkRoundAt фиксирует, что авто-раунд за слот at выполнен.
func (r *FeedbackScheduleRepo) MarkRoundAt(ctx context.Context, at time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO `+FeedbackScheduleTable+` (`+FSColID+`, `+FSColLastRoundAt+`, `+FSColMessageText+`)
		 VALUES (1, $1, $2)
		 ON CONFLICT (`+FSColID+`) DO UPDATE SET `+FSColLastRoundAt+` = EXCLUDED.`+FSColLastRoundAt,
		at.UTC(), DefaultFeedbackMessage,
	)
	return err
}

// GetMessageText — текст просьбы об ОС из БД.
func (r *FeedbackScheduleRepo) GetMessageText(ctx context.Context) (string, error) {
	var text string
	err := r.db.QueryRowContext(ctx,
		`SELECT `+FSColMessageText+` FROM `+FeedbackScheduleTable+` WHERE `+FSColID+` = 1`,
	).Scan(&text)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultFeedbackMessage, nil
	}
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return DefaultFeedbackMessage, nil
	}
	return text, nil
}

// SetMessageText сохраняет текст просьбы об ОС.
func (r *FeedbackScheduleRepo) SetMessageText(ctx context.Context, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		text = DefaultFeedbackMessage
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO `+FeedbackScheduleTable+` (`+FSColID+`, `+FSColMessageText+`) VALUES (1, $1)
		 ON CONFLICT (`+FSColID+`) DO UPDATE SET `+FSColMessageText+` = EXCLUDED.`+FSColMessageText,
		text,
	)
	return err
}
