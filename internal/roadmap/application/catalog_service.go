package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
)

// CatalogService serves public roadmap catalog reads.
type CatalogService struct {
	blocks    domain.BlockRepository
	materials domain.MaterialRepository
}

// NewCatalogService builds CatalogService.
func NewCatalogService(blocks domain.BlockRepository, materials domain.MaterialRepository) *CatalogService {
	return &CatalogService{blocks: blocks, materials: materials}
}

// ListPublishedBlocks returns published active blocks without materials.
func (s *CatalogService) ListPublishedBlocks(ctx context.Context) ([]BlockDTO, error) {
	blocks, err := s.blocks.ListPublishedBlocks(ctx)
	if err != nil {
		return nil, err
	}
	return toBlocksDTO(blocks), nil
}

// GetPublishedBlock returns a visible block with student-visible materials.
func (s *CatalogService) GetPublishedBlock(ctx context.Context, blockID string) (BlockWithMaterialsDTO, error) {
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
