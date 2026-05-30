package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
)

// BonusLedgerService appends ledger entries transactionally.
type BonusLedgerService struct {
	tx       domain.Transactor
	accounts domain.BonusAccountRepository
	ledger   domain.BonusTransactionRepository
	events   domain.EventRecorder
}

// NewBonusLedgerService builds BonusLedgerService.
func NewBonusLedgerService(
	tx domain.Transactor,
	accounts domain.BonusAccountRepository,
	ledger domain.BonusTransactionRepository,
	events domain.EventRecorder,
) *BonusLedgerService {
	return &BonusLedgerService{tx: tx, accounts: accounts, ledger: ledger, events: events}
}

// CreditAchievement grants bonus for achievement (idempotent by reference).
func (s *BonusLedgerService) CreditAchievement(ctx context.Context, userID, achievementCode, sourceEventID string) (bool, error) {
	var created bool
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		created, err = s.creditAchievementInTx(ctx, userID, achievementCode, sourceEventID)
		return err
	})
	return created, err
}

func (s *BonusLedgerService) creditAchievementInTx(ctx context.Context, userID, achievementCode, sourceEventID string) (bool, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return false, err
	}
	amount, ok := domain.CreditForAchievement(domain.AchievementCode(achievementCode))
	if !ok || amount == 0 {
		return false, nil
	}
	ref := domain.AchievementCreditReference(domain.AchievementCode(achievementCode), sourceEventID)
	if _, found, err := s.ledger.FindByReference(ctx, uid, ref); err != nil {
		return false, err
	} else if found {
		return false, nil
	}
	entry, err := s.appendEntry(ctx, uid, domain.TransactionCredit, amount, ref, nil)
	if err != nil {
		if err == domain.ErrDuplicateOperation {
			return false, nil
		}
		return false, err
	}
	_ = s.events.Record(ctx, domain.EventCredited, map[string]any{
		"userId": userID, "amount": int64(amount), "reference": ref, "transactionId": entry.ID,
	})
	return true, nil
}

// AppendConvert debits points for conversion.
func (s *BonusLedgerService) AppendConvert(ctx context.Context, userID domain.UserID, points domain.BonusAmount, reference string, idempotencyKey string) (domain.BonusTransaction, error) {
	var out domain.BonusTransaction
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		key := idempotencyKey
		tx, err := s.appendEntry(ctx, userID, domain.TransactionConvert, points, reference, &key)
		if err != nil {
			return err
		}
		out = tx
		_ = s.events.Record(ctx, domain.EventConverted, map[string]any{
			"userId": string(userID), "points": int64(points), "transactionId": tx.ID,
		})
		return nil
	})
	return out, err
}

func (s *BonusLedgerService) appendEntry(
	ctx context.Context,
	userID domain.UserID,
	typ domain.TransactionType,
	amount domain.BonusAmount,
	reference string,
	idempotencyKey *string,
) (domain.BonusTransaction, error) {
	if err := s.accounts.EnsureAccount(ctx, userID); err != nil {
		return domain.BonusTransaction{}, err
	}
	acc, err := s.accounts.GetForUpdate(ctx, userID)
	if err != nil {
		return domain.BonusTransaction{}, err
	}
	delta := domain.SignedDelta(typ, amount)
	if int64(acc.Balance)+delta < 0 {
		return domain.BonusTransaction{}, domain.ErrInsufficientBalance
	}
	entry := domain.BonusTransaction{
		UserID:         userID,
		Amount:         amount,
		Type:           typ,
		Reference:      reference,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.ledger.Append(ctx, entry); err != nil {
		return domain.BonusTransaction{}, err
	}
	newBal := domain.BonusAmount(int64(acc.Balance) + delta)
	if err := s.accounts.UpdateBalance(ctx, userID, newBal); err != nil {
		return domain.BonusTransaction{}, err
	}
	return entry, nil
}

// DebitOneOnOne debits bonus in its own transaction.
func (s *BonusLedgerService) DebitOneOnOne(ctx context.Context, userID, requestID string, amount int64) error {
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		return s.DebitOneOnOneInTx(ctx, userID, requestID, amount)
	})
}

// DebitOneOnOneInTx debits bonus for an approved 1:1 request (idempotent by reference). Caller must run inside a transaction.
func (s *BonusLedgerService) DebitOneOnOneInTx(ctx context.Context, userID, requestID string, amount int64) error {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return err
	}
	pts, err := domain.ParseBonusAmount(amount)
	if err != nil {
		return err
	}
	ref := "one_on_one:" + requestID
	if _, found, err := s.ledger.FindByReference(ctx, uid, ref); err != nil {
		return err
	} else if found {
		return nil
	}
	_, err = s.appendEntry(ctx, uid, domain.TransactionDebit, pts, ref, nil)
	if err == domain.ErrDuplicateOperation {
		return nil
	}
	return err
}

// GetBalance returns current bonus balance for a user.
func (s *BonusLedgerService) GetBalance(ctx context.Context, userID string) (int64, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return 0, err
	}
	_ = s.accounts.EnsureAccount(ctx, uid)
	acc, err := s.accounts.Get(ctx, uid)
	if err != nil {
		return 0, err
	}
	return int64(acc.Balance), nil
}

// MarshalConvertReference builds JSON reference for convert rows.
func MarshalConvertReference(discountPercent int, points int64, idempotencyKey string) string {
	raw, _ := json.Marshal(map[string]any{
		"discountPercent": discountPercent,
		"points":          points,
		"idempotencyKey":  idempotencyKey,
	})
	return string(raw)
}
