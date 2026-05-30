package http

import (
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
)

func toBlockOrders(items []reorderBlockItem) ([]domain.BlockOrder, error) {
	orders := make([]domain.BlockOrder, len(items))
	for i, it := range items {
		id, err := domain.ParseBlockID(it.ID)
		if err != nil {
			return nil, err
		}
		orders[i] = domain.BlockOrder{BlockID: id, SortOrder: it.SortOrder}
	}
	return orders, nil
}

func toMaterialOrders(items []reorderMaterialItem) ([]domain.MaterialOrder, error) {
	orders := make([]domain.MaterialOrder, len(items))
	for i, it := range items {
		id, err := domain.ParseMaterialID(it.ID)
		if err != nil {
			return nil, err
		}
		orders[i] = domain.MaterialOrder{MaterialID: id, SortOrder: it.SortOrder}
	}
	return orders, nil
}
