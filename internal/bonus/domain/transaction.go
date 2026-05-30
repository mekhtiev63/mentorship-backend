package domain

import "time"

// TransactionType is ledger entry type.
type TransactionType string

const (
	TransactionCredit  TransactionType = "credit"
	TransactionDebit   TransactionType = "debit"
	TransactionConvert TransactionType = "convert"
)

// BonusTransaction is an append-only ledger line.
type BonusTransaction struct {
	ID             string
	UserID         UserID
	Amount         BonusAmount
	Type           TransactionType
	Reference      string
	IdempotencyKey *string
	CreatedAt      time.Time
}

// SignedDelta returns balance delta for a transaction.
func SignedDelta(t TransactionType, amount BonusAmount) int64 {
	switch t {
	case TransactionCredit:
		return int64(amount)
	case TransactionDebit, TransactionConvert:
		return -int64(amount)
	default:
		return 0
	}
}
