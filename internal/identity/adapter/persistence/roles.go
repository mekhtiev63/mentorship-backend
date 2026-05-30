package persistence

import (
	"context"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/identity/domain"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
)

// RoleRepo reads user_roles.
type RoleRepo struct {
	pool *postgres.Pool
}

// NewRoleRepo creates a RoleRepo.
func NewRoleRepo(pool *postgres.Pool) *RoleRepo {
	return &RoleRepo{pool: pool}
}

// ListByUser returns roles for a user.
func (r *RoleRepo) ListByUser(ctx context.Context, userID domain.UserID) (domain.RoleSet, error) {
	const q = `
		SELECT role::text
		FROM user_roles
		WHERE user_id = $1
		ORDER BY role
	`
	rows, err := r.pool.Query(ctx, q, userID.String())
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	var roles domain.RoleSet
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, domain.Role(role))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return roles, nil
}
