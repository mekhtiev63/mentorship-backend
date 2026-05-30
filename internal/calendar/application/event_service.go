package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
	"github.com/google/uuid"
)

// CalendarEventService mutates calendar events.
type CalendarEventService struct {
	repo domain.EventRepository
	tx   domain.Transactor
}

// NewCalendarEventService builds CalendarEventService.
func NewCalendarEventService(repo domain.EventRepository, tx domain.Transactor) *CalendarEventService {
	return &CalendarEventService{repo: repo, tx: tx}
}

// Create creates an event for organizer.
func (s *CalendarEventService) Create(ctx context.Context, actorID string, in CreateEventInput) (EventDTO, error) {
	org, err := domain.ParseUserID(actorID)
	if err != nil {
		return EventDTO{}, err
	}
	attendees, err := parseAttendees(in.AttendeeIDs)
	if err != nil {
		return EventDTO{}, err
	}
	now := time.Now().UTC()
	id := domain.EventID(uuid.NewString())
	e, err := domain.NewEvent(id, org, in.Title, in.Description, in.StartsAt, in.EndsAt, in.RelatedType, in.RelatedID, attendees, now)
	if err != nil {
		return EventDTO{}, err
	}
	if err := s.repo.Insert(ctx, e); err != nil {
		return EventDTO{}, err
	}
	return toDTO(e), nil
}

// Update updates event when actor can manage.
func (s *CalendarEventService) Update(ctx context.Context, actorID, eventID string, in UpdateEventInput, isAdmin bool) (EventDTO, error) {
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return EventDTO{}, err
	}
	eid, err := domain.ParseEventID(eventID)
	if err != nil {
		return EventDTO{}, err
	}
	attendees, err := parseAttendees(in.AttendeeIDs)
	if err != nil {
		return EventDTO{}, err
	}
	var out domain.CalendarEvent
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		e, err := s.repo.GetForUpdate(ctx, eid)
		if err != nil {
			return err
		}
		if !e.CanManage(actor, isAdmin) {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		if err := e.Update(in.Title, in.Description, in.StartsAt, in.EndsAt, attendees, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, e); err != nil {
			return err
		}
		out = e
		return nil
	})
	if err != nil {
		return EventDTO{}, err
	}
	return toDTO(out), nil
}

// Cancel cancels event.
func (s *CalendarEventService) Cancel(ctx context.Context, actorID, eventID string, isAdmin bool) (EventDTO, error) {
	return s.mutate(ctx, actorID, eventID, isAdmin, func(e *domain.CalendarEvent, now time.Time) error {
		return e.Cancel(now)
	})
}

// Delete soft-deletes event.
func (s *CalendarEventService) Delete(ctx context.Context, actorID, eventID string, isAdmin bool) error {
	_, err := s.mutate(ctx, actorID, eventID, isAdmin, func(e *domain.CalendarEvent, now time.Time) error {
		return e.Delete(now)
	})
	return err
}

func (s *CalendarEventService) mutate(ctx context.Context, actorID, eventID string, isAdmin bool, fn func(*domain.CalendarEvent, time.Time) error) (EventDTO, error) {
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return EventDTO{}, err
	}
	eid, err := domain.ParseEventID(eventID)
	if err != nil {
		return EventDTO{}, err
	}
	var out domain.CalendarEvent
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		e, err := s.repo.GetForUpdate(ctx, eid)
		if err != nil {
			return err
		}
		if !e.CanManage(actor, isAdmin) {
			return domain.ErrForbidden
		}
		now := time.Now().UTC()
		if err := fn(&e, now); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, e); err != nil {
			return err
		}
		out = e
		return nil
	})
	if err != nil {
		return EventDTO{}, err
	}
	return toDTO(out), nil
}
