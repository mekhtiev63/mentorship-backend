package persistence

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/user/domain"
)

// RoleRepo implements domain.RoleRepository.
type RoleRepo struct {
	pool *postgres.Pool
}

// NewRoleRepo creates RoleRepo.
func NewRoleRepo(pool *postgres.Pool) *RoleRepo {
	return &RoleRepo{pool: pool}
}

// ListByUser returns roles for a user.
func (r *RoleRepo) ListByUser(ctx context.Context, userID domain.UserID) (domain.RoleSet, error) {
	const q = `SELECT role::text FROM user_roles WHERE user_id = $1 ORDER BY role`
	rows, err := r.pool.Query(ctx, q, string(userID))
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles domain.RoleSet
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, domain.Role(role))
	}
	return roles, rows.Err()
}

// ReplaceRoles replaces all roles in a transaction.
func (r *RoleRepo) ReplaceRoles(ctx context.Context, userID domain.UserID, roles domain.RoleSet) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, string(userID)); err != nil {
		return fmt.Errorf("delete roles: %w", err)
	}
	for _, role := range roles {
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role) VALUES ($1, $2)`, string(userID), string(role)); err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
	}
	return tx.Commit(ctx)
}
