package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/bonus/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TransactionRepo implements BonusTransactionRepository.
type TransactionRepo struct {
	pool *postgres.Pool
}

// NewTransactionRepo creates TransactionRepo.
func NewTransactionRepo(pool *postgres.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

// Append inserts ledger line.
func (r *TransactionRepo) Append(ctx context.Context, tx domain.BonusTransaction) error {
	id := tx.ID
	if id == "" {
		id = uuid.NewString()
	}
	const q = `
		INSERT INTO bonus_transactions (id, user_id, amount, type, reference, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(tx.UserID), int64(tx.Amount), string(tx.Type), nullString(tx.Reference), tx.IdempotencyKey,
	)
	if isUnique(err) {
		return domain.ErrDuplicateOperation
	}
	return err
}

// FindByIdempotencyKey finds convert by idempotency key.
func (r *TransactionRepo) FindByIdempotencyKey(ctx context.Context, key string) (domain.BonusTransaction, bool, error) {
	const q = `
		SELECT id, user_id, amount, type, reference, idempotency_key, created_at
		FROM bonus_transactions WHERE idempotency_key = $1
	`
	return scanOne(r, ctx, q, key)
}

// FindByReference finds credit by reference.
func (r *TransactionRepo) FindByReference(ctx context.Context, userID domain.UserID, reference string) (domain.BonusTransaction, bool, error) {
	const q = `
		SELECT id, user_id, amount, type, reference, idempotency_key, created_at
		FROM bonus_transactions WHERE user_id = $1 AND reference = $2
	`
	return scanOne(r, ctx, q, string(userID), reference)
}

// ListByUser returns paginated ledger.
func (r *TransactionRepo) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Params) ([]domain.BonusTransaction, int64, error) {
	const countQ = `SELECT COUNT(*) FROM bonus_transactions WHERE user_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, string(userID)).Scan(&total); err != nil {
		return nil, 0, err
	}
	const q = `
		SELECT id, user_id, amount, type, reference, idempotency_key, created_at
		FROM bonus_transactions WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, string(userID), page.PageSize, page.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.BonusTransaction
	for rows.Next() {
		tx, err := scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, tx)
	}
	return out, total, rows.Err()
}

type convertRef struct {
	DiscountPercent int `json:"discountPercent"`
}

// SumConvertDiscountPercent sums discount % from convert references.
func (r *TransactionRepo) SumConvertDiscountPercent(ctx context.Context, userID domain.UserID) (int, error) {
	const q = `
		SELECT reference FROM bonus_transactions
		WHERE user_id = $1 AND type = 'convert' AND reference IS NOT NULL
	`
	rows, err := r.pool.Query(ctx, q, string(userID))
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	sum := 0
	for rows.Next() {
		var ref *string
		if err := rows.Scan(&ref); err != nil {
			return 0, err
		}
		if ref == nil {
			continue
		}
		var cr convertRef
		if err := json.Unmarshal([]byte(*ref), &cr); err != nil {
			continue
		}
		sum += cr.DiscountPercent
		if sum > domain.MaxDiscountPercent {
			return domain.MaxDiscountPercent, nil
		}
	}
	return sum, rows.Err()
}

func scanOne(r *TransactionRepo, ctx context.Context, q string, args ...any) (domain.BonusTransaction, bool, error) {
	row := db(ctx, r.pool).QueryRow(ctx, q, args...)
	tx, err := scanRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BonusTransaction{}, false, nil
	}
	if err != nil {
		return domain.BonusTransaction{}, false, err
	}
	return tx, true, nil
}

func scanRow(row pgx.Row) (domain.BonusTransaction, error) {
	var tx domain.BonusTransaction
	var uid string
	var amount int64
	var typ string
	var ref *string
	var key *string
	if err := row.Scan(&tx.ID, &uid, &amount, &typ, &ref, &key, &tx.CreatedAt); err != nil {
		return domain.BonusTransaction{}, err
	}
	tx.UserID = domain.UserID(uid)
	tx.Amount = domain.BonusAmount(amount)
	tx.Type = domain.TransactionType(typ)
	if ref != nil {
		tx.Reference = *ref
	}
	tx.IdempotencyKey = key
	return tx, nil
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
