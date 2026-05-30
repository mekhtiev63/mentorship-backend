package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/user/application"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
)

type memUsers struct {
	byEmail map[string]domain.User
}

func (m *memUsers) Create(_ context.Context, email domain.Email, _ string, status domain.UserStatus) (domain.User, error) {
	if _, ok := m.byEmail[string(email)]; ok {
		return domain.User{}, domain.ErrEmailTaken
	}
	u := domain.User{
		ID:        domain.UserID(uuid.NewString()),
		Email:     email,
		Status:    status,
		CreatedAt: time.Now(),
	}
	m.byEmail[string(email)] = u
	return u, nil
}

func (m *memUsers) UpdateStatus(context.Context, domain.UserID, domain.UserStatus) error { return nil }
func (m *memUsers) SoftDelete(context.Context, domain.UserID) error                     { return nil }
func (m *memUsers) FindByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	for _, u := range m.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}
func (m *memUsers) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	u, ok := m.byEmail[string(email)]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}
func (m *memUsers) List(context.Context, domain.UserListFilter, pagination.Params) ([]domain.User, int64, error) {
	return nil, 0, nil
}

type memRoles struct {
	roles map[string]domain.RoleSet
}

func (m *memRoles) ListByUser(_ context.Context, id domain.UserID) (domain.RoleSet, error) {
	return m.roles[string(id)], nil
}
func (m *memRoles) ReplaceRoles(_ context.Context, id domain.UserID, roles domain.RoleSet) error {
	m.roles[string(id)] = roles
	return nil
}

type memHasher struct{}

func (memHasher) Hash(password string) (string, error) { return "hash:" + password, nil }

type memProfiles struct{}

func (memProfiles) EnsureEmpty(context.Context, domain.UserID) error { return nil }

type memSessions struct{}

func (memSessions) RevokeAllForUser(context.Context, domain.UserID) error { return nil }

type memAssign struct{}

func (memAssign) Assign(context.Context, domain.UserID, domain.UserID) (domain.BuddyAssignment, error) {
	return domain.BuddyAssignment{}, nil
}
func (memAssign) DeactivateActiveForStudent(context.Context, domain.UserID) error { return nil }
func (memAssign) FindActiveByStudent(context.Context, domain.UserID) (domain.BuddyAssignment, error) {
	return domain.BuddyAssignment{}, domain.ErrAssignmentNotFound
}
func (memAssign) ListActiveStudentsForBuddy(context.Context, domain.UserID, pagination.Params) ([]domain.StudentSummary, int64, error) {
	return nil, 0, nil
}
func (memAssign) SoftDelete(context.Context, domain.AssignmentID) error { return nil }
func (memAssign) FindByID(context.Context, domain.AssignmentID) (domain.BuddyAssignment, error) {
	return domain.BuddyAssignment{}, domain.ErrAssignmentNotFound
}
func (memAssign) IsAssignedBuddy(context.Context, domain.UserID, domain.UserID) (bool, error) {
	return false, nil
}
func (memAssign) DeactivateForBuddy(context.Context, domain.UserID) error { return nil }

func TestCreateUserDuplicateEmail(t *testing.T) {
	users := &memUsers{byEmail: map[string]domain.User{}}
	roles := &memRoles{roles: map[string]domain.RoleSet{}}
	svc := application.NewAdminService(users, roles, memHasher{}, memProfiles{}, memSessions{}, memAssign{})

	_, err := svc.CreateUser(context.Background(), application.CreateUserInput{
		Email: "a@b.com", Password: "password1", Roles: []string{"student"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateUser(context.Background(), application.CreateUserInput{
		Email: "a@b.com", Password: "password1", Roles: []string{"student"},
	})
	if err != domain.ErrEmailTaken {
		t.Fatalf("expected email taken, got %v", err)
	}
}
