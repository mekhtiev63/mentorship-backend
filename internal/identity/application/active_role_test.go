package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/application"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

func TestSetActiveRoleNotAllowed(t *testing.T) {
	userID, _ := domain.ParseUserID("11111111-1111-1111-1111-111111111111")
	email, _ := domain.ParseEmail("student@example.com")
	creds := domain.Credentials{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hash:x",
		Account:      domain.AccountState{Status: domain.AccountStatusActive},
	}

	svc := application.NewActiveRoleService(
		&mockCredentials{byEmail: map[string]domain.Credentials{email.String(): creds}},
		&mockRoles{roles: map[string]domain.RoleSet{userID.String(): {domain.RoleStudent}}},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Now()},
		time.Minute,
		time.Hour,
	)

	_, _, err := svc.SetActiveRole(context.Background(), userID, "admin")
	if err != domain.ErrActiveRoleNotAllowed {
		t.Fatalf("expected forbidden role, got %v", err)
	}
}

func TestSetActiveRoleSuccess(t *testing.T) {
	userID, _ := domain.ParseUserID("11111111-1111-1111-1111-111111111111")
	email, _ := domain.ParseEmail("multi@example.com")
	creds := domain.Credentials{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hash:x",
		Account:      domain.AccountState{Status: domain.AccountStatusActive},
	}

	svc := application.NewActiveRoleService(
		&mockCredentials{byEmail: map[string]domain.Credentials{email.String(): creds}},
		&mockRoles{roles: map[string]domain.RoleSet{
			userID.String(): {domain.RoleStudent, domain.RoleBuddy},
		}},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Now()},
		time.Minute,
		time.Hour,
	)

	_, user, err := svc.SetActiveRole(context.Background(), userID, "buddy")
	if err != nil {
		t.Fatalf("set active role: %v", err)
	}
	if user.ActiveRole == nil || *user.ActiveRole != "buddy" {
		t.Fatalf("expected buddy, got %+v", user.ActiveRole)
	}
}
