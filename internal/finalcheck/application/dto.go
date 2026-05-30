package application

import (
	"time"

	"github.com/go-mentorship-platform/backend/internal/finalcheck/domain"
)

// TrackDTO is API view of one check leg.
type TrackDTO struct {
	Status      string     `json:"status"`
	ReviewerID  *string    `json:"reviewer_id,omitempty"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Feedback    *string    `json:"feedback,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	FailReason  *string    `json:"fail_reason,omitempty"`
}

// FinalCheckDTO is full assessment status.
type FinalCheckDTO struct {
	AssessmentID         string   `json:"assessment_id,omitempty"`
	StudentID            string   `json:"student_id"`
	ProgramCompleted     bool     `json:"program_completed"`
	Tech                 TrackDTO `json:"tech"`
	Roast                TrackDTO `json:"roast"`
	FinalistEventEmitted bool     `json:"finalist_event_emitted"`
}

func toTrackDTO(t domain.CheckTrack) TrackDTO {
	d := TrackDTO{Status: string(t.Status)}
	if t.ReviewerID != nil {
		s := string(*t.ReviewerID)
		d.ReviewerID = &s
	}
	d.ScheduledAt = t.ScheduledAt
	d.Feedback = t.Feedback
	d.CompletedAt = t.CompletedAt
	d.FailedAt = t.FailedAt
	d.FailReason = t.FailReason
	return d
}

func toDTO(a domain.FinalAssessment, programCompleted bool) FinalCheckDTO {
	return FinalCheckDTO{
		AssessmentID:         string(a.ID),
		StudentID:            string(a.StudentID),
		ProgramCompleted:     programCompleted,
		Tech:                 toTrackDTO(a.Tech),
		Roast:                toTrackDTO(a.Roast),
		FinalistEventEmitted: a.FinalistEventEmitted,
	}
}

func syntheticNotAvailable(student domain.StudentID, programCompleted bool) FinalCheckDTO {
	return FinalCheckDTO{
		StudentID:        string(student),
		ProgramCompleted: programCompleted,
		Tech:             TrackDTO{Status: string(domain.StatusNotAvailable)},
		Roast:            TrackDTO{Status: string(domain.StatusNotAvailable)},
	}
}
