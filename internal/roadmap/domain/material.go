package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaterialID identifies a material.
type MaterialID string

// ParseMaterialID parses material UUID.
func ParseMaterialID(raw string) (MaterialID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrMaterialNotFound
	}
	return MaterialID(raw), nil
}

// MaterialType matches DB enum.
type MaterialType string

const (
	MaterialTypeVideo   MaterialType = "video"
	MaterialTypeArticle MaterialType = "article"
	MaterialTypeTask    MaterialType = "task"
	MaterialTypeLink    MaterialType = "link"
)

// ParseMaterialType parses material type.
func ParseMaterialType(raw string) (MaterialType, error) {
	switch MaterialType(raw) {
	case MaterialTypeVideo, MaterialTypeArticle, MaterialTypeTask, MaterialTypeLink:
		return MaterialType(raw), nil
	default:
		return "", ErrInvalidMaterialType
	}
}

// Material is an entity within RoadmapBlock aggregate.
type Material struct {
	ID           MaterialID
	BlockID      BlockID
	SortOrder    int
	Title        string
	MaterialType MaterialType
	URL          string
	Required     bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// IsVisibleToStudent reports student catalog visibility.
func (m Material) IsVisibleToStudent() bool {
	return m.DeletedAt == nil && m.IsActive
}

// ValidateURL ensures url is set.
func ValidateURL(url string) error {
	if strings.TrimSpace(url) == "" {
		return ErrURLRequired
	}
	return nil
}
