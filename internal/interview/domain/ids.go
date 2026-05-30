package domain

import "github.com/google/uuid"

// InterviewID identifies an interview.
type InterviewID string

// StudentID identifies a student.
type StudentID string

// UserID identifies a user.
type UserID string

// BuddyID identifies a buddy (interviewer for mock).
type BuddyID string

// ParseInterviewID parses interview UUID.
func ParseInterviewID(raw string) (InterviewID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return InterviewID(raw), nil
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
