package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/google/uuid"
)

func TestStudentGetRoadmapFiltersInactiveMaterials(t *testing.T) {
	blockID := domain.BlockID(uuid.NewString())
	now := time.Now()
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		string(blockID): {
			ID:          blockID,
			Title:       "Pub",
			Status:      domain.BlockStatusPublished,
			IsActive:    true,
			PublishedAt: &now,
			SortOrder:   10,
		},
	}}
	activeID := domain.MaterialID(uuid.NewString())
	inactiveID := domain.MaterialID(uuid.NewString())
	materials := &memMaterials{byID: map[string]domain.Material{
		string(activeID): {
			ID: activeID, BlockID: blockID, Title: "A", IsActive: true, SortOrder: 10,
			MaterialType: domain.MaterialTypeVideo, URL: "https://example.com",
		},
		string(inactiveID): {
			ID: inactiveID, BlockID: blockID, Title: "I", IsActive: false, SortOrder: 20,
			MaterialType: domain.MaterialTypeVideo, URL: "https://example.com/2",
		},
	}}
	svc := application.NewStudentRoadmapService(blocks, materials)
	roadmap, err := svc.GetRoadmap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(roadmap.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(roadmap.Blocks))
	}
	if len(roadmap.Blocks[0].Materials) != 1 {
		t.Fatalf("expected 1 material, got %d", len(roadmap.Blocks[0].Materials))
	}
}
