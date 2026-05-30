package domain

import (
	"strings"
	"time"
)

// OneOnOneRequest is the aggregate root.
type OneOnOneRequest struct {
	ID              RequestID
	StudentID       StudentID
	BuddyID         BuddyID
	Status          RequestStatus
	Message         string
	PreferredSlots  []byte
	CalendarEventID *string
	RejectReason    *string
	ApprovedBy      *UserID
	ApprovedAt      *time.Time
	BonusDebitedAt  *time.Time
	BonusReference  *string
	CancelledAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Approve transitions pending to accepted after bonus debit.
func (r *OneOnOneRequest) Approve(admin UserID, now time.Time, bonusRef string) error {
	if r.Status != StatusPending {
		return ErrInvalidTransition
	}
	if r.BonusDebitedAt != nil {
		return ErrInvalidTransition
	}
	r.Status = StatusAccepted
	r.ApprovedBy = &admin
	r.ApprovedAt = &now
	r.BonusDebitedAt = &now
	ref := bonusRef
	r.BonusReference = &ref
	r.UpdatedAt = now
	return nil
}

// Reject transitions pending to cancelled.
func (r *OneOnOneRequest) Reject(reason string, now time.Time) error {
	if r.Status != StatusPending {
		return ErrInvalidTransition
	}
	rs := strings.TrimSpace(reason)
	if rs == "" {
		return ErrRejectReason
	}
	r.Status = StatusCancelled
	r.RejectReason = &rs
	r.CancelledAt = &now
	r.UpdatedAt = now
	return nil
}

// Cancel transitions pending to cancelled by student.
func (r *OneOnOneRequest) Cancel(now time.Time) error {
	if r.Status != StatusPending {
		return ErrInvalidTransition
	}
	r.Status = StatusCancelled
	r.CancelledAt = &now
	r.UpdatedAt = now
	return nil
}
