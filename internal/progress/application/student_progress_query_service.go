package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
)

// StudentProgressQueryService reads student progress.
type StudentProgressQueryService struct {
	progress domain.BlockProgressRepository
	views    domain.MaterialViewRepository
	roadmap  domain.RoadmapProgressPolicyPort
}

// NewStudentProgressQueryService builds StudentProgressQueryService.
func NewStudentProgressQueryService(
	progress domain.BlockProgressRepository,
	views domain.MaterialViewRepository,
	roadmap domain.RoadmapProgressPolicyPort,
) *StudentProgressQueryService {
	return &StudentProgressQueryService{progress: progress, views: views, roadmap: roadmap}
}

// ListMyBlocks returns progress for all published blocks.
func (s *StudentProgressQueryService) ListMyBlocks(ctx context.Context, studentID string) ([]BlockProgressDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return nil, err
	}
	ordered, err := s.roadmap.ListPublishedBlocksOrdered(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.progress.ListByStudent(ctx, sid)
	if err != nil {
		return nil, err
	}
	byBlock := progressByBlock(rows)

	out := make([]BlockProgressDTO, 0, len(ordered))
	for _, ref := range ordered {
		p := byBlock[ref.BlockID]
		status := string(p.Status)
		if status == "" {
			status = string(domain.StatusNotStarted)
		}
		required, err := s.roadmap.ListRequiredMaterialIDs(ctx, ref.BlockID)
		if err != nil {
			return nil, err
		}
		viewed, err := s.views.CountViewedInSet(ctx, sid, required)
		if err != nil {
			return nil, err
		}
		out = append(out, BlockProgressDTO{
			BlockID:   string(ref.BlockID),
			SortOrder: ref.SortOrder,
			Title:     ref.Title,
			Status:    status,
			SubmittedAt: formatTimePtr(p.SubmittedAt),
			ApprovedAt:  formatTimePtr(p.ApprovedAt),
			RejectedAt:  formatTimePtr(p.RejectedAt),
			RejectReason: p.RejectReason,
			Required:  len(required),
			Viewed:    viewed,
		})
	}
	return out, nil
}

// GetMyBlock returns block detail with materials.
func (s *StudentProgressQueryService) GetMyBlock(ctx context.Context, studentID, blockID string) (BlockDetailDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	visible, err := s.roadmap.IsBlockVisibleToStudent(ctx, bid)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	if !visible {
		return BlockDetailDTO{}, domain.ErrBlockNotVisible
	}
	ordered, err := s.roadmap.ListPublishedBlocksOrdered(ctx)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	var ref *domain.RoadmapBlockRef
	for i := range ordered {
		if ordered[i].BlockID == bid {
			ref = &ordered[i]
			break
		}
	}
	if ref == nil {
		return BlockDetailDTO{}, domain.ErrBlockNotVisible
	}

	key := domain.BlockProgressKey{StudentID: sid, BlockID: bid}
	p, err := s.progress.Get(ctx, key)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	status := string(p.Status)
	if status == "" {
		status = string(domain.StatusNotStarted)
	}
	required, err := s.roadmap.ListRequiredMaterialIDs(ctx, bid)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	viewedIDs, err := s.views.ListViewedMaterialIDs(ctx, sid, bid)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	viewedSet := make(map[domain.MaterialID]struct{}, len(viewedIDs))
	for _, id := range viewedIDs {
		viewedSet[id] = struct{}{}
	}
	viewedCount, err := s.views.CountViewedInSet(ctx, sid, required)
	if err != nil {
		return BlockDetailDTO{}, err
	}
	firstViewed, err := s.views.FirstViewedAtByMaterials(ctx, sid, required)
	if err != nil {
		return BlockDetailDTO{}, err
	}

	materials := make([]MaterialItemDTO, 0, len(required))
	for _, id := range required {
		_, ok := viewedSet[id]
		var first *string
		if ts, has := firstViewed[id]; has {
			first = &ts
		}
		materials = append(materials, MaterialItemDTO{
			MaterialID:      string(id),
			Required:        true,
			Viewed:          ok,
			FirstViewedAt:   first,
		})
	}

	blockDTO := BlockProgressDTO{
		BlockID:      blockID,
		SortOrder:    ref.SortOrder,
		Title:        ref.Title,
		Status:       status,
		SubmittedAt:  formatTimePtr(p.SubmittedAt),
		ApprovedAt:   formatTimePtr(p.ApprovedAt),
		RejectedAt:   formatTimePtr(p.RejectedAt),
		RejectReason: p.RejectReason,
		Required:     len(required),
		Viewed:       viewedCount,
	}
	return BlockDetailDTO{Block: blockDTO, Materials: materials}, nil
}
