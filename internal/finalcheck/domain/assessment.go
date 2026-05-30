package domain

import (
	"strings"
	"time"
)

// CheckTrack is one final check leg.
type CheckTrack struct {
	Status      TrackStatus
	ReviewerID  *UserID
	ScheduledAt *time.Time
	Feedback    *string
	CompletedAt *time.Time
	FailedAt    *time.Time
	FailReason  *string
}

// FinalAssessment is the aggregate root (one per student).
type FinalAssessment struct {
	ID                   AssessmentID
	StudentID            StudentID
	Tech                 CheckTrack
	Roast                CheckTrack
	FinalistEventEmitted bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewEligibleAssessment creates assessment with tech available.
func NewEligibleAssessment(id AssessmentID, student StudentID, now time.Time) FinalAssessment {
	return FinalAssessment{
		ID:        id,
		StudentID: student,
		Tech:      CheckTrack{Status: StatusAvailable},
		Roast:     CheckTrack{Status: StatusNotAvailable},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// OpenTechAvailability sets tech to available when program completed.
func (a *FinalAssessment) OpenTechAvailability(now time.Time) {
	if a.Tech.Status == StatusNotAvailable {
		a.Tech.Status = StatusAvailable
		a.UpdatedAt = now
	}
}

func (a *FinalAssessment) track(kind CheckKind) *CheckTrack {
	if kind == CheckRoast {
		return &a.Roast
	}
	return &a.Tech
}

// Schedule assigns reviewer and scheduled time.
func (a *FinalAssessment) Schedule(kind CheckKind, reviewer UserID, scheduledAt time.Time, now time.Time) error {
	if scheduledAt.IsZero() {
		return ErrScheduledRequired
	}
	t := a.track(kind)
	if t.Status.isTerminal() {
		return ErrRetryNotAllowed
	}
	if t.Status != StatusAvailable {
		return ErrInvalidTransition
	}
	if kind == CheckRoast && a.Tech.Status != StatusCompleted {
		return ErrInvalidTransition
	}
	t.Status = StatusScheduled
	t.ReviewerID = &reviewer
	t.ScheduledAt = &scheduledAt
	a.UpdatedAt = now
	return nil
}

// Complete confirms track success.
func (a *FinalAssessment) Complete(kind CheckKind, feedback string, now time.Time) error {
	fb := strings.TrimSpace(feedback)
	if fb == "" {
		return ErrFeedbackRequired
	}
	if len(fb) > maxFeedbackLen {
		fb = fb[:maxFeedbackLen]
	}
	t := a.track(kind)
	if t.Status.isTerminal() {
		return ErrRetryNotAllowed
	}
	if t.Status != StatusScheduled {
		return ErrInvalidTransition
	}
	t.Status = StatusCompleted
	t.Feedback = &fb
	t.CompletedAt = &now
	a.UpdatedAt = now
	if kind == CheckTech && a.Roast.Status == StatusNotAvailable {
		a.Roast.Status = StatusAvailable
	}
	return nil
}

// Fail marks track failed (terminal).
func (a *FinalAssessment) Fail(kind CheckKind, reason string, now time.Time) error {
	rs := strings.TrimSpace(reason)
	if rs == "" {
		return ErrReasonRequired
	}
	if len(rs) > maxReasonLen {
		rs = rs[:maxReasonLen]
	}
	t := a.track(kind)
	if t.Status.isTerminal() {
		return ErrRetryNotAllowed
	}
	if t.Status != StatusScheduled {
		return ErrInvalidTransition
	}
	t.Status = StatusFailed
	t.FailReason = &rs
	t.FailedAt = &now
	a.UpdatedAt = now
	return nil
}

// BothCompleted reports whether both tracks completed.
func (a *FinalAssessment) BothCompleted() bool {
	return a.Tech.Status == StatusCompleted && a.Roast.Status == StatusCompleted
}
