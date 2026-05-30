package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/identity/application"
	"github.com/go-mentorship-platform/backend/internal/identity/domain"
)

func TestLogoutRevokesToken(t *testing.T) {
	refresh := &mockRefresh{
		tokens: []domain.RefreshToken{
			{
				ID:        "tid",
				TokenHash: "abc",
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}

	_, hash, err := domain.HashRefreshToken("raw")
	if err != nil {
		t.Fatal(err)
	}
	refresh.tokens[0].TokenHash = hash

	svc := application.NewLogoutService(refresh)
	if err := svc.Logout(context.Background(), "raw"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if refresh.tokens[0].RevokedAt == nil {
		t.Fatal("expected revoked token")
	}
}
