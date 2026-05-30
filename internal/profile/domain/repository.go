package domain

import "context"

// ProfileRepository persists profiles.
type ProfileRepository interface {
	EnsureEmpty(ctx context.Context, userID UserID) error
	GetByUserID(ctx context.Context, userID UserID) (Profile, error)
	Update(ctx context.Context, profile Profile) error
}

// UserExistsReader checks target user is viewable (exists, not deleted).
type UserExistsReader interface {
	ExistsActive(ctx context.Context, userID UserID) (bool, error)
}

// BuddyAssignmentReader checks buddy relationship.
type BuddyAssignmentReader interface {
	IsAssignedBuddy(ctx context.Context, buddyID, studentID UserID) (bool, error)
}
