package domain

import "errors"

var (
	ErrNotFound          = errors.New("achievement not found")
	ErrDefinitionInactive = errors.New("achievement definition inactive")
	ErrInvalidRule       = errors.New("invalid achievement rule")
	ErrForbidden         = errors.New("forbidden")
)
