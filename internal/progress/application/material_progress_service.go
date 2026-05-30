package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
)

// MaterialProgressService handles material views and block submit.
type MaterialProgressService struct {
	tx        domain.Transactor
	progress  domain.BlockProgressRepository
	views     domain.MaterialViewRepository
	roadmap   domain.RoadmapProgressPolicyPort
	events    domain.EventRecorder
	sequential domain.SequentialBlockPolicy
}

// NewMaterialProgressService builds MaterialProgressService.
func NewMaterialProgressService(
	tx domain.Transactor,
	progress domain.BlockProgressRepository,
	views domain.MaterialViewRepository,
	roadmap domain.RoadmapProgressPolicyPort,
	events domain.EventRecorder,
) *MaterialProgressService {
	return &MaterialProgressService{
		tx:         tx,
		progress:   progress,
		views:      views,
		roadmap:    roadmap,
		events:     events,
		sequential: domain.SequentialBlockPolicy{},
	}
}

// RecordMaterialView records first view and updates block progress.
func (s *MaterialProgressService) RecordMaterialView(ctx context.Context, studentID, materialID string, idempotencyKey *string) (MaterialViewResultDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return MaterialViewResultDTO{}, err
	}
	mid, err := domain.ParseMaterialID(materialID)
	if err != nil {
		return MaterialViewResultDTO{}, err
	}

	var result MaterialViewResultDTO
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		visible, err := s.roadmap.IsMaterialVisibleToStudent(ctx, mid)
		if err != nil {
			return err
		}
		if !visible {
			return domain.ErrMaterialNotFound
		}
		blockID, err := s.roadmap.MaterialBlockID(ctx, mid)
		if err != nil {
			return err
		}

		created, err := s.views.RecordFirstView(ctx, domain.MaterialView{
			StudentID:      sid,
			MaterialID:     mid,
			FirstViewedAt:  time.Now().UTC(),
			IdempotencyKey: idempotencyKey,
		})
		if err != nil {
			return err
		}

		key := domain.BlockProgressKey{StudentID: sid, BlockID: blockID}
		p, err := s.progress.Get(ctx, key)
		if err != nil {
			return err
		}
		before := p.Status
		if before == "" {
			before = domain.StatusNotStarted
		}
		now := time.Now().UTC()
		started := before == domain.StatusNotStarted
		p.StudentID = sid
		p.BlockID = blockID
		p.OnMaterialViewed(now)
		if err := persistProgress(ctx, s.progress, &p, before); err != nil {
			return err
		}

		_ = s.events.Record(ctx, domain.EventMaterialViewed, map[string]any{
			"studentId":  studentID,
			"materialId": materialID,
			"blockId":    string(blockID),
		})
		if started && p.Status == domain.StatusInProgress {
			_ = s.events.Record(ctx, domain.EventBlockStarted, map[string]any{
				"studentId": studentID,
				"blockId":   string(blockID),
			})
		}

		result = MaterialViewResultDTO{
			MaterialID: materialID,
			BlockID:    string(blockID),
			Created:    created,
			Status:     string(p.Status),
		}
		return nil
	})
	return result, err
}

// SubmitBlock submits block for approval.
func (s *MaterialProgressService) SubmitBlock(ctx context.Context, studentID, blockID string) (BlockProgressDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return BlockProgressDTO{}, err
	}
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return BlockProgressDTO{}, err
	}

	var dto BlockProgressDTO
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		visible, err := s.roadmap.IsBlockVisibleToStudent(ctx, bid)
		if err != nil {
			return err
		}
		if !visible {
			return domain.ErrBlockNotVisible
		}

		required, err := s.roadmap.ListRequiredMaterialIDs(ctx, bid)
		if err != nil {
			return err
		}
		viewed, err := s.views.CountViewedInSet(ctx, sid, required)
		if err != nil {
			return err
		}
		if viewed < len(required) {
			return domain.ErrRequiredViews
		}

		ordered, err := s.roadmap.ListPublishedBlocksOrdered(ctx)
		if err != nil {
			return err
		}
		all, err := s.progress.ListByStudent(ctx, sid)
		if err != nil {
			return err
		}
		if err := s.sequential.CanSubmit(bid, ordered, progressByBlock(all)); err != nil {
			return err
		}

		key := domain.BlockProgressKey{StudentID: sid, BlockID: bid}
		p, err := s.progress.Get(ctx, key)
		if err != nil {
			return err
		}
		p.StudentID = sid
		p.BlockID = bid
		before := p.Status
		if before == "" {
			before = domain.StatusNotStarted
		}
		now := time.Now().UTC()
		if err := p.Submit(now); err != nil {
			return err
		}
		if err := persistProgress(ctx, s.progress, &p, before); err != nil {
			return err
		}
		_ = s.events.Record(ctx, domain.EventBlockSubmitted, map[string]any{
			"studentId": studentID,
			"blockId":   blockID,
		})
		dto = BlockProgressDTO{BlockID: blockID, Status: string(p.Status), SubmittedAt: formatTimePtr(p.SubmittedAt)}
		return nil
	})
	return dto, err
}
