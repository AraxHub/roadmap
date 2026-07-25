package domain

import "time"

// UserSprintTimer — персональный таймер спринта.
type UserSprintTimer struct {
	UserID      string
	SprintID    string
	StartedAt   time.Time
	DeadlineAt  time.Time
	CompletedAt *time.Time
}

// IsCompleted — спринт закрыт для пользователя.
func (t UserSprintTimer) IsCompleted() bool {
	return t.CompletedAt != nil
}

// IsOverdue — дедлайн прошёл, а спринт ещё не закрыт.
func (t UserSprintTimer) IsOverdue(now time.Time) bool {
	return !t.IsCompleted() && now.After(t.DeadlineAt)
}
