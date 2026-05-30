package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
)

func persistProgress(ctx context.Context, repo domain.BlockProgressRepository, p *domain.BlockProgress, before domain.ProgressStatus) error {
	if !p.Exists {
		if err := repo.Insert(ctx, *p); err != nil {
			return err
		}
		p.Exists = true
		return nil
	}
	return repo.Save(ctx, *p, before)
}

func progressByBlock(list []domain.BlockProgress) map[domain.BlockID]domain.BlockProgress {
	m := make(map[domain.BlockID]domain.BlockProgress, len(list))
	for _, p := range list {
		m[p.BlockID] = p
	}
	return m
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
