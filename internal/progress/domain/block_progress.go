package domain

import "time"

// BlockProgress is the aggregate root for student progress on a block.
type BlockProgress struct {
	StudentID    StudentID
	BlockID      BlockID
	Status       ProgressStatus
	SubmittedAt  *time.Time
	ApprovedBy   *UserID
	ApprovedAt   *time.Time
	RejectedAt   *time.Time
	RejectReason *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Exists       bool
}

// Key returns aggregate identity.
func (p BlockProgress) Key() BlockProgressKey {
	return BlockProgressKey{StudentID: p.StudentID, BlockID: p.BlockID}
}

// OnMaterialViewed transitions progress when student views a material.
func (p *BlockProgress) OnMaterialViewed(now time.Time) {
	switch p.Status {
	case StatusNotStarted, "":
		p.Status = StatusInProgress
		if !p.Exists {
			p.CreatedAt = now
		}
		p.UpdatedAt = now
	case StatusRejected:
		p.clearDecision()
		p.Status = StatusInProgress
		p.UpdatedAt = now
	case StatusInProgress:
		p.UpdatedAt = now
	}
}

// Submit moves block to awaiting approval.
func (p *BlockProgress) Submit(now time.Time) error {
	switch p.Status {
	case StatusInProgress, StatusRejected:
		if p.Status == StatusRejected {
			p.clearDecision()
		}
		p.Status = StatusAwaitingApproval
		p.SubmittedAt = &now
		p.UpdatedAt = now
		return nil
	default:
		return ErrInvalidTransition
	}
}

// Approve marks block approved.
func (p *BlockProgress) Approve(approver UserID, now time.Time) error {
	if p.Status != StatusAwaitingApproval {
		return ErrInvalidTransition
	}
	p.Status = StatusApproved
	p.ApprovedBy = &approver
	p.ApprovedAt = &now
	p.RejectedAt = nil
	p.RejectReason = nil
	p.UpdatedAt = now
	return nil
}

// Reject marks block rejected.
func (p *BlockProgress) Reject(approver UserID, reason RejectReason, now time.Time) error {
	if p.Status != StatusAwaitingApproval {
		return ErrInvalidTransition
	}
	rs := string(reason)
	p.Status = StatusRejected
	p.ApprovedBy = &approver
	p.ApprovedAt = &now
	p.RejectedAt = &now
	p.RejectReason = &rs
	p.UpdatedAt = now
	return nil
}

func (p *BlockProgress) clearDecision() {
	p.ApprovedBy = nil
	p.ApprovedAt = nil
	p.RejectedAt = nil
	p.RejectReason = nil
	p.SubmittedAt = nil
}
