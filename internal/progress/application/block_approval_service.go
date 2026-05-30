package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/progress/domain"
)

// BlockApprovalService approves or rejects block progress.
type BlockApprovalService struct {
	tx       domain.Transactor
	progress domain.BlockProgressRepository
	buddy    domain.BuddyScopePort
	events   domain.EventRecorder
}

// NewBlockApprovalService builds BlockApprovalService.
func NewBlockApprovalService(
	tx domain.Transactor,
	progress domain.BlockProgressRepository,
	buddy domain.BuddyScopePort,
	events domain.EventRecorder,
) *BlockApprovalService {
	return &BlockApprovalService{tx: tx, progress: progress, buddy: buddy, events: events}
}

// ApproveBlockAsBuddy approves when buddy is assigned.
func (s *BlockApprovalService) ApproveBlockAsBuddy(ctx context.Context, buddyID, studentID, blockID string) error {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, sid)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return s.approve(ctx, bid, sid, blockID, "buddy")
}

// RejectBlockAsBuddy rejects with reason.
func (s *BlockApprovalService) RejectBlockAsBuddy(ctx context.Context, buddyID, studentID, blockID, reason string) error {
	bid, err := domain.ParseUserID(buddyID)
	if err != nil {
		return err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	ok, err := s.buddy.IsActiveBuddyOf(ctx, bid, sid)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return s.reject(ctx, bid, sid, blockID, reason, "buddy")
}

// ApproveBlockAsAdmin approves without buddy check.
func (s *BlockApprovalService) ApproveBlockAsAdmin(ctx context.Context, adminID, studentID, blockID string) error {
	aid, err := domain.ParseUserID(adminID)
	if err != nil {
		return err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	return s.approve(ctx, aid, sid, blockID, "admin")
}

// RejectBlockAsAdmin rejects as admin.
func (s *BlockApprovalService) RejectBlockAsAdmin(ctx context.Context, adminID, studentID, blockID, reason string) error {
	aid, err := domain.ParseUserID(adminID)
	if err != nil {
		return err
	}
	sid, err := domain.ParseStudentID(studentID)
	if err != nil {
		return err
	}
	return s.reject(ctx, aid, sid, blockID, reason, "admin")
}

func (s *BlockApprovalService) approve(ctx context.Context, approver domain.UserID, studentID domain.StudentID, blockID, role string) error {
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		key := domain.BlockProgressKey{StudentID: studentID, BlockID: bid}
		p, err := s.progress.Get(ctx, key)
		if err != nil {
			return err
		}
		if !p.Exists {
			return domain.ErrNotFound
		}
		before := p.Status
		now := time.Now().UTC()
		if err := p.Approve(approver, now); err != nil {
			return err
		}
		if err := persistProgress(ctx, s.progress, &p, before); err != nil {
			return err
		}
		_ = s.events.Record(ctx, domain.EventBlockApproved, map[string]any{
			"studentId":  string(studentID),
			"blockId":    blockID,
			"approvedBy": string(approver),
			"role":       role,
		})
		return nil
	})
}

func (s *BlockApprovalService) reject(ctx context.Context, approver domain.UserID, studentID domain.StudentID, blockID, reason, role string) error {
	bid, err := domain.ParseBlockID(blockID)
	if err != nil {
		return err
	}
	rr, err := domain.ParseRejectReason(reason)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		key := domain.BlockProgressKey{StudentID: studentID, BlockID: bid}
		p, err := s.progress.Get(ctx, key)
		if err != nil {
			return err
		}
		if !p.Exists {
			return domain.ErrNotFound
		}
		before := p.Status
		now := time.Now().UTC()
		if err := p.Reject(approver, rr, now); err != nil {
			return err
		}
		if err := persistProgress(ctx, s.progress, &p, before); err != nil {
			return err
		}
		_ = s.events.Record(ctx, domain.EventBlockRejected, map[string]any{
			"studentId":  string(studentID),
			"blockId":    blockID,
			"approvedBy": string(approver),
			"role":       role,
			"reason":     string(rr),
		})
		return nil
	})
}
