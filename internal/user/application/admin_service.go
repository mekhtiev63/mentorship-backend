package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// AdminService handles admin user operations.
type AdminService struct {
	users     domain.UserRepository
	roles     domain.RoleRepository
	hasher    domain.PasswordHasher
	profiles  domain.ProfileBootstrap
	sessions  domain.SessionRevoker
	assignments domain.BuddyAssignmentRepository
}

// NewAdminService builds AdminService.
func NewAdminService(
	users domain.UserRepository,
	roles domain.RoleRepository,
	hasher domain.PasswordHasher,
	profiles domain.ProfileBootstrap,
	sessions domain.SessionRevoker,
	assignments domain.BuddyAssignmentRepository,
) *AdminService {
	return &AdminService{
		users:       users,
		roles:       roles,
		hasher:      hasher,
		profiles:    profiles,
		sessions:    sessions,
		assignments: assignments,
	}
}

// CreateUserInput is input for user creation.
type CreateUserInput struct {
	Email    string
	Password string
	Roles    []string
}

// UserDTO is returned to HTTP layer.
type UserDTO struct {
	ID        string
	Email     string
	Status    string
	Roles     []string
	CreatedAt string
}

// CreateUser registers a new user with roles.
func (s *AdminService) CreateUser(ctx context.Context, in CreateUserInput) (UserDTO, error) {
	email, err := domain.ParseEmail(in.Email)
	if err != nil {
		return UserDTO{}, err
	}
	if len(in.Password) < 8 {
		return UserDTO{}, domain.ErrWeakPassword
	}
	roleSet, err := parseRoles(in.Roles)
	if err != nil {
		return UserDTO{}, err
	}
	if roleSet.Len() == 0 {
		roleSet = domain.RoleSet{domain.RoleStudent}
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return UserDTO{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, hash, domain.UserStatusActive)
	if err != nil {
		return UserDTO{}, err
	}
	if err := s.roles.ReplaceRoles(ctx, user.ID, roleSet); err != nil {
		return UserDTO{}, err
	}
	if err := s.profiles.EnsureEmpty(ctx, user.ID); err != nil {
		return UserDTO{}, err
	}
	return s.toDTO(ctx, user)
}

// GetUser returns user by id.
func (s *AdminService) GetUser(ctx context.Context, id string) (UserDTO, error) {
	userID, err := domain.ParseUserID(id)
	if err != nil {
		return UserDTO{}, domain.ErrNotFound
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return UserDTO{}, err
	}
	return s.toDTO(ctx, user)
}

// ListUsers lists users with pagination.
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int, emailPrefix string) ([]UserDTO, pagination.Meta, error) {
	params := pagination.Normalize(page, pageSize)
	users, total, err := s.users.List(ctx, domain.UserListFilter{EmailPrefix: emailPrefix}, params)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		dto, err := s.toDTO(ctx, u)
		if err != nil {
			return nil, pagination.Meta{}, err
		}
		out = append(out, dto)
	}
	return out, pagination.NewMeta(params.Page, params.PageSize, total), nil
}

// UpdateStatus changes account status.
func (s *AdminService) UpdateStatus(ctx context.Context, id, status string) (UserDTO, error) {
	userID, err := domain.ParseUserID(id)
	if err != nil {
		return UserDTO{}, domain.ErrNotFound
	}
	st, err := domain.ParseUserStatus(status)
	if err != nil {
		return UserDTO{}, err
	}
	if err := s.users.UpdateStatus(ctx, userID, st); err != nil {
		return UserDTO{}, err
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return UserDTO{}, err
	}
	return s.toDTO(ctx, user)
}

// ReplaceRoles replaces all roles for a user.
func (s *AdminService) ReplaceRoles(ctx context.Context, id string, roles []string) (UserDTO, error) {
	userID, err := domain.ParseUserID(id)
	if err != nil {
		return UserDTO{}, domain.ErrNotFound
	}
	roleSet, err := parseRoles(roles)
	if err != nil {
		return UserDTO{}, err
	}
	if err := s.roles.ReplaceRoles(ctx, userID, roleSet); err != nil {
		return UserDTO{}, err
	}
	if !roleSet.Contains(domain.RoleBuddy) {
		_ = s.assignments.DeactivateForBuddy(ctx, userID)
	}
	if !roleSet.Contains(domain.RoleStudent) {
		_ = s.assignments.DeactivateActiveForStudent(ctx, userID)
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return UserDTO{}, err
	}
	return s.toDTO(ctx, user)
}

// SoftDeleteUser soft-deletes a user.
func (s *AdminService) SoftDeleteUser(ctx context.Context, id string) error {
	userID, err := domain.ParseUserID(id)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := s.users.SoftDelete(ctx, userID); err != nil {
		return err
	}
	_ = s.sessions.RevokeAllForUser(ctx, userID)
	return nil
}

// CreateBuddyAssignment assigns buddy to student.
func (s *AdminService) CreateBuddyAssignment(ctx context.Context, studentID, buddyID string) error {
	stID, err := domain.ParseUserID(studentID)
	if err != nil {
		return domain.ErrInvalidAssignment
	}
	buID, err := domain.ParseUserID(buddyID)
	if err != nil {
		return domain.ErrInvalidAssignment
	}
	if stID == buID {
		return domain.ErrInvalidAssignment
	}

	studentRoles, err := s.roles.ListByUser(ctx, stID)
	if err != nil {
		return err
	}
	buddyRoles, err := s.roles.ListByUser(ctx, buID)
	if err != nil {
		return err
	}
	if !studentRoles.Contains(domain.RoleStudent) {
		return domain.ErrStudentRoleRequired
	}
	if !buddyRoles.Contains(domain.RoleBuddy) {
		return domain.ErrBuddyRoleRequired
	}

	_, err = s.assignments.Assign(ctx, stID, buID)
	return err
}

// DeleteBuddyAssignment soft-deletes assignment.
func (s *AdminService) DeleteBuddyAssignment(ctx context.Context, assignmentID string) error {
	id, err := domain.ParseAssignmentID(assignmentID)
	if err != nil {
		return err
	}
	return s.assignments.SoftDelete(ctx, id)
}

func (s *AdminService) toDTO(ctx context.Context, user domain.User) (UserDTO, error) {
	roles, err := s.roles.ListByUser(ctx, user.ID)
	if err != nil {
		return UserDTO{}, err
	}
	return UserDTO{
		ID:        string(user.ID),
		Email:     string(user.Email),
		Status:    string(user.Status),
		Roles:     roles.Strings(),
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func parseRoles(raw []string) (domain.RoleSet, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	set := make(domain.RoleSet, 0, len(raw))
	for _, r := range raw {
		role, err := domain.ParseRole(r)
		if err != nil {
			return nil, err
		}
		set = append(set, role)
	}
	return set, nil
}
