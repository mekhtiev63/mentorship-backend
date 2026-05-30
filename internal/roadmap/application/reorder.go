package application

import (
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
)

func validateBlockReorder(orders []domain.BlockOrder) error {
	if len(orders) == 0 {
		return domain.ErrInvalidReorder
	}
	seen := make(map[domain.BlockID]struct{}, len(orders))
	for _, o := range orders {
		if err := domain.ValidateSortOrder(o.SortOrder); err != nil {
			return err
		}
		if _, dup := seen[o.BlockID]; dup {
			return domain.ErrInvalidReorder
		}
		seen[o.BlockID] = struct{}{}
	}
	return nil
}

func validateMaterialReorder(orders []domain.MaterialOrder) error {
	if len(orders) == 0 {
		return domain.ErrInvalidReorder
	}
	seen := make(map[domain.MaterialID]struct{}, len(orders))
	for _, o := range orders {
		if err := domain.ValidateSortOrder(o.SortOrder); err != nil {
			return err
		}
		if _, dup := seen[o.MaterialID]; dup {
			return domain.ErrInvalidReorder
		}
		seen[o.MaterialID] = struct{}{}
	}
	return nil
}
