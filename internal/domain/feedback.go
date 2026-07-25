package domain

import "time"

const (
	FeedbackPending  = "pending"
	FeedbackAnswered = "answered"
)

// FeedbackRequest — просьба об обратной связи (обычно через TG).
type FeedbackRequest struct {
	ID          string
	UserID      string
	RequestedAt time.Time
	TgMessageID *int64
	Status      string
	AnsweredAt  *time.Time
	AnswerText  string
}
