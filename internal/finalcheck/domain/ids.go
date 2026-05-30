package domain

import "github.com/google/uuid"

// AssessmentID identifies final assessment.
type AssessmentID string

// StudentID identifies student.
type StudentID string

// UserID identifies reviewer or actor.
type UserID string

// ParseAssessmentID parses UUID.
func ParseAssessmentID(raw string) (AssessmentID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return AssessmentID(raw), nil
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
