package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
)

type memBlocks struct {
	byID map[string]domain.RoadmapBlock
}

func (m *memBlocks) Create(_ context.Context, block domain.RoadmapBlock) error {
	m.byID[string(block.ID)] = block
	return nil
}
func (m *memBlocks) Update(_ context.Context, block domain.RoadmapBlock) error {
	if _, ok := m.byID[string(block.ID)]; !ok {
		return domain.ErrNotFound
	}
	m.byID[string(block.ID)] = block
	return nil
}
func (m *memBlocks) SoftDelete(_ context.Context, id domain.BlockID) error {
	b, ok := m.byID[string(id)]
	if !ok {
		return domain.ErrNotFound
	}
	now := time.Now()
	b.DeletedAt = &now
	m.byID[string(id)] = b
	return nil
}
func (m *memBlocks) SoftDeleteMaterialsByBlock(context.Context, domain.BlockID) error { return nil }
func (m *memBlocks) SetActive(_ context.Context, id domain.BlockID, active bool) error {
	b, ok := m.byID[string(id)]
	if !ok {
		return domain.ErrNotFound
	}
	b.IsActive = active
	m.byID[string(id)] = b
	return nil
}
func (m *memBlocks) FindByID(_ context.Context, id domain.BlockID) (domain.RoadmapBlock, error) {
	b, ok := m.byID[string(id)]
	if !ok || b.DeletedAt != nil {
		return domain.RoadmapBlock{}, domain.ErrNotFound
	}
	return b, nil
}
func (m *memBlocks) ListAdmin(context.Context, domain.AdminBlockFilter, pagination.Params) ([]domain.RoadmapBlock, int64, error) {
	return nil, 0, nil
}
func (m *memBlocks) ListPublishedBlocks(_ context.Context) ([]domain.RoadmapBlock, error) {
	var out []domain.RoadmapBlock
	for _, b := range m.byID {
		if b.DeletedAt == nil && b.IsVisibleToStudent() {
			out = append(out, b)
		}
	}
	return out, nil
}
func (m *memBlocks) ReorderBlocks(context.Context, []domain.BlockOrder) error { return nil }
func (m *memBlocks) NextBlockSortOrder(context.Context) (int, error)          { return 10, nil }

type memMaterials struct {
	byID map[string]domain.Material
}

func (m *memMaterials) Create(_ context.Context, mat domain.Material) error {
	m.byID[string(mat.ID)] = mat
	return nil
}
func (m *memMaterials) Update(_ context.Context, mat domain.Material) error {
	if _, ok := m.byID[string(mat.ID)]; !ok {
		return domain.ErrMaterialNotFound
	}
	m.byID[string(mat.ID)] = mat
	return nil
}
func (m *memMaterials) SoftDelete(_ context.Context, id domain.MaterialID) error {
	delete(m.byID, string(id))
	return nil
}
func (m *memMaterials) SetActive(_ context.Context, id domain.MaterialID, active bool) error {
	mat, ok := m.byID[string(id)]
	if !ok {
		return domain.ErrMaterialNotFound
	}
	mat.IsActive = active
	m.byID[string(id)] = mat
	return nil
}
func (m *memMaterials) FindByID(_ context.Context, id domain.MaterialID) (domain.Material, error) {
	mat, ok := m.byID[string(id)]
	if !ok {
		return domain.Material{}, domain.ErrMaterialNotFound
	}
	return mat, nil
}
func (m *memMaterials) ListByBlock(_ context.Context, blockID domain.BlockID, studentOnly bool) ([]domain.Material, error) {
	var out []domain.Material
	for _, mat := range m.byID {
		if mat.BlockID != blockID || mat.DeletedAt != nil {
			continue
		}
		if studentOnly && !mat.IsActive {
			continue
		}
		out = append(out, mat)
	}
	return out, nil
}
func (m *memMaterials) ListByBlocks(_ context.Context, blockIDs []domain.BlockID, studentOnly bool) ([]domain.Material, error) {
	set := make(map[domain.BlockID]struct{}, len(blockIDs))
	for _, id := range blockIDs {
		set[id] = struct{}{}
	}
	var out []domain.Material
	for _, mat := range m.byID {
		if _, ok := set[mat.BlockID]; !ok || mat.DeletedAt != nil {
			continue
		}
		if studentOnly && !mat.IsActive {
			continue
		}
		out = append(out, mat)
	}
	return out, nil
}
func (m *memMaterials) CountActiveByBlock(_ context.Context, blockID domain.BlockID) (int, error) {
	n := 0
	for _, mat := range m.byID {
		if mat.BlockID == blockID && mat.DeletedAt == nil {
			n++
		}
	}
	return n, nil
}
func (m *memMaterials) ReorderMaterials(context.Context, domain.BlockID, []domain.MaterialOrder) error {
	return nil
}
func (m *memMaterials) NextMaterialSortOrder(context.Context, domain.BlockID) (int, error) {
	return 10, nil
}

type memProgress struct{ has bool }

func (m memProgress) HasProgressForBlock(context.Context, domain.BlockID) (bool, error) {
	return m.has, nil
}

type memEvents struct{}

func (memEvents) Record(context.Context, string, map[string]any) error { return nil }

func newAdminTest(progress bool) *application.AdminService {
	blockID := domain.BlockID(uuid.NewString())
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		string(blockID): {
			ID:        blockID,
			SortOrder: 10,
			Title:     "Block",
			Status:    domain.BlockStatusDraft,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}}
	return application.NewAdminService(blocks, &memMaterials{byID: map[string]domain.Material{}}, memProgress{has: progress}, memEvents{})
}

func TestCreateBlockEmptyTitle(t *testing.T) {
	svc := newAdminTest(false)
	_, err := svc.CreateBlock(context.Background(), application.CreateBlockInput{Title: "  "})
	if err != domain.ErrTitleRequired {
		t.Fatalf("expected title error, got %v", err)
	}
}

func TestPublishBlockWithoutMaterials(t *testing.T) {
	svc := newAdminTest(false)
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{}}
	id := domain.BlockID(uuid.NewString())
	blocks.byID[string(id)] = domain.RoadmapBlock{
		ID: id, Title: "T", Status: domain.BlockStatusDraft, IsActive: true,
	}
	svc = application.NewAdminService(blocks, &memMaterials{byID: map[string]domain.Material{}}, memProgress{}, memEvents{})
	_, err := svc.PublishBlock(context.Background(), string(id))
	if err != domain.ErrCannotPublish {
		t.Fatalf("expected cannot publish, got %v", err)
	}
}

func TestDeleteBlockWithProgress(t *testing.T) {
	blockID := domain.BlockID(uuid.NewString())
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		string(blockID): {ID: blockID, Title: "B", Status: domain.BlockStatusDraft, IsActive: true},
	}}
	svc := application.NewAdminService(blocks, &memMaterials{byID: map[string]domain.Material{}}, memProgress{has: true}, memEvents{})
	err := svc.DeleteBlock(context.Background(), string(blockID))
	if err != domain.ErrHasProgress {
		t.Fatalf("expected has progress, got %v", err)
	}
}

func TestReorderBlocksEmpty(t *testing.T) {
	svc := newAdminTest(false)
	err := svc.ReorderBlocks(context.Background(), nil)
	if err != domain.ErrInvalidReorder {
		t.Fatalf("expected invalid reorder, got %v", err)
	}
}

func TestCreateMaterialInvalidURL(t *testing.T) {
	blockID := uuid.NewString()
	blocks := &memBlocks{byID: map[string]domain.RoadmapBlock{
		blockID: {ID: domain.BlockID(blockID), Title: "B", Status: domain.BlockStatusDraft, IsActive: true},
	}}
	svc := application.NewAdminService(blocks, &memMaterials{byID: map[string]domain.Material{}}, memProgress{}, memEvents{})
	_, err := svc.CreateMaterial(context.Background(), application.CreateMaterialInput{
		BlockID: blockID, Title: "M", MaterialType: "video", URL: " ",
	})
	if err != domain.ErrURLRequired {
		t.Fatalf("expected url error, got %v", err)
	}
}
