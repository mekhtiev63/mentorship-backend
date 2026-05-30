package domain

import (
	"context"

	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// BlockProgressRepository persists block progress.
type BlockProgressRepository interface {
	Get(ctx context.Context, key BlockProgressKey) (BlockProgress, error)
	Insert(ctx context.Context, progress BlockProgress) error
	Save(ctx context.Context, progress BlockProgress, expectedStatus ProgressStatus) error
	ListByStudent(ctx context.Context, studentID StudentID) ([]BlockProgress, error)
	ListAwaitingForBuddy(ctx context.Context, buddyID UserID, page pagination.Params) ([]BlockProgress, int64, error)
}

// MaterialViewRepository persists material views.
type MaterialViewRepository interface {
	RecordFirstView(ctx context.Context, view MaterialView) (created bool, err error)
	CountViewedInSet(ctx context.Context, studentID StudentID, materialIDs []MaterialID) (int, error)
	ListViewedMaterialIDs(ctx context.Context, studentID StudentID, blockID BlockID) ([]MaterialID, error)
	FirstViewedAtByMaterials(ctx context.Context, studentID StudentID, materialIDs []MaterialID) (map[MaterialID]string, error)
}

// RoadmapProgressPolicyPort reads roadmap visibility rules.
type RoadmapProgressPolicyPort interface {
	MaterialBlockID(ctx context.Context, materialID MaterialID) (BlockID, error)
	IsMaterialVisibleToStudent(ctx context.Context, materialID MaterialID) (bool, error)
	IsBlockVisibleToStudent(ctx context.Context, blockID BlockID) (bool, error)
	ListRequiredMaterialIDs(ctx context.Context, blockID BlockID) ([]MaterialID, error)
	ListPublishedBlocksOrdered(ctx context.Context) ([]RoadmapBlockRef, error)
}

// BuddyScopePort checks buddy assignments.
type BuddyScopePort interface {
	IsActiveBuddyOf(ctx context.Context, buddyID UserID, studentID StudentID) (bool, error)
	ListActiveStudentIDsForBuddy(ctx context.Context, buddyID UserID, page pagination.Params) ([]StudentID, int64, error)
}

// EventRecorder stores outbox events.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}

// Transactor runs a function inside a database transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
