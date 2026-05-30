package domain

import (
	"strings"
	"time"
)

// CalendarEvent is the aggregate root.
type CalendarEvent struct {
	ID          EventID
	OrganizerID UserID
	Title       string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	RelatedType RelatedType
	RelatedID   *string
	AttendeeIDs []UserID
	CancelledAt *time.Time
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewEvent creates an active event.
func NewEvent(id EventID, organizer UserID, title, description string, startsAt, endsAt time.Time, related RelatedType, relatedID *string, attendees []UserID, now time.Time) (CalendarEvent, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return CalendarEvent{}, ErrTitleRequired
	}
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen]
	}
	description = trimMax(description, maxDescriptionLen)
	if !endsAt.After(startsAt) {
		return CalendarEvent{}, ErrInvalidTimeRange
	}
	return CalendarEvent{
		ID:          id,
		OrganizerID: organizer,
		Title:       title,
		Description: description,
		StartsAt:    startsAt.UTC(),
		EndsAt:      endsAt.UTC(),
		RelatedType: related,
		RelatedID:   relatedID,
		AttendeeIDs: dedupeUsers(attendees),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update edits mutable fields.
func (e *CalendarEvent) Update(title, description string, startsAt, endsAt time.Time, attendees []UserID, now time.Time) error {
	if e.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	if e.CancelledAt != nil {
		return ErrAlreadyCancelled
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen]
	}
	if !endsAt.After(startsAt) {
		return ErrInvalidTimeRange
	}
	e.Title = title
	e.Description = trimMax(description, maxDescriptionLen)
	e.StartsAt = startsAt.UTC()
	e.EndsAt = endsAt.UTC()
	e.AttendeeIDs = dedupeUsers(attendees)
	e.UpdatedAt = now
	return nil
}

// Cancel marks event cancelled.
func (e *CalendarEvent) Cancel(now time.Time) error {
	if e.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	if e.CancelledAt != nil {
		return ErrAlreadyCancelled
	}
	e.CancelledAt = &now
	e.UpdatedAt = now
	return nil
}

// Delete soft-deletes event.
func (e *CalendarEvent) Delete(now time.Time) error {
	if e.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	e.DeletedAt = &now
	e.UpdatedAt = now
	return nil
}

// IsVisible reports whether event appears in default lists.
func (e *CalendarEvent) IsVisible() bool {
	return e.DeletedAt == nil
}

// CanRead reports read access for user.
func (e *CalendarEvent) CanRead(user UserID, isAdmin bool) bool {
	if !e.IsVisible() && !isAdmin {
		return false
	}
	if isAdmin || e.OrganizerID == user {
		return true
	}
	for _, a := range e.AttendeeIDs {
		if a == user {
			return true
		}
	}
	return false
}

// CanManage reports write access.
func (e *CalendarEvent) CanManage(user UserID, isAdmin bool) bool {
	if e.DeletedAt != nil {
		return isAdmin
	}
	return isAdmin || e.OrganizerID == user
}

func trimMax(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max]
	}
	return s
}

func dedupeUsers(in []UserID) []UserID {
	seen := map[UserID]struct{}{}
	var out []UserID
	for _, u := range in {
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
