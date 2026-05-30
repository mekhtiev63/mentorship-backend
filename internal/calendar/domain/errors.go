package domain

import "errors"

var (
	ErrNotFound          = errors.New("calendar event not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTimeRange  = errors.New("ends_at must be after starts_at")
	ErrTitleRequired     = errors.New("title is required")
	ErrAlreadyCancelled  = errors.New("event already cancelled")
	ErrAlreadyDeleted    = errors.New("event already deleted")
)

const maxTitleLen = 200
const maxDescriptionLen = 4000
