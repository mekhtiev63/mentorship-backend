package domain

import (
	"strings"
	"time"
)

// Interview is the aggregate root.
type Interview struct {
	ID                 InterviewID
	Kind               InterviewKind
	StudentID          StudentID
	InterviewerID      *UserID
	Status             InterviewStatus
	Outcome            InterviewOutcome
	ScheduledAt        *time.Time
	Company            string
	Position           string
	StudentNotes       string
	ExternalInterviewer *string
	Feedback           *string
	ReviewedBy         *UserID
	ReviewedAt         *time.Time
	CancelReason       *string
	CatalogPublished   bool
	CompletedAt        *time.Time
	CancelledAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewRealInterview builds a real interview in submitted state.
func NewRealInterview(id InterviewID, student StudentID, company, position string, scheduledAt time.Time, notes string, external *string, now time.Time) (Interview, error) {
	company = strings.TrimSpace(company)
	position = strings.TrimSpace(position)
	if company == "" {
		return Interview{}, ErrCompanyRequired
	}
	if position == "" {
		return Interview{}, ErrPositionRequired
	}
	if scheduledAt.IsZero() {
		return Interview{}, ErrScheduledRequired
	}
	notes = trimMax(notes, maxNotesLen)
	return Interview{
		ID:          id,
		Kind:        KindReal,
		StudentID:   student,
		Status:      StatusSubmitted,
		Outcome:     OutcomePending,
		ScheduledAt: &scheduledAt,
		Company:     company,
		Position:    position,
		StudentNotes: notes,
		ExternalInterviewer: external,
		CatalogPublished:    false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// UpdateReal edits fields while submitted.
func (i *Interview) UpdateReal(company, position string, scheduledAt time.Time, notes string, external *string, now time.Time) error {
	if i.Kind != KindReal || i.Status != StatusSubmitted {
		return ErrInvalidTransition
	}
	company = strings.TrimSpace(company)
	position = strings.TrimSpace(position)
	if company == "" {
		return ErrCompanyRequired
	}
	if position == "" {
		return ErrPositionRequired
	}
	if scheduledAt.IsZero() {
		return ErrScheduledRequired
	}
	i.Company = company
	i.Position = position
	i.ScheduledAt = &scheduledAt
	i.StudentNotes = trimMax(notes, maxNotesLen)
	i.ExternalInterviewer = external
	i.UpdatedAt = now
	return nil
}

// CompleteReal marks real interview completed and publishes to catalog.
func (i *Interview) CompleteReal(outcome InterviewOutcome, now time.Time) error {
	if i.Kind != KindReal || i.Status != StatusSubmitted {
		return ErrInvalidTransition
	}
	if !IsFinalRealOutcome(outcome) {
		return ErrInvalidOutcome
	}
	i.Status = StatusCompleted
	i.Outcome = outcome
	i.CompletedAt = &now
	i.CatalogPublished = true
	i.UpdatedAt = now
	return nil
}

// NewMockInterview builds mock in scheduled state.
func NewMockInterview(id InterviewID, student StudentID, buddy UserID, scheduledAt time.Time, notes string, now time.Time) (Interview, error) {
	if scheduledAt.IsZero() {
		return Interview{}, ErrScheduledRequired
	}
	b := buddy
	return Interview{
		ID:            id,
		Kind:          KindMock,
		StudentID:     student,
		InterviewerID: &b,
		Status:        StatusScheduled,
		Outcome:       OutcomePending,
		ScheduledAt:   &scheduledAt,
		StudentNotes:  trimMax(notes, maxNotesLen),
		CatalogPublished: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// CompleteMock transitions to completed with feedback.
func (i *Interview) CompleteMock(feedback string, outcome InterviewOutcome, now time.Time) error {
	if i.Kind != KindMock || i.Status != StatusScheduled {
		return ErrInvalidTransition
	}
	fb := strings.TrimSpace(feedback)
	if fb == "" {
		return ErrFeedbackRequired
	}
	if len(fb) > maxFeedbackLen {
		fb = fb[:maxFeedbackLen]
	}
	if outcome == "" || outcome == OutcomePending {
		outcome = OutcomeNoResult
	}
	i.Status = StatusCompleted
	i.Outcome = outcome
	i.Feedback = &fb
	i.CompletedAt = &now
	i.CatalogPublished = false
	i.UpdatedAt = now
	return nil
}

func trimMax(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max]
	}
	return s
}
