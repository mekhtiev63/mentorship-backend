package application

import (
	"context"
	"strings"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
)

// DiscountConversionService converts bonus points to discount.
type DiscountConversionService struct {
	ledger   *BonusLedgerService
	accounts domain.BonusAccountRepository
	ledgerRO domain.BonusTransactionRepository
}

// NewDiscountConversionService builds DiscountConversionService.
func NewDiscountConversionService(
	ledger *BonusLedgerService,
	accounts domain.BonusAccountRepository,
	ledgerRO domain.BonusTransactionRepository,
) *DiscountConversionService {
	return &DiscountConversionService{ledger: ledger, accounts: accounts, ledgerRO: ledgerRO}
}

// Convert spends points for discount (idempotent).
func (s *DiscountConversionService) Convert(ctx context.Context, userID string, points int64, idempotencyKey string) (ConvertResultDTO, error) {
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return ConvertResultDTO{}, domain.ErrInvalidIdempotencyKey
	}
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	amount, err := domain.ParseBonusAmount(points)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	if existing, found, err := s.ledgerRO.FindByIdempotencyKey(ctx, key); err != nil {
		return ConvertResultDTO{}, err
	} else if found {
		return s.resultFromExisting(ctx, uid, existing)
	}

	added := domain.DiscountPercentFromPoints(amount)
	active, err := s.ledgerRO.SumConvertDiscountPercent(ctx, uid)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	if err := domain.ApplyConvertHeadroom(active, added); err != nil {
		return ConvertResultDTO{}, err
	}

	ref := MarshalConvertReference(added, int64(amount), key)
	tx, err := s.ledger.AppendConvert(ctx, uid, amount, ref, key)
	if err != nil {
		if err == domain.ErrDuplicateOperation {
			existing, found, findErr := s.ledgerRO.FindByIdempotencyKey(ctx, key)
			if findErr != nil || !found {
				return ConvertResultDTO{}, err
			}
			return s.resultFromExisting(ctx, uid, existing)
		}
		return ConvertResultDTO{}, err
	}

	acc, err := s.accounts.Get(ctx, uid)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	newActive := active + added
	return ConvertResultDTO{
		TransactionID:         tx.ID,
		PointsConverted:       int64(amount),
		DiscountPercentAdded:  added,
		ActiveDiscountPercent: newActive,
		BalanceAfter:          int64(acc.Balance),
	}, nil
}

func (s *DiscountConversionService) resultFromExisting(ctx context.Context, uid domain.UserID, tx domain.BonusTransaction) (ConvertResultDTO, error) {
	active, err := s.ledgerRO.SumConvertDiscountPercent(ctx, uid)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	acc, err := s.accounts.Get(ctx, uid)
	if err != nil {
		return ConvertResultDTO{}, err
	}
	added := domain.DiscountPercentFromPoints(tx.Amount)
	return ConvertResultDTO{
		TransactionID:         tx.ID,
		PointsConverted:       int64(tx.Amount),
		DiscountPercentAdded:  added,
		ActiveDiscountPercent: active,
		BalanceAfter:          int64(acc.Balance),
	}, nil
}
