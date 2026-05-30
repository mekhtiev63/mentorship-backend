package persistence

import (
	"context"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
	"github.com/jackc/pgx/v5"
)

// AccountRepo implements BonusAccountRepository.
type AccountRepo struct {
	pool *postgres.Pool
}

// NewAccountRepo creates AccountRepo.
func NewAccountRepo(pool *postgres.Pool) *AccountRepo {
	return &AccountRepo{pool: pool}
}

// Get loads account without lock.
func (r *AccountRepo) Get(ctx context.Context, userID domain.UserID) (domain.BonusAccount, error) {
	const q = `SELECT user_id, balance, updated_at FROM bonus_accounts WHERE user_id = $1`
	var acc domain.BonusAccount
	var uid string
	var bal int64
	err := r.pool.QueryRow(ctx, q, string(userID)).Scan(&uid, &bal, &acc.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BonusAccount{UserID: userID, Balance: 0}, nil
	}
	if err != nil {
		return domain.BonusAccount{}, err
	}
	acc.UserID = domain.UserID(uid)
	acc.Balance = domain.BonusAmount(bal)
	return acc, nil
}

// GetForUpdate loads account with row lock.
func (r *AccountRepo) GetForUpdate(ctx context.Context, userID domain.UserID) (domain.BonusAccount, error) {
	const q = `SELECT user_id, balance, updated_at FROM bonus_accounts WHERE user_id = $1 FOR UPDATE`
	var acc domain.BonusAccount
	var uid string
	var bal int64
	err := db(ctx, r.pool).QueryRow(ctx, q, string(userID)).Scan(&uid, &bal, &acc.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BonusAccount{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.BonusAccount{}, err
	}
	acc.UserID = domain.UserID(uid)
	acc.Balance = domain.BonusAmount(bal)
	return acc, nil
}

// EnsureAccount creates zero balance account if missing.
func (r *AccountRepo) EnsureAccount(ctx context.Context, userID domain.UserID) error {
	const q = `
		INSERT INTO bonus_accounts (user_id, balance)
		VALUES ($1, 0)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := db(ctx, r.pool).Exec(ctx, q, string(userID))
	return err
}

// UpdateBalance sets materialized balance.
func (r *AccountRepo) UpdateBalance(ctx context.Context, userID domain.UserID, balance domain.BonusAmount) error {
	const q = `UPDATE bonus_accounts SET balance = $2, updated_at = now() WHERE user_id = $1`
	ct, err := db(ctx, r.pool).Exec(ctx, q, string(userID), int64(balance))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
