package application

import (
	"context"

	"github.com/go-mentorship-platform/backend/internal/user/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// BuddyService serves buddy-facing queries.
type BuddyService struct {
	assignments domain.BuddyAssignmentRepository
}

// NewBuddyService builds BuddyService.
func NewBuddyService(assignments domain.BuddyAssignmentRepository) *BuddyService {
	return &BuddyService{assignments: assignments}
}

// StudentDTO is a student list item.
type StudentDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ListStudents returns students assigned to buddy.
func (s *BuddyService) ListStudents(ctx context.Context, buddyID string, page, pageSize int) ([]StudentDTO, pagination.Meta, error) {
	id, err := domain.ParseUserID(buddyID)
	if err != nil {
		return nil, pagination.Meta{}, domain.ErrForbidden
	}
	params := pagination.Normalize(page, pageSize)
	students, total, err := s.assignments.ListActiveStudentsForBuddy(ctx, id, params)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	out := make([]StudentDTO, len(students))
	for i, st := range students {
		out[i] = StudentDTO{ID: string(st.ID), Email: string(st.Email)}
	}
	return out, pagination.NewMeta(params.Page, params.PageSize, total), nil
}

// EnsureAssigned returns forbidden if buddy is not assigned to student.
func (s *BuddyService) EnsureAssigned(ctx context.Context, buddyID, studentID string) error {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return domain.ErrForbidden
	}
	sid, err := domain.ParseUserID(studentID)
	if err != nil {
		return domain.ErrForbidden
	}
	ok, err := s.assignments.IsAssignedBuddy(ctx, bid, sid)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}
