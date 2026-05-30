package bonus

import (
	"context"
	"errors"

	bonusapp "github.com/go-mentorship-platform/backend/internal/bonus/application"
	bonusdomain "github.com/go-mentorship-platform/backend/internal/bonus/domain"
	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
)

// LedgerAdapter implements BonusOneOnOnePort via BonusLedgerService.
type LedgerAdapter struct {
	ledger *bonusapp.BonusLedgerService
}

// NewLedgerAdapter creates LedgerAdapter.
func NewLedgerAdapter(ledger *bonusapp.BonusLedgerService) *LedgerAdapter {
	return &LedgerAdapter{ledger: ledger}
}

// HasSufficientBalance checks bonus balance.
func (a *LedgerAdapter) HasSufficientBalance(ctx context.Context, studentID domain.StudentID, amount int64) (bool, error) {
	bal, err := a.ledger.GetBalance(ctx, string(studentID))
	if err != nil {
		return false, err
	}
	return bal >= amount, nil
}

// DebitForRequest debits bonus for approved request (idempotent).
func (a *LedgerAdapter) DebitForRequest(ctx context.Context, studentID domain.StudentID, requestID domain.RequestID, amount int64) error {
	err := a.ledger.DebitOneOnOne(ctx, string(studentID), string(requestID), amount)
	if errors.Is(err, bonusdomain.ErrInsufficientBalance) {
		return domain.ErrInsufficientBonus
	}
	return err
}
