package domain

import "time"

// SubmoduleContent — markdown-содержимое подмодуля.
type SubmoduleContent struct {
	SubmoduleID string    `json:"submodule_id"`
	BodyMD      string    `json:"body_md"`
	UpdatedAt   time.Time `json:"updated_at"`
}
