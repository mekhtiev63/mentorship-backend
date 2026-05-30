package application

import (
	"time"

	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// EventDTO is API representation.
type EventDTO struct {
	ID          string   `json:"id"`
	OrganizerID string   `json:"organizer_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	RelatedType string   `json:"related_type"`
	RelatedID   *string  `json:"related_id,omitempty"`
	AttendeeIDs []string `json:"attendee_ids"`
	Cancelled   bool     `json:"cancelled"`
	Deleted     bool     `json:"deleted"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListResult paginated events.
type ListResult struct {
	Items []EventDTO      `json:"items"`
	Meta  pagination.Meta `json:"meta"`
}

// CreateEventInput create payload.
type CreateEventInput struct {
	Title       string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	RelatedType domain.RelatedType
	RelatedID   *string
	AttendeeIDs []string
}

// UpdateEventInput update payload.
type UpdateEventInput struct {
	Title       string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	AttendeeIDs []string
}

// ListEventsQuery list filters.
type ListEventsQuery struct {
	From             *time.Time
	To               *time.Time
	RelatedType      *string
	IncludeCancelled bool
	IncludeDeleted   bool
	Page             int
	PageSize         int
}

func toDTO(e domain.CalendarEvent) EventDTO {
	ids := make([]string, 0, len(e.AttendeeIDs))
	for _, a := range e.AttendeeIDs {
		ids = append(ids, string(a))
	}
	return EventDTO{
		ID:          string(e.ID),
		OrganizerID: string(e.OrganizerID),
		Title:       e.Title,
		Description: e.Description,
		StartsAt:    e.StartsAt,
		EndsAt:      e.EndsAt,
		RelatedType: string(e.RelatedType),
		RelatedID:   e.RelatedID,
		AttendeeIDs: ids,
		Cancelled:   e.CancelledAt != nil,
		Deleted:     e.DeletedAt != nil,
		CancelledAt: e.CancelledAt,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func listDTO(items []domain.CalendarEvent, p pagination.Params, total int64) ListResult {
	dtos := make([]EventDTO, 0, len(items))
	for _, e := range items {
		dtos = append(dtos, toDTO(e))
	}
	return ListResult{Items: dtos, Meta: pagination.NewMeta(p.Page, p.PageSize, total)}
}

func parseAttendees(raw []string) ([]domain.UserID, error) {
	out := make([]domain.UserID, 0, len(raw))
	for _, s := range raw {
		u, err := domain.ParseUserID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}
