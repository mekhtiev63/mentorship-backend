package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
)

// AdminService manages roadmap blocks and materials for admins.
type AdminService struct {
	blocks   domain.BlockRepository
	materials domain.MaterialRepository
	progress domain.ProgressExistenceReader
	events   domain.EventRecorder
}

// NewAdminService builds AdminService.
func NewAdminService(
	blocks domain.BlockRepository,
	materials domain.MaterialRepository,
	progress domain.ProgressExistenceReader,
	events domain.EventRecorder,
) *AdminService {
	return &AdminService{
		blocks:    blocks,
		materials: materials,
		progress:  progress,
		events:    events,
	}
}

// CreateBlockInput is input for block creation.
type CreateBlockInput struct {
	Title       string
	Description string
}

// CreateBlock creates a draft block at the end of the sort order.
func (s *AdminService) CreateBlock(ctx context.Context, in CreateBlockInput) (BlockDTO, error) {
	if err := domain.ValidateTitle(in.Title); err != nil {
		return BlockDTO{}, err
	}
	sortOrder, err := s.blocks.NextBlockSortOrder(ctx)
	if err != nil {
		return BlockDTO{}, fmt.Errorf("next sort order: %w", err)
	}
	now := time.Now().UTC()
	block := domain.RoadmapBlock{
		ID:          domain.BlockID(uuid.NewString()),
		SortOrder:   sortOrder,
		Title:       in.Title,
		Description: in.Description,
		Status:      domain.BlockStatusDraft,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.blocks.Create(ctx, block); err != nil {
		return BlockDTO{}, err
	}
	return toBlockDTO(block), nil
}

// UpdateBlockInput updates block content and optional status.
type UpdateBlockInput struct {
	Title       string
	Description string
	Status      *string
}

// UpdateBlock updates block fields.
func (s *AdminService) UpdateBlock(ctx context.Context, blockID string, in UpdateBlockInput) (BlockDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockDTO{}, err
	}
	if err := domain.ValidateTitle(in.Title); err != nil {
		return BlockDTO{}, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return BlockDTO{}, err
	}
	block.Title = in.Title
	block.Description = in.Description
	if in.Status != nil {
		st, err := domain.ParseBlockStatus(*in.Status)
		if err != nil {
			return BlockDTO{}, err
		}
		block, err = applyBlockStatus(block, st)
		if err != nil {
			return BlockDTO{}, err
		}
	}
	if err := s.blocks.Update(ctx, block); err != nil {
		return BlockDTO{}, err
	}
	s.recordStatusEvents(ctx, block)
	return toBlockDTO(block), nil
}

// DeleteBlock soft-deletes block and its materials when no progress exists.
func (s *AdminService) DeleteBlock(ctx context.Context, blockID string) error {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return err
	}
	has, err := s.progress.HasProgressForBlock(ctx, id)
	if err != nil {
		return fmt.Errorf("progress check: %w", err)
	}
	if has {
		return domain.ErrHasProgress
	}
	if err := s.blocks.SoftDeleteMaterialsByBlock(ctx, id); err != nil {
		return err
	}
	return s.blocks.SoftDelete(ctx, id)
}

// SetBlockActive activates or deactivates a block.
func (s *AdminService) SetBlockActive(ctx context.Context, blockID string, active bool) error {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return err
	}
	if err := s.blocks.SetActive(ctx, id, active); err != nil {
		return err
	}
	if !active {
		_ = s.events.Record(ctx, domain.EventBlockDeactivated, map[string]any{"blockId": blockID})
	}
	return nil
}

// PublishBlock sets status to published.
func (s *AdminService) PublishBlock(ctx context.Context, blockID string) (BlockDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockDTO{}, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return BlockDTO{}, err
	}
	count, err := s.materials.CountActiveByBlock(ctx, id)
	if err != nil {
		return BlockDTO{}, err
	}
	if count == 0 {
		return BlockDTO{}, domain.ErrCannotPublish
	}
	block, err = applyBlockStatus(block, domain.BlockStatusPublished)
	if err != nil {
		return BlockDTO{}, err
	}
	if err := s.blocks.Update(ctx, block); err != nil {
		return BlockDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventBlockPublished, map[string]any{"blockId": blockID})
	return toBlockDTO(block), nil
}

// UnpublishBlock sets status back to draft.
func (s *AdminService) UnpublishBlock(ctx context.Context, blockID string) (BlockDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockDTO{}, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return BlockDTO{}, err
	}
	block, err = applyBlockStatus(block, domain.BlockStatusDraft)
	if err != nil {
		return BlockDTO{}, err
	}
	if err := s.blocks.Update(ctx, block); err != nil {
		return BlockDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventBlockUnpublished, map[string]any{"blockId": blockID})
	return toBlockDTO(block), nil
}

// ReorderBlocks updates block sort orders.
func (s *AdminService) ReorderBlocks(ctx context.Context, orders []domain.BlockOrder) error {
	if err := validateBlockReorder(orders); err != nil {
		return err
	}
	if err := s.blocks.ReorderBlocks(ctx, orders); err != nil {
		return err
	}
	_ = s.events.Record(ctx, domain.EventBlockReordered, map[string]any{"count": len(orders)})
	return nil
}

// ListBlocksAdmin returns paginated admin list.
func (s *AdminService) ListBlocksAdmin(ctx context.Context, filter domain.AdminBlockFilter, page pagination.Params) ([]BlockDTO, int64, error) {
	blocks, total, err := s.blocks.ListAdmin(ctx, filter, page)
	if err != nil {
		return nil, 0, err
	}
	return toBlocksDTO(blocks), total, nil
}

// GetBlockAdmin returns block by id.
func (s *AdminService) GetBlockAdmin(ctx context.Context, blockID string) (BlockDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockDTO{}, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return BlockDTO{}, err
	}
	return toBlockDTO(block), nil
}

// CreateMaterialInput is input for material creation.
type CreateMaterialInput struct {
	BlockID      string
	Title        string
	MaterialType string
	URL          string
	Required     bool
}

// CreateMaterial adds material to a block.
func (s *AdminService) CreateMaterial(ctx context.Context, in CreateMaterialInput) (MaterialDTO, error) {
	blockID, err := domain.ParseBlockID(in.BlockID)
	if err != nil {
		return MaterialDTO{}, err
	}
	if _, err := s.blocks.FindByID(ctx, blockID); err != nil {
		return MaterialDTO{}, err
	}
	if err := domain.ValidateTitle(in.Title); err != nil {
		return MaterialDTO{}, err
	}
	mt, err := domain.ParseMaterialType(in.MaterialType)
	if err != nil {
		return MaterialDTO{}, err
	}
	if err := domain.ValidateURL(in.URL); err != nil {
		return MaterialDTO{}, err
	}
	sortOrder, err := s.materials.NextMaterialSortOrder(ctx, blockID)
	if err != nil {
		return MaterialDTO{}, fmt.Errorf("next material sort: %w", err)
	}
	now := time.Now().UTC()
	material := domain.Material{
		ID:           domain.MaterialID(uuid.NewString()),
		BlockID:      blockID,
		SortOrder:    sortOrder,
		Title:        in.Title,
		MaterialType: mt,
		URL:          in.URL,
		Required:     in.Required,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.materials.Create(ctx, material); err != nil {
		return MaterialDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventMaterialCreated, map[string]any{
		"materialId": string(material.ID),
		"blockId":    string(blockID),
	})
	return toMaterialDTO(material), nil
}

// UpdateMaterialInput updates material fields.
type UpdateMaterialInput struct {
	Title        string
	MaterialType string
	URL          string
	Required     bool
}

// UpdateMaterial updates a material.
func (s *AdminService) UpdateMaterial(ctx context.Context, materialID string, in UpdateMaterialInput) (MaterialDTO, error) {
	id, err := domain.ParseMaterialID(materialID)
	if err != nil {
		return MaterialDTO{}, err
	}
	if err := domain.ValidateTitle(in.Title); err != nil {
		return MaterialDTO{}, err
	}
	mt, err := domain.ParseMaterialType(in.MaterialType)
	if err != nil {
		return MaterialDTO{}, err
	}
	if err := domain.ValidateURL(in.URL); err != nil {
		return MaterialDTO{}, err
	}
	material, err := s.materials.FindByID(ctx, id)
	if err != nil {
		return MaterialDTO{}, err
	}
	material.Title = in.Title
	material.MaterialType = mt
	material.URL = in.URL
	material.Required = in.Required
	if err := s.materials.Update(ctx, material); err != nil {
		return MaterialDTO{}, err
	}
	return toMaterialDTO(material), nil
}

// DeleteMaterial soft-deletes a material.
func (s *AdminService) DeleteMaterial(ctx context.Context, materialID string) error {
	id, err := domain.ParseMaterialID(materialID)
	if err != nil {
		return err
	}
	return s.materials.SoftDelete(ctx, id)
}

// SetMaterialActive toggles material visibility.
func (s *AdminService) SetMaterialActive(ctx context.Context, materialID string, active bool) error {
	id, err := domain.ParseMaterialID(materialID)
	if err != nil {
		return err
	}
	if err := s.materials.SetActive(ctx, id, active); err != nil {
		return err
	}
	if !active {
		_ = s.events.Record(ctx, domain.EventMaterialDeactivated, map[string]any{"materialId": materialID})
	}
	return nil
}

// ReorderMaterials updates material order within a block.
func (s *AdminService) ReorderMaterials(ctx context.Context, blockID string, orders []domain.MaterialOrder) error {
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return err
	}
	if err := validateMaterialReorder(orders); err != nil {
		return err
	}
	if err := s.materials.ReorderMaterials(ctx, bid, orders); err != nil {
		return err
	}
	_ = s.events.Record(ctx, domain.EventMaterialReordered, map[string]any{
		"blockId": blockID,
		"count":   len(orders),
	})
	return nil
}

// ListMaterialsAdmin lists all materials for a block.
func (s *AdminService) ListMaterialsAdmin(ctx context.Context, blockID string) ([]MaterialDTO, error) {
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return nil, err
	}
	materials, err := s.materials.ListByBlock(ctx, bid, false)
	if err != nil {
		return nil, err
	}
	out := make([]MaterialDTO, len(materials))
	for i, m := range materials {
		out[i] = toMaterialDTO(m)
	}
	return out, nil
}

func applyBlockStatus(block domain.RoadmapBlock, status domain.BlockStatus) (domain.RoadmapBlock, error) {
	switch status {
	case domain.BlockStatusDraft:
		block.Status = domain.BlockStatusDraft
		block.PublishedAt = nil
	case domain.BlockStatusPublished:
		now := time.Now().UTC()
		block.Status = domain.BlockStatusPublished
		block.PublishedAt = &now
	default:
		return block, domain.ErrInvalidStatus
	}
	return block, nil
}

func (s *AdminService) recordStatusEvents(ctx context.Context, block domain.RoadmapBlock) {
	if block.Status == domain.BlockStatusPublished {
		_ = s.events.Record(ctx, domain.EventBlockPublished, map[string]any{"blockId": string(block.ID)})
	}
}
