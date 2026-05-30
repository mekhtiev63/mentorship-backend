package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/application"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

type mockCredentials struct {
	byEmail map[string]domain.Credentials
}

func (m *mockCredentials) FindByEmail(_ context.Context, email domain.Email) (domain.Credentials, error) {
	c, ok := m.byEmail[email.String()]
	if !ok {
		return domain.Credentials{}, domain.ErrInvalidCredentials
	}
	return c, nil
}

func (m *mockCredentials) FindByID(_ context.Context, id domain.UserID) (domain.Credentials, error) {
	for _, c := range m.byEmail {
		if c.UserID == id {
			return c, nil
		}
	}
	return domain.Credentials{}, domain.ErrAccountInactive
}

type mockRoles struct {
	roles map[string]domain.RoleSet
}

func (m *mockRoles) ListByUser(_ context.Context, id domain.UserID) (domain.RoleSet, error) {
	return m.roles[id.String()], nil
}

type mockPreferences struct {
	store map[string]*domain.Role
}

func (m *mockPreferences) Get(_ context.Context, id domain.UserID) (domain.UserPreferences, error) {
	role := m.store[id.String()]
	return domain.UserPreferences{UserID: id, ActiveRole: role}, nil
}

func (m *mockPreferences) SetActiveRole(_ context.Context, id domain.UserID, role *domain.Role) error {
	if role == nil {
		delete(m.store, id.String())
		return nil
	}
	v := *role
	m.store[id.String()] = &v
	return nil
}

type mockRefresh struct {
	tokens []domain.RefreshToken
}

func (m *mockRefresh) Store(_ context.Context, t domain.RefreshToken) error {
	m.tokens = append(m.tokens, t)
	return nil
}

func (m *mockRefresh) FindByHash(_ context.Context, hash string) (domain.RefreshToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return domain.RefreshToken{}, domain.ErrInvalidToken
}

func (m *mockRefresh) Revoke(_ context.Context, id domain.RefreshTokenID) error {
	for i, t := range m.tokens {
		if t.ID == id {
			now := time.Now()
			m.tokens[i].RevokedAt = &now
		}
	}
	return nil
}

type mockHasher struct{}

func (mockHasher) Hash(password string) (domain.PasswordHash, error) {
	return domain.PasswordHash("hash:" + password), nil
}

func (mockHasher) Verify(hash domain.PasswordHash, password string) error {
	if string(hash) != "hash:"+password {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type mockIssuer struct{}

func (mockIssuer) Issue(claims domain.AccessTokenClaims) (string, error) {
	return "access-token", nil
}

type mockRefreshGen struct{}

func (mockRefreshGen) Generate() (string, string, error) {
	return "raw-refresh", "hash-refresh", nil
}

func TestLoginSuccessSingleRole(t *testing.T) {
	userID, _ := domain.ParseUserID("11111111-1111-1111-1111-111111111111")
	email, _ := domain.ParseEmail("student@example.com")

	creds := domain.Credentials{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hash:secret",
		Account:      domain.AccountState{Status: domain.AccountStatusActive},
	}

	svc := application.NewLoginService(
		&mockCredentials{byEmail: map[string]domain.Credentials{email.String(): creds}},
		&mockRoles{roles: map[string]domain.RoleSet{userID.String(): {domain.RoleStudent}}},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockHasher{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Unix(1_700_000_000, 0)},
		15*time.Minute,
		24*time.Hour,
	)

	pair, user, err := svc.Login(context.Background(), "student@example.com", "secret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected tokens")
	}
	if user.ActiveRole == nil || *user.ActiveRole != "student" {
		t.Fatalf("expected active student role, got %+v", user.ActiveRole)
	}
	if pair.RequiresRoleSelection {
		t.Fatal("did not expect role selection")
	}
}

func TestLoginDeletedUser(t *testing.T) {
	_, _, err := application.NewLoginService(
		&mockCredentials{byEmail: map[string]domain.Credentials{}},
		&mockRoles{},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockHasher{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Now()},
		time.Minute,
		time.Hour,
	).Login(context.Background(), "gone@example.com", "secret")

	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLoginDisabledUser(t *testing.T) {
	userID, _ := domain.ParseUserID("11111111-1111-1111-1111-111111111111")
	email, _ := domain.ParseEmail("disabled@example.com")
	creds := domain.Credentials{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hash:secret",
		Account:      domain.AccountState{Status: domain.AccountStatusDisabled},
	}

	_, _, err := application.NewLoginService(
		&mockCredentials{byEmail: map[string]domain.Credentials{email.String(): creds}},
		&mockRoles{roles: map[string]domain.RoleSet{userID.String(): {domain.RoleStudent}}},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockHasher{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Now()},
		time.Minute,
		time.Hour,
	).Login(context.Background(), email.String(), "secret")

	if err != domain.ErrAccountInactive {
		t.Fatalf("expected account inactive, got %v", err)
	}
}

func TestLoginRequiresRoleSelection(t *testing.T) {
	userID, _ := domain.ParseUserID("11111111-1111-1111-1111-111111111111")
	email, _ := domain.ParseEmail("multi@example.com")
	creds := domain.Credentials{
		UserID:       userID,
		Email:        email,
		PasswordHash: "hash:secret",
		Account:      domain.AccountState{Status: domain.AccountStatusActive},
	}

	pair, _, err := application.NewLoginService(
		&mockCredentials{byEmail: map[string]domain.Credentials{email.String(): creds}},
		&mockRoles{roles: map[string]domain.RoleSet{
			userID.String(): {domain.RoleStudent, domain.RoleBuddy},
		}},
		&mockPreferences{store: map[string]*domain.Role{}},
		&mockRefresh{},
		mockHasher{},
		mockIssuer{},
		mockRefreshGen{},
		fixedClock{now: time.Now()},
		time.Minute,
		time.Hour,
	).Login(context.Background(), email.String(), "secret")

	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if !pair.RequiresRoleSelection {
		t.Fatal("expected role selection")
	}
}
