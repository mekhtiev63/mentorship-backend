package persistence

import (
	"context"
	"encoding/json"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/achievement/domain"
)

// DefinitionRepo implements AchievementDefinitionRepository.
type DefinitionRepo struct {
	pool *postgres.Pool
}

// NewDefinitionRepo creates DefinitionRepo.
func NewDefinitionRepo(pool *postgres.Pool) *DefinitionRepo {
	return &DefinitionRepo{pool: pool}
}

// ListActive returns non-deleted definitions.
func (r *DefinitionRepo) ListActive(ctx context.Context) ([]domain.AchievementDefinition, error) {
	const q = `
		SELECT code, title, description, rule_json, created_at, updated_at
		FROM achievement_definitions
		WHERE deleted_at IS NULL
		ORDER BY code ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDefinitions(rows)
}

// ListActiveByEvent returns definitions whose rule.on matches event.
func (r *DefinitionRepo) ListActiveByEvent(ctx context.Context, eventName string) ([]domain.AchievementDefinition, error) {
	const q = `
		SELECT code, title, description, rule_json, created_at, updated_at
		FROM achievement_definitions
		WHERE deleted_at IS NULL AND rule_json->>'on' = $1
		ORDER BY code ASC
	`
	rows, err := r.pool.Query(ctx, q, eventName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDefinitions(rows)
}

func scanDefinitions(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.AchievementDefinition, error) {
	var out []domain.AchievementDefinition
	for rows.Next() {
		var d domain.AchievementDefinition
		var code string
		var raw json.RawMessage
		if err := rows.Scan(&code, &d.Title, &d.Description, &raw, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		rule, err := domain.ParseRule(raw)
		if err != nil {
			return nil, err
		}
		d.Code = domain.AchievementCode(code)
		d.Rule = rule
		out = append(out, d)
	}
	return out, rows.Err()
}
