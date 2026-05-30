package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// InterviewQueryService reads interviews and catalog.
type InterviewQueryService struct {
	repo domain.InterviewRepository
}

// NewInterviewQueryService builds InterviewQueryService.
func NewInterviewQueryService(repo domain.InterviewRepository) *InterviewQueryService {
	return &InterviewQueryService{repo: repo}
}

// GetRealForStudent returns student's real interview.
func (s *InterviewQueryService) GetRealForStudent(ctx context.Context, studentID, interviewID string) (InterviewDTO, error) {
	return s.getForStudent(ctx, studentID, interviewID, domain.KindReal)
}

// ListRealForStudent lists student's real interviews.
func (s *InterviewQueryService) ListRealForStudent(ctx context.Context, studentID string, page, pageSize int, status *string) (ListResult, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(page, pageSize)
	var st *domain.InterviewStatus
	if status != nil && *status != "" {
		sv := domain.InterviewStatus(*status)
		st = &sv
	}
	items, total, err := s.repo.ListByStudent(ctx, sid, domain.KindReal, st, p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}

// GetMockForStudent returns student's mock interview.
func (s *InterviewQueryService) GetMockForStudent(ctx context.Context, studentID, interviewID string) (InterviewDTO, error) {
	return s.getForStudent(ctx, studentID, interviewID, domain.KindMock)
}

// ListMockForStudent lists student's mock interviews.
func (s *InterviewQueryService) ListMockForStudent(ctx context.Context, studentID string, page, pageSize int, status *string) (ListResult, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(page, pageSize)
	var st *domain.InterviewStatus
	if status != nil && *status != "" {
		sv := domain.InterviewStatus(*status)
		st = &sv
	}
	items, total, err := s.repo.ListByStudent(ctx, sid, domain.KindMock, st, p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}

// ListCatalog lists public real interviews (no mock).
func (s *InterviewQueryService) ListCatalog(ctx context.Context, page, pageSize int, company *string, outcome *string) (CatalogListResult, error) {
	p := pagination.Normalize(page, pageSize)
	var oc *domain.InterviewOutcome
	if outcome != nil && *outcome != "" {
		o := domain.InterviewOutcome(*outcome)
		oc = &o
	}
	items, total, err := s.repo.ListCatalog(ctx, p, company, oc)
	if err != nil {
		return CatalogListResult{}, err
	}
	return catalogListDTO(items, p, total), nil
}

func (s *InterviewQueryService) getForStudent(ctx context.Context, studentID, interviewID string, kind domain.InterviewKind) (InterviewDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return InterviewDTO{}, err
	}
	iid, err := domain.ParseInterviewID(interviewID)
	if err != nil {
		return InterviewDTO{}, err
	}
	interview, err := s.repo.GetByID(ctx, iid)
	if err != nil {
		return InterviewDTO{}, err
	}
	if interview.StudentID != sid || interview.Kind != kind {
		return InterviewDTO{}, domain.ErrForbidden
	}
	return toDTO(interview), nil
}
