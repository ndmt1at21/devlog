package domain

import "errors"

// Sentinel errors shared across repository implementations. Mirrors the IAM
// service's error mapping (sql.ErrNoRows -> ErrNotFound, duplicate -> ErrConflict).
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)
