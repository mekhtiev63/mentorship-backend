package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
)

// MockInterviewService handles buddy mock interviews.
type MockInterviewService struct {
	repo   domain.InterviewRepository
	buddy  domain.BuddyScopePort
	events domain.EventRecorder
}

// NewMockInterviewService builds MockInterviewService.
func NewMockInterviewService(repo domain.InterviewRepository, buddy domain.BuddyScopePort, events domain.EventRecorder) *MockInterviewService {
	return &MockInterviewService{repo: repo, buddy: buddy, events: events}
}

// Create schedules mock for assigned student.
func (s *MockInterviewService) Create(ctx context.Context, buddyID string, in MockCreateInput) (InterviewDTO, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return InterviewDTO{}, err
	}
	sid, err := domain.ParseStudentID(in.StudentID)
	if err != nil {
		return InterviewDTO{}, err
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, sid)
	if err != nil {
		return InterviewDTO{}, err
	}
	if !ok {
		return InterviewDTO{}, domain.ErrForbidden
	}
	now := time.Now().UTC()
	id := domain.InterviewID(uuid.NewString())
	interview, err := domain.NewMockInterview(id, sid, bid, in.ScheduledAt.UTC(), in.StudentNotes, now)
	if err != nil {
		return InterviewDTO{}, err
	}
	if err := s.repo.Insert(ctx, interview); err != nil {
		return InterviewDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventMockScheduled, map[string]any{
		"interviewId": string(id), "studentId": in.StudentID, "buddyId": buddyID,
	})
	return toDTO(interview), nil
}

// List lists mock interviews for buddy.
func (s *MockInterviewService) List(ctx context.Context, buddyID string, page, pageSize int, status *string) (ListResult, error) {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(page, pageSize)
	var st *domain.InterviewStatus
	if status != nil && *status != "" {
		sv := domain.InterviewStatus(*status)
		st = &sv
	}
	items, total, err := s.repo.ListByInterviewer(ctx, bid, domain.KindMock, st, p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}
