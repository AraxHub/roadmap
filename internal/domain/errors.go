package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrLocked             = errors.New("submodule locked")
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid token")
	ErrLoginTaken         = errors.New("login already taken")
	ErrConflict           = errors.New("conflict")
	ErrInvalidInput       = errors.New("invalid input")
)
