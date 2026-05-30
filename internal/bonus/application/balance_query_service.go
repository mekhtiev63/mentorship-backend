package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// BonusBalanceQueryService reads balance and transaction history.
type BonusBalanceQueryService struct {
	accounts domain.BonusAccountRepository
	ledger   domain.BonusTransactionRepository
}

// NewBonusBalanceQueryService builds BonusBalanceQueryService.
func NewBonusBalanceQueryService(
	accounts domain.BonusAccountRepository,
	ledger domain.BonusTransactionRepository,
) *BonusBalanceQueryService {
	return &BonusBalanceQueryService{accounts: accounts, ledger: ledger}
}

// GetBalance returns balance and discount state.
func (s *BonusBalanceQueryService) GetBalance(ctx context.Context, userID string) (BalanceDTO, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return BalanceDTO{}, err
	}
	_ = s.accounts.EnsureAccount(ctx, uid)
	acc, err := s.accounts.Get(ctx, uid)
	if err != nil {
		return BalanceDTO{}, err
	}
	active, err := s.ledger.SumConvertDiscountPercent(ctx, uid)
	if err != nil {
		return BalanceDTO{}, err
	}
	state := domain.NewDiscountState(active)
	return BalanceDTO{
		Balance:                   int64(acc.Balance),
		ActiveDiscountPercent:     state.ActiveDiscountPercent,
		RemainingDiscountHeadroom: state.RemainingDiscountHeadroom,
	}, nil
}

// ListTransactions returns paginated ledger for user.
func (s *BonusBalanceQueryService) ListTransactions(ctx context.Context, userID string, page, pageSize int) ([]TransactionDTO, pagination.Meta, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	params := pagination.Normalize(page, pageSize)
	rows, total, err := s.ledger.ListByUser(ctx, uid, params)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	out := make([]TransactionDTO, len(rows))
	for i, tx := range rows {
		var ref *string
		if tx.Reference != "" {
			r := tx.Reference
			ref = &r
		}
		out[i] = TransactionDTO{
			ID:        tx.ID,
			Amount:    int64(tx.Amount),
			Type:      string(tx.Type),
			Reference: ref,
			CreatedAt: tx.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return out, pagination.NewMeta(params.Page, params.PageSize, total), nil
}
