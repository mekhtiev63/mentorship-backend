package domain

import "github.com/google/uuid"

// RequestID identifies a 1:1 request.
type RequestID string

// StudentID identifies student.
type StudentID string

// BuddyID identifies buddy.
type BuddyID string

// UserID identifies admin or user.
type UserID string

// ParseRequestID parses request UUID.
func ParseRequestID(raw string) (RequestID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return RequestID(raw), nil
}

// ParseStudentID parses student UUID.
func ParseStudentID(raw string) (StudentID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return StudentID(raw), nil
}

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return UserID(raw), nil
}

// BonusReference builds ledger reference.
func BonusReference(requestID RequestID) string {
	return "one_on_one:" + string(requestID)
}
