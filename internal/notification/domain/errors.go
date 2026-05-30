package domain

import "errors"

var (
	ErrNotFound   = errors.New("notification: not found")
	ErrDuplicate  = errors.New("notification: duplicate")
	ErrValidation = errors.New("notification: validation")
)
