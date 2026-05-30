package domain

import (
	"context"
)

// FinalAssessmentRepository persists assessments.
type FinalAssessmentRepository interface {
	GetByStudentID(ctx context.Context, studentID StudentID) (FinalAssessment, error)
	GetForUpdateByStudentID(ctx context.Context, studentID StudentID) (FinalAssessment, error)
	Insert(ctx context.Context, a FinalAssessment) error
	Save(ctx context.Context, a FinalAssessment, expectedTech, expectedRoast TrackStatus) error
}

// RoadmapCompletionPort checks program completion.
type RoadmapCompletionPort interface {
	IsProgramCompleted(ctx context.Context, studentID StudentID) (bool, error)
}

// BuddyScopePort checks buddy assignment.
type BuddyScopePort interface {
	IsActiveBuddyOf(ctx context.Context, buddyID UserID, studentID StudentID) (bool, error)
}

// Transactor runs DB transactions.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventRecorder writes outbox events.
type EventRecorder interface {
	Record(ctx context.Context, name string, payload map[string]any) error
}
