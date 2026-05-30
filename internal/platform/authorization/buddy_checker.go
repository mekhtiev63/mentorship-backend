package authorization

import (
	"context"

	userdomain "github.com/go-mentorship-platform/backend/internal/user/domain"
)

// BuddyAssignmentChecker checks buddy-student assignments.
type BuddyAssignmentChecker interface {
	IsAssignedBuddy(ctx context.Context, buddyID, studentID userdomain.UserID) (bool, error)
}

// Service performs authorization checks.
type Service struct {
	buddies BuddyAssignmentChecker
}

// NewService creates an authorization service.
func NewService(buddies BuddyAssignmentChecker) *Service {
	return &Service{buddies: buddies}
}

// HasRole reports whether the user has the given role (from JWT claims in HTTP layer).
func (s *Service) HasRole(roles []Role, role Role) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAssignedBuddy reports whether buddyID is assigned to studentID.
func (s *Service) IsAssignedBuddy(ctx context.Context, buddyID, studentID string) bool {
	if s.buddies == nil {
		return false
	}
	bid, err := userdomain.ParseUserID(buddyID)
	if err != nil {
		return false
	}
	sid, err := userdomain.ParseUserID(studentID)
	if err != nil {
		return false
	}
	ok, err := s.buddies.IsAssignedBuddy(ctx, bid, sid)
	return err == nil && ok
}
