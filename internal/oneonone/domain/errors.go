package domain

import "errors"

var (
	ErrNotFound            = errors.New("one on one request not found")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrInsufficientBonus   = errors.New("insufficient bonus balance")
	ErrRejectReason        = errors.New("reject reason is required")
	ErrInvalidMessage      = errors.New("invalid message")
)

// OneOnOneCostPoints is bonus cost for 1:1 session.
const OneOnOneCostPoints int64 = 1000
