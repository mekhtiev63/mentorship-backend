package domain

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentID identifies a buddy assignment row.
type AssignmentID string

// ParseAssignmentID validates assignment UUID.
func ParseAssignmentID(raw string) (AssignmentID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrAssignmentNotFound
	}
	return AssignmentID(raw), nil
}

// BuddyAssignment is the buddy assignment aggregate root.
type BuddyAssignment struct {
	ID        AssignmentID
	StudentID UserID
	BuddyID   UserID
	Active    bool
	ValidFrom time.Time
	ValidTo   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// StudentSummary is a minimal student row for buddy lists.
type StudentSummary struct {
	ID    UserID
	Email Email
}
