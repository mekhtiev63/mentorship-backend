package domain_test

import (
	"testing"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
)

func TestApplyConvertHeadroom(t *testing.T) {
	if err := domain.ApplyConvertHeadroom(12, 3); err != nil {
		t.Fatal(err)
	}
	if err := domain.ApplyConvertHeadroom(12, 4); err != domain.ErrDiscountLimit {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestDiscountPercentFromPoints(t *testing.T) {
	if got := domain.DiscountPercentFromPoints(250); got != 2 {
		t.Fatalf("got %d", got)
	}
}
