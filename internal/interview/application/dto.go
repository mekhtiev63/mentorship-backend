package application

import (
	"time"

	"github.com/go-mentorship-platform/backend/internal/interview/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// InterviewDTO is API representation.
type InterviewDTO struct {
	ID                  string     `json:"id"`
	Kind                string     `json:"kind"`
	StudentID           string     `json:"student_id,omitempty"`
	InterviewerID       *string    `json:"interviewer_id,omitempty"`
	Status              string     `json:"status"`
	Outcome             string     `json:"outcome"`
	ScheduledAt         *time.Time `json:"scheduled_at,omitempty"`
	Company             string     `json:"company,omitempty"`
	Position            string     `json:"position,omitempty"`
	StudentNotes        string     `json:"student_notes,omitempty"`
	ExternalInterviewer *string    `json:"external_interviewer,omitempty"`
	Feedback            *string    `json:"feedback,omitempty"`
	CatalogPublished    bool       `json:"catalog_published"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CatalogInterviewDTO hides student identity.
type CatalogInterviewDTO struct {
	ID                  string          `json:"id"`
	Company             string          `json:"company"`
	Position            string          `json:"position"`
	Outcome             string          `json:"outcome"`
	ScheduledAt         *time.Time      `json:"scheduled_at,omitempty"`
	StudentNotes        string          `json:"student_notes,omitempty"`
	ExternalInterviewer *string         `json:"external_interviewer,omitempty"`
	CompletedAt         *time.Time      `json:"completed_at,omitempty"`
}

// ListResult paginated interviews.
type ListResult struct {
	Items []InterviewDTO  `json:"items"`
	Meta  pagination.Meta `json:"meta"`
}

// CatalogListResult paginated catalog.
type CatalogListResult struct {
	Items []CatalogInterviewDTO `json:"items"`
	Meta  pagination.Meta       `json:"meta"`
}

func toDTO(i domain.Interview) InterviewDTO {
	d := InterviewDTO{
		ID:                  string(i.ID),
		Kind:                string(i.Kind),
		StudentID:           string(i.StudentID),
		Status:              string(i.Status),
		Outcome:             string(i.Outcome),
		ScheduledAt:         i.ScheduledAt,
		Company:             i.Company,
		Position:            i.Position,
		StudentNotes:        i.StudentNotes,
		ExternalInterviewer: i.ExternalInterviewer,
		Feedback:            i.Feedback,
		CatalogPublished:    i.CatalogPublished,
		CompletedAt:         i.CompletedAt,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
	if i.InterviewerID != nil {
		s := string(*i.InterviewerID)
		d.InterviewerID = &s
	}
	return d
}

func toCatalogDTO(i domain.Interview) CatalogInterviewDTO {
	return CatalogInterviewDTO{
		ID:                  string(i.ID),
		Company:             i.Company,
		Position:            i.Position,
		Outcome:             string(i.Outcome),
		ScheduledAt:         i.ScheduledAt,
		StudentNotes:        i.StudentNotes,
		ExternalInterviewer: i.ExternalInterviewer,
		CompletedAt:         i.CompletedAt,
	}
}

func listDTO(items []domain.Interview, p pagination.Params, total int64) ListResult {
	dtos := make([]InterviewDTO, 0, len(items))
	for _, i := range items {
		dtos = append(dtos, toDTO(i))
	}
	return ListResult{Items: dtos, Meta: pagination.NewMeta(p.Page, p.PageSize, total)}
}

func catalogListDTO(items []domain.Interview, p pagination.Params, total int64) CatalogListResult {
	dtos := make([]CatalogInterviewDTO, 0, len(items))
	for _, i := range items {
		dtos = append(dtos, toCatalogDTO(i))
	}
	return CatalogListResult{Items: dtos, Meta: pagination.NewMeta(p.Page, p.PageSize, total)}
}

// RealCreateInput create real interview.
type RealCreateInput struct {
	Company             string
	Position            string
	ScheduledAt         time.Time
	StudentNotes        string
	ExternalInterviewer *string
}

// RealUpdateInput update real interview.
type RealUpdateInput struct {
	Company             string
	Position            string
	ScheduledAt         time.Time
	StudentNotes        string
	ExternalInterviewer *string
}

// MockCreateInput create mock interview.
type MockCreateInput struct {
	StudentID    string
	ScheduledAt  time.Time
	StudentNotes string
}

// FeedbackInput complete mock with feedback.
type FeedbackInput struct {
	Feedback string
	Outcome  domain.InterviewOutcome
}
