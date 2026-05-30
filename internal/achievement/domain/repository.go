package domain

import (
	"context"
	"encoding/json"
	"time"
)

// OutboxMessage is a pending outbox row.
type OutboxMessage struct {
	ID        string
	EventName string
	Payload   json.RawMessage
	CreatedAt time.Time
}

// AchievementDefinitionRepository loads definitions.
type AchievementDefinitionRepository interface {
	ListActive(ctx context.Context) ([]AchievementDefinition, error)
	ListActiveByEvent(ctx context.Context, eventName string) ([]AchievementDefinition, error)
}

// UserAchievementRepository persists grants.
type UserAchievementRepository interface {
	Exists(ctx context.Context, userID UserID, code AchievementCode) (bool, error)
	Grant(ctx context.Context, achievement UserAchievement) (created bool, err error)
	ListByUser(ctx context.Context, userID UserID) ([]UserAchievement, error)
}

// ProgressStatsPort reads progress metrics for rules.
type ProgressStatsPort interface {
	CountApprovedBlocks(ctx context.Context, studentID UserID) (int, error)
	CountMaterialViews(ctx context.Context, studentID UserID) (int, error)
}

// RoadmapStatsPort reads roadmap metrics.
type RoadmapStatsPort interface {
	CountPublishedBlocks(ctx context.Context) (int, error)
}

// OutboxReader reads and updates outbox.
type OutboxReader interface {
	ListPendingProgressEvents(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkDone(ctx context.Context, id string) error
}

// BuddyScopePort checks buddy assignment.
type BuddyScopePort interface {
	IsActiveBuddyOf(ctx context.Context, buddyID, studentID UserID) (bool, error)
}

// Transactor runs operations in a DB transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventRecorder writes outbox events after grant.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}
