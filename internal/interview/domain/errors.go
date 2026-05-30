package domain

import "errors"

var (
	ErrNotFound          = errors.New("interview not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrInvalidKind       = errors.New("invalid interview kind")
	ErrFeedbackRequired  = errors.New("feedback is required")
	ErrInvalidOutcome    = errors.New("invalid outcome")
	ErrCompanyRequired   = errors.New("company is required")
	ErrPositionRequired  = errors.New("position is required")
	ErrScheduledRequired = errors.New("scheduled_at is required")
)

const (
	maxCompanyLen  = 200
	maxPositionLen = 200
	maxNotesLen    = 4000
	maxFeedbackLen = 8000
)
