package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// OneOnOneQueryService reads requests.
type OneOnOneQueryService struct {
	repo domain.RequestRepository
}

// NewOneOnOneQueryService builds OneOnOneQueryService.
func NewOneOnOneQueryService(repo domain.RequestRepository) *OneOnOneQueryService {
	return &OneOnOneQueryService{repo: repo}
}

// GetStudentRequest returns request for student if owned.
func (s *OneOnOneQueryService) GetStudentRequest(ctx context.Context, studentID, requestID string) (RequestDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return RequestDTO{}, err
	}
	rid, err := domain.ParseRequestID(requestID)
	if err != nil {
		return RequestDTO{}, err
	}
	req, err := s.repo.GetByID(ctx, rid)
	if err != nil {
		return RequestDTO{}, err
	}
	if req.StudentID != sid {
		return RequestDTO{}, domain.ErrForbidden
	}
	return toDTO(req), nil
}

// ListStudentRequests lists student's requests.
func (s *OneOnOneQueryService) ListStudentRequests(ctx context.Context, studentID string, page, pageSize int) (ListResult, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(page, pageSize)
	items, total, err := s.repo.ListByStudent(ctx, sid, p)
	if err != nil {
		return ListResult{}, err
	}
	return listToResult(items, p, total), nil
}

// GetAdminRequest returns any request for admin.
func (s *OneOnOneQueryService) GetAdminRequest(ctx context.Context, requestID string) (RequestDTO, error) {
	rid, err := domain.ParseRequestID(requestID)
	if err != nil {
		return RequestDTO{}, err
	}
	req, err := s.repo.GetByID(ctx, rid)
	if err != nil {
		return RequestDTO{}, err
	}
	return toDTO(req), nil
}

// ListAdminRequests lists all requests with optional status filter.
func (s *OneOnOneQueryService) ListAdminRequests(ctx context.Context, page, pageSize int, status *string) (ListResult, error) {
	p := pagination.Normalize(page, pageSize)
	var st *domain.RequestStatus
	if status != nil && *status != "" {
		sv := domain.RequestStatus(*status)
		st = &sv
	}
	items, total, err := s.repo.ListAll(ctx, p, st)
	if err != nil {
		return ListResult{}, err
	}
	return listToResult(items, p, total), nil
}

func listToResult(items []domain.OneOnOneRequest, p pagination.Params, total int64) ListResult {
	dtos := make([]RequestDTO, 0, len(items))
	for _, req := range items {
		dtos = append(dtos, toDTO(req))
	}
	return ListResult{
		Items: dtos,
		Meta:  pagination.NewMeta(p.Page, p.PageSize, total),
	}
}
