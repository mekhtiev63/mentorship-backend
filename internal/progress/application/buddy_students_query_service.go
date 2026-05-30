package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// BuddyStudentsQueryService reads buddy-facing progress data.
type BuddyStudentsQueryService struct {
	progress      domain.BlockProgressRepository
	buddy         domain.BuddyScopePort
	studentQuery  *StudentProgressQueryService
}

// NewBuddyStudentsQueryService builds BuddyStudentsQueryService.
func NewBuddyStudentsQueryService(
	progress domain.BlockProgressRepository,
	buddy domain.BuddyScopePort,
	studentQuery *StudentProgressQueryService,
) *BuddyStudentsQueryService {
	return &BuddyStudentsQueryService{
		progress:     progress,
		buddy:        buddy,
		studentQuery: studentQuery,
	}
}

// ListStudentsWithProgress lists assigned students with status counts.
func (s *BuddyStudentsQueryService) ListStudentsWithProgress(ctx context.Context, buddyID string, page, pageSize int) ([]StudentProgressItemDTO, pagination.Meta, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	params := pagination.Normalize(page, pageSize)
	studentIDs, total, err := s.buddy.ListActiveStudentIDsForBuddy(ctx, bid, params)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	out := make([]StudentProgressItemDTO, 0, len(studentIDs))
	for _, sid := range studentIDs {
		rows, err := s.progress.ListByStudent(ctx, sid)
		if err != nil {
			return nil, pagination.Meta{}, err
		}
		item := StudentProgressItemDTO{StudentID: string(sid)}
		for _, p := range rows {
			switch p.Status {
			case domain.StatusAwaitingApproval:
				item.AwaitingCount++
			case domain.StatusApproved:
				item.ApprovedCount++
			case domain.StatusInProgress, domain.StatusRejected:
				item.InProgressCount++
			}
		}
		out = append(out, item)
	}
	return out, pagination.NewMeta(params.Page, params.PageSize, total), nil
}

// GetStudentProgress returns progress for assigned student.
func (s *BuddyStudentsQueryService) GetStudentProgress(ctx context.Context, buddyID, studentID string) ([]BlockProgressDTO, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return nil, err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return nil, err
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, sid)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}
	return s.studentQuery.ListMyBlocks(ctx, studentID)
}

// ListApprovalQueue returns awaiting items for buddy.
func (s *BuddyStudentsQueryService) ListApprovalQueue(ctx context.Context, buddyID string, page, pageSize int) ([]ApprovalQueueItemDTO, pagination.Meta, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	params := pagination.Normalize(page, pageSize)
	rows, total, err := s.progress.ListAwaitingForBuddy(ctx, bid, params)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	out := make([]ApprovalQueueItemDTO, 0, len(rows))
	for _, p := range rows {
		submitted := ""
		if p.SubmittedAt != nil {
			submitted = p.SubmittedAt.UTC().Format(time.RFC3339)
		}
		out = append(out, ApprovalQueueItemDTO{
			StudentID:   string(p.StudentID),
			BlockID:     string(p.BlockID),
			SubmittedAt: submitted,
		})
	}
	return out, pagination.NewMeta(params.Page, params.PageSize, total), nil
}
