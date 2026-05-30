package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/activity/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// ActivityQueryService reads activity feeds.
type ActivityQueryService struct {
	journal domain.ActivityJournalRepository
	buddy   domain.BuddyScopePort
}

// NewActivityQueryService builds ActivityQueryService.
func NewActivityQueryService(journal domain.ActivityJournalRepository, buddy domain.BuddyScopePort) *ActivityQueryService {
	return &ActivityQueryService{journal: journal, buddy: buddy}
}

// ListMyActivity returns student feed.
func (s *ActivityQueryService) ListMyActivity(ctx context.Context, userID string, q ListQuery) (ListResult, error) {
	uid, err := domain.ParseUserID(userID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(q.Page, q.PageSize)
	items, total, err := s.journal.ListBySubject(ctx, uid, toFilter(q), p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}

// ListStudentActivity returns buddy feed for assigned student.
func (s *ActivityQueryService) ListStudentActivity(ctx context.Context, buddyID, studentID string, q ListQuery) (ListResult, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return ListResult{}, domain.ErrForbidden
	}
	sid, err := domain.ParseUserID(studentID)
	if err != nil {
		return ListResult{}, domain.ErrForbidden
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, sid)
	if err != nil {
		return ListResult{}, err
	}
	if !ok {
		return ListResult{}, domain.ErrForbidden
	}
	p := pagination.Normalize(q.Page, q.PageSize)
	items, total, err := s.journal.ListBySubject(ctx, sid, toFilter(q), p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}

// ListGlobal returns admin feed.
func (s *ActivityQueryService) ListGlobal(ctx context.Context, q ListQuery) (ListResult, error) {
	p := pagination.Normalize(q.Page, q.PageSize)
	items, total, err := s.journal.ListAll(ctx, toFilter(q), p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}
