package domain

import "time"

// UserSubmoduleProgress — факт завершения подмодуля пользователем.
type UserSubmoduleProgress struct {
	UserID      string
	SubmoduleID string
	CompletedAt time.Time
}
