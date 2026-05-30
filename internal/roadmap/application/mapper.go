package application

import (
	"time"

	"github.com/go-mentorship-platform/backend/internal/roadmap/domain"
)

func toBlockDTO(b domain.RoadmapBlock) BlockDTO {
	skills := b.ExpectedSkills
	if skills == nil {
		skills = []string{}
	}
	dto := BlockDTO{
		ID:             string(b.ID),
		SortOrder:      b.SortOrder,
		Title:          b.Title,
		Description:    b.Description,
		ExpectedSkills: skills,
		Status:         string(b.Status),
		IsActive:       b.IsActive,
		CreatedAt:      b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      b.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if b.PublishedAt != nil {
		s := b.PublishedAt.UTC().Format(time.RFC3339)
		dto.PublishedAt = &s
	}
	return dto
}

func toMaterialDTO(m domain.Material) MaterialDTO {
	return MaterialDTO{
		ID:           string(m.ID),
		BlockID:      string(m.BlockID),
		SortOrder:    m.SortOrder,
		Title:        m.Title,
		MaterialType: string(m.MaterialType),
		URL:          m.URL,
		Required:     m.Required,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    m.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toBlocksDTO(blocks []domain.RoadmapBlock) []BlockDTO {
	out := make([]BlockDTO, len(blocks))
	for i, b := range blocks {
		out[i] = toBlockDTO(b)
	}
	return out
}

func groupMaterialsByBlock(materials []domain.Material) map[domain.BlockID][]domain.Material {
	m := make(map[domain.BlockID][]domain.Material)
	for _, mat := range materials {
		m[mat.BlockID] = append(m[mat.BlockID], mat)
	}
	return m
}

func buildRoadmap(blocks []domain.RoadmapBlock, materials []domain.Material) RoadmapDTO {
	byBlock := groupMaterialsByBlock(materials)
	out := RoadmapDTO{Blocks: make([]BlockWithMaterialsDTO, 0, len(blocks))}
	for _, b := range blocks {
		mats := byBlock[b.ID]
		mdtos := make([]MaterialDTO, len(mats))
		for i, mat := range mats {
			mdtos[i] = toMaterialDTO(mat)
		}
		out.Blocks = append(out.Blocks, BlockWithMaterialsDTO{
			Block:     toBlockDTO(b),
			Materials: mdtos,
		})
	}
	return out
}
