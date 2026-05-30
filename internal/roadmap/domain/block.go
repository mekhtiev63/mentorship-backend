package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// BlockID identifies a roadmap block.
type BlockID string

// ParseBlockID parses a block UUID.
func ParseBlockID(raw string) (BlockID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return BlockID(raw), nil
}

// BlockStatus is publish state.
type BlockStatus string

const (
	BlockStatusDraft     BlockStatus = "draft"
	BlockStatusPublished BlockStatus = "published"
)

// ParseBlockStatus parses block status.
func ParseBlockStatus(raw string) (BlockStatus, error) {
	switch BlockStatus(raw) {
	case BlockStatusDraft, BlockStatusPublished:
		return BlockStatus(raw), nil
	default:
		return "", ErrInvalidStatus
	}
}

// RoadmapBlock is the roadmap block aggregate root.
type RoadmapBlock struct {
	ID              BlockID
	SortOrder       int
	Title           string
	Description     string
	ExpectedSkills  []string
	Status          BlockStatus
	IsActive        bool
	PublishedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// IsVisibleToStudent reports catalog visibility.
func (b RoadmapBlock) IsVisibleToStudent() bool {
	return b.DeletedAt == nil && b.IsActive && b.Status == BlockStatusPublished
}

// ValidateTitle ensures title is non-empty.
func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return ErrTitleRequired
	}
	return nil
}

// ValidateSortOrder ensures sort order is positive.
func ValidateSortOrder(order int) error {
	if order <= 0 {
		return ErrInvalidSortOrder
	}
	return nil
}
