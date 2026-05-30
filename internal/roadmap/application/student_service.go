package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
)

// StudentRoadmapService loads roadmap for authenticated students.
type StudentRoadmapService struct {
	blocks    domain.BlockRepository
	materials domain.MaterialRepository
}

// NewStudentRoadmapService builds StudentRoadmapService.
func NewStudentRoadmapService(blocks domain.BlockRepository, materials domain.MaterialRepository) *StudentRoadmapService {
	return &StudentRoadmapService{blocks: blocks, materials: materials}
}

// GetRoadmap returns full visible roadmap.
func (s *StudentRoadmapService) GetRoadmap(ctx context.Context) (RoadmapDTO, error) {
	blocks, err := s.blocks.ListPublishedBlocks(ctx)
	if err != nil {
		return RoadmapDTO{}, err
	}
	blockIDs := make([]domain.BlockID, len(blocks))
	for i, b := range blocks {
		blockIDs[i] = b.ID
	}
	materials, err := s.materials.ListByBlocks(ctx, blockIDs, true)
	if err != nil {
		return RoadmapDTO{}, err
	}
	return buildRoadmap(blocks, materials), nil
}

// GetBlock returns one visible block with materials.
func (s *StudentRoadmapService) GetBlock(ctx context.Context, blockID string) (BlockWithMaterialsDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockWithMaterialsDTO{}, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return BlockWithMaterialsDTO{}, err
	}
	if !block.IsVisibleToStudent() {
		return BlockWithMaterialsDTO{}, domain.ErrNotFound
	}
	materials, err := s.materials.ListByBlock(ctx, id, true)
	if err != nil {
		return BlockWithMaterialsDTO{}, err
	}
	mdtos := make([]MaterialDTO, len(materials))
	for i, m := range materials {
		mdtos[i] = toMaterialDTO(m)
	}
	return BlockWithMaterialsDTO{Block: toBlockDTO(block), Materials: mdtos}, nil
}

// ListMaterials returns materials for a visible block.
func (s *StudentRoadmapService) ListMaterials(ctx context.Context, blockID string) ([]MaterialDTO, error) {
	id, err := domain.ParseBlockID(blockID)
	if err != nil {
		return nil, err
	}
	block, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !block.IsVisibleToStudent() {
		return nil, domain.ErrNotFound
	}
	materials, err := s.materials.ListByBlock(ctx, id, true)
	if err != nil {
		return nil, err
	}
	out := make([]MaterialDTO, len(materials))
	for i, m := range materials {
		out[i] = toMaterialDTO(m)
	}
	return out, nil
}
