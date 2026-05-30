package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/oneonone/domain"
)

// OneOnOneAdminService handles admin approve/reject.
type OneOnOneAdminService struct {
	repo   domain.RequestRepository
	bonus  domain.BonusOneOnOnePort
	tx     domain.Transactor
	events domain.EventRecorder
}

// NewOneOnOneAdminService builds OneOnOneAdminService.
func NewOneOnOneAdminService(
	repo domain.RequestRepository,
	bonus domain.BonusOneOnOnePort,
	tx domain.Transactor,
	events domain.EventRecorder,
) *OneOnOneAdminService {
	return &OneOnOneAdminService{repo: repo, bonus: bonus, tx: tx, events: events}
}

// ApproveRequest approves pending request and debits bonus.
func (s *OneOnOneAdminService) ApproveRequest(ctx context.Context, adminID, requestID string) error {
	aid, err := domain.ParseUserID(adminID)
	if err != nil {
		return err
	}
	rid, err := domain.ParseRequestID(requestID)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		req, err := s.repo.GetForUpdate(ctx, rid)
		if err != nil {
			return err
		}
		if req.Status == domain.StatusAccepted {
			return nil
		}
		if req.Status != domain.StatusPending {
			return domain.ErrInvalidTransition
		}
		if err := s.bonus.DebitForRequest(ctx, req.StudentID, req.ID, domain.OneOnOneCostPoints); err != nil {
			return err
		}
		now := time.Now().UTC()
		ref := domain.BonusReference(req.ID)
		if err := req.Approve(aid, now, ref); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, req, domain.StatusPending); err != nil {
			return err
		}
		return s.events.Record(ctx, domain.EventRequestApproved, map[string]any{
			"requestId": requestID, "studentId": string(req.StudentID), "adminId": adminID,
			"bonusReference": ref, "amount": domain.OneOnOneCostPoints,
		})
	})
}

// RejectRequest rejects pending request.
func (s *OneOnOneAdminService) RejectRequest(ctx context.Context, adminID, requestID, reason string) error {
	if _, err := domain.ParseUserID(adminID); err != nil {
		return err
	}
	rid, err := domain.ParseRequestID(requestID)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		req, err := s.repo.GetForUpdate(ctx, rid)
		if err != nil {
			return err
		}
		if req.Status == domain.StatusCancelled && req.RejectReason != nil {
			return nil
		}
		if req.Status != domain.StatusPending {
			return domain.ErrInvalidTransition
		}
		now := time.Now().UTC()
		if err := req.Reject(reason, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, req, domain.StatusPending); err != nil {
			return err
		}
		return s.events.Record(ctx, domain.EventRequestRejected, map[string]any{
			"requestId": requestID, "adminId": adminID, "reason": reason,
		})
	})
}
