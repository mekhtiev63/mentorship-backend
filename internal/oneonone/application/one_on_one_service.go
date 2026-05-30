package application

import (
	"context"
	"strings"
	"time"

	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
	"github.com/google/uuid"
)

// OneOnOneService handles student commands.
type OneOnOneService struct {
	repo   domain.RequestRepository
	buddy  domain.BuddyAssignmentPort
	bonus  domain.BonusOneOnOnePort
	events domain.EventRecorder
}

// NewOneOnOneService builds OneOnOneService.
func NewOneOnOneService(
	repo domain.RequestRepository,
	buddy domain.BuddyAssignmentPort,
	bonus domain.BonusOneOnOnePort,
	events domain.EventRecorder,
) *OneOnOneService {
	return &OneOnOneService{repo: repo, buddy: buddy, bonus: bonus, events: events}
}

// CreateRequestInput is create payload.
type CreateRequestInput struct {
	Message        string
	PreferredSlots []byte
}

// CreateRequest creates a pending 1:1 request.
func (s *OneOnOneService) CreateRequest(ctx context.Context, studentID string, in CreateRequestInput) (RequestDTO, error) {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return RequestDTO{}, err
	}
	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		return RequestDTO{}, domain.ErrInvalidMessage
	}
	slots, err := domain.PreferredSlotsJSON(in.PreferredSlots)
	if err != nil {
		return RequestDTO{}, err
	}
	ok, err := s.bonus.HasSufficientBalance(ctx, sid, domain.OneOnOneCostPoints)
	if err != nil {
		return RequestDTO{}, err
	}
	if !ok {
		return RequestDTO{}, domain.ErrInsufficientBonus
	}
	buddyID, err := s.buddy.GetActiveBuddyID(ctx, sid)
	if err != nil {
		return RequestDTO{}, err
	}
	now := time.Now().UTC()
	req := domain.OneOnOneRequest{
		ID:             domain.RequestID(uuid.NewString()),
		StudentID:      sid,
		BuddyID:        buddyID,
		Status:         domain.StatusPending,
		Message:        msg,
		PreferredSlots: slots,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.Insert(ctx, req); err != nil {
		return RequestDTO{}, err
	}
	_ = s.events.Record(ctx, domain.EventRequestCreated, map[string]any{
		"requestId": string(req.ID), "studentId": studentID, "buddyId": string(buddyID),
	})
	return toDTO(req), nil
}

// CancelRequest cancels own pending request.
func (s *OneOnOneService) CancelRequest(ctx context.Context, studentID, requestID string) error {
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	rid, err := domain.ParseRequestID(requestID)
	if err != nil {
		return err
	}
	req, err := s.repo.GetByID(ctx, rid)
	if err != nil {
		return err
	}
	if req.StudentID != sid {
		return domain.ErrForbidden
	}
	now := time.Now().UTC()
	if err := req.Cancel(now); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, req, domain.StatusPending); err != nil {
		return err
	}
	_ = s.events.Record(ctx, domain.EventRequestCancelled, map[string]any{
		"requestId": requestID, "studentId": studentID,
	})
	return nil
}
