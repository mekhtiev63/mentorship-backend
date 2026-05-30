package domain

import "strings"

const maxRejectReasonLen = 2000

// RejectReason is a non-empty rejection comment.
type RejectReason string

// ParseRejectReason validates reject reason.
func ParseRejectReason(raw string) (RejectReason, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrRejectReason
	}
	if len(s) > maxRejectReasonLen {
		return "", ErrInvalidRejectReason
	}
	return RejectReason(s), nil
}
