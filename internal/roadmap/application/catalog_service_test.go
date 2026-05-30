package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/google/uuid"
)

func TestGetPublishedBlockHiddenDraft(t *testing.T) {
	blockID := domain.BlockID(uuid.NewString())
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		string(blockID): {
			ID:       blockID,
			Title:    "Draft",
			Status:   domain.BlockStatusDraft,
			IsActive: true,
		},
	}}
	svc := application.NewCatalogService(blocks, &memMaterials{byID: map[string]domain.Material{}})
	_, err := svc.GetPublishedBlock(context.Background(), string(blockID))
	if err != domain.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetPublishedBlockVisible(t *testing.T) {
	blockID := domain.BlockID(uuid.NewString())
	now := time.Now()
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		string(blockID): {
			ID:          blockID,
			Title:       "Pub",
			Status:      domain.BlockStatusPublished,
			IsActive:    true,
			PublishedAt: &now,
		},
	}}
	materials := &memMaterials{byID: map[string]domain.Material{}}
	svc := application.NewCatalogService(blocks, materials)
	got, err := svc.GetPublishedBlock(context.Background(), string(blockID))
	if err != nil {
		t.Fatal(err)
	}
	if got.Block.ID != string(blockID) {
		t.Fatalf("unexpected block id %s", got.Block.ID)
	}
}
