package domain

import (
	"context"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// BlockOrder maps block id to new sort order.
type BlockOrder struct {
	BlockID   BlockID
	SortOrder int
}

// MaterialOrder maps material id to new sort order.
type MaterialOrder struct {
	MaterialID MaterialID
	SortOrder  int
}

// AdminBlockFilter filters admin block list.
type AdminBlockFilter struct {
	Status   *BlockStatus
	IsActive *bool
}

// BlockRepository persists roadmap blocks.
type BlockRepository interface {
	Create(ctx context.Context, block RoadmapBlock) error
	Update(ctx context.Context, block RoadmapBlock) error
	SoftDelete(ctx context.Context, blockID BlockID) error
	SoftDeleteMaterialsByBlock(ctx context.Context, blockID BlockID) error
	SetActive(ctx context.Context, blockID BlockID, active bool) error
	FindByID(ctx context.Context, blockID BlockID) (RoadmapBlock, error)
	ListAdmin(ctx context.Context, filter AdminBlockFilter, page pagination.Params) ([]RoadmapBlock, int64, error)
	ListPublishedBlocks(ctx context.Context) ([]RoadmapBlock, error)
	ReorderBlocks(ctx context.Context, orders []BlockOrder) error
	NextBlockSortOrder(ctx context.Context) (int, error)
}

// MaterialRepository persists materials.
type MaterialRepository interface {
	Create(ctx context.Context, material Material) error
	Update(ctx context.Context, material Material) error
	SoftDelete(ctx context.Context, materialID MaterialID) error
	SetActive(ctx context.Context, materialID MaterialID, active bool) error
	FindByID(ctx context.Context, materialID MaterialID) (Material, error)
	ListByBlock(ctx context.Context, blockID BlockID, studentVisibleOnly bool) ([]Material, error)
	ListByBlocks(ctx context.Context, blockIDs []BlockID, studentVisibleOnly bool) ([]Material, error)
	CountActiveByBlock(ctx context.Context, blockID BlockID) (int, error)
	ReorderMaterials(ctx context.Context, blockID BlockID, orders []MaterialOrder) error
	NextMaterialSortOrder(ctx context.Context, blockID BlockID) (int, error)
}

// ProgressExistenceReader checks progress rows for a block.
type ProgressExistenceReader interface {
	HasProgressForBlock(ctx context.Context, blockID BlockID) (bool, error)
}

// EventRecorder stores domain events (outbox).
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}
