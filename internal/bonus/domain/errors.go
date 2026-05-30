package domain

import "errors"

var (
	ErrNotFound              = errors.New("bonus account not found")
	ErrInsufficientBalance   = errors.New("insufficient bonus balance")
	ErrInvalidAmount         = errors.New("invalid bonus amount")
	ErrDiscountLimit         = errors.New("discount limit exceeded")
	ErrInvalidIdempotencyKey = errors.New("invalid idempotency key")
	ErrDuplicateOperation    = errors.New("duplicate operation")
)
