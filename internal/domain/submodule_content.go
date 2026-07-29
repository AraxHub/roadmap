package domain

import "time"

// SubmoduleContent — блочное содержимое подмодуля.
type SubmoduleContent struct {
	SubmoduleID string         `json:"submodule_id"`
	Blocks      []ContentBlock `json:"blocks"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
