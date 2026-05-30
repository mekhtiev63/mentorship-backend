package domain_test

import (
	"testing"

	"github.com/go-mentorship-platform/backend/internal/profile/domain"
)

type stubRel struct {
	ok bool
}

func (s stubRel) IsAssignedBuddy(_, _ domain.UserID) bool { return s.ok }

func TestCanViewPublic(t *testing.T) {
	owner := domain.UserID("11111111-1111-1111-1111-111111111111")
	viewer := domain.UserID("22222222-2222-2222-2222-222222222222")
	ctx := domain.ViewerContext{ViewerID: viewer, Relationship: stubRel{false}}
	if !domain.CanView(ctx, owner, domain.VisibilityPublic) {
		t.Fatal("expected public view")
	}
}

func TestCanViewBuddiesOnly(t *testing.T) {
	owner := domain.UserID("11111111-1111-1111-1111-111111111111")
	buddy := domain.UserID("22222222-2222-2222-2222-222222222222")
	ctx := domain.ViewerContext{ViewerID: buddy, Relationship: stubRel{true}}
	if !domain.CanView(ctx, owner, domain.VisibilityBuddiesOnly) {
		t.Fatal("expected buddy view")
	}
	ctx2 := domain.ViewerContext{ViewerID: domain.UserID("33333333-3333-3333-3333-333333333333"), Relationship: stubRel{false}}
	if domain.CanView(ctx2, owner, domain.VisibilityBuddiesOnly) {
		t.Fatal("expected forbidden")
	}
}
