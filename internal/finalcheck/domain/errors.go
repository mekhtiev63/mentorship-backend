package domain

import "errors"

var (
	ErrNotFound          = errors.New("final assessment not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrProgramIncomplete = errors.New("roadmap program not completed")
	ErrFeedbackRequired  = errors.New("feedback is required")
	ErrReasonRequired    = errors.New("reason is required")
	ErrScheduledRequired = errors.New("scheduled_at is required")
	ErrRetryNotAllowed   = errors.New("retry not allowed")
)

const maxFeedbackLen = 8000
const maxReasonLen = 2000
