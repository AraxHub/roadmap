package domain

import "time"

// ContentImage — картинка урока, хранится в Postgres (BYTEA).
type ContentImage struct {
	ID          string
	SubmoduleID string
	MimeType    string
	Bytes       []byte
	ByteSize    int
	CreatedAt   time.Time
}
