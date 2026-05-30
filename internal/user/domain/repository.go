package domain

import (
	"context"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// UserListFilter filters admin user list.
type UserListFilter struct {
	EmailPrefix string
	Status      *UserStatus
	Role        *Role
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, email Email, passwordHash string, status UserStatus) (User, error)
	UpdateStatus(ctx context.Context, id UserID, status UserStatus) error
	SoftDelete(ctx context.Context, id UserID) error
	FindByID(ctx context.Context, id UserID) (User, error)
	FindByEmail(ctx context.Context, email Email) (User, error)
	List(ctx context.Context, filter UserListFilter, page pagination.Params) ([]User, int64, error)
}

// RoleRepository manages user_roles.
type RoleRepository interface {
	ListByUser(ctx context.Context, userID UserID) (RoleSet, error)
	ReplaceRoles(ctx context.Context, userID UserID, roles RoleSet) error
}

// BuddyAssignmentRepository manages buddy assignments.
type BuddyAssignmentRepository interface {
	Assign(ctx context.Context, studentID, buddyID UserID) (BuddyAssignment, error)
	DeactivateActiveForStudent(ctx context.Context, studentID UserID) error
	FindActiveByStudent(ctx context.Context, studentID UserID) (BuddyAssignment, error)
	ListActiveStudentsForBuddy(ctx context.Context, buddyID UserID, page pagination.Params) ([]StudentSummary, int64, error)
	SoftDelete(ctx context.Context, id AssignmentID) error
	FindByID(ctx context.Context, id AssignmentID) (BuddyAssignment, error)
	IsAssignedBuddy(ctx context.Context, buddyID, studentID UserID) (bool, error)
	DeactivateForBuddy(ctx context.Context, buddyID UserID) error
}

// PasswordHasher hashes plaintext passwords (implemented by identity bcrypt).
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// ProfileBootstrap ensures a profile row exists.
type ProfileBootstrap interface {
	EnsureEmpty(ctx context.Context, userID UserID) error
}

// SessionRevoker revokes refresh tokens on user delete.
type SessionRevoker interface {
	RevokeAllForUser(ctx context.Context, userID UserID) error
}
