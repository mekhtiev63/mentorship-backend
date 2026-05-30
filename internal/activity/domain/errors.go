package domain

import "errors"

var (
	ErrNotFound   = errors.New("activity entry not found")
	ErrForbidden  = errors.New("forbidden")
	ErrDuplicate  = errors.New("duplicate activity entry")
)
