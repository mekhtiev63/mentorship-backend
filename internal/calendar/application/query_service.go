package application

import (
	"context"
	"time"

	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// CalendarQueryService reads calendar events.
type CalendarQueryService struct {
	repo domain.EventRepository
}

// NewCalendarQueryService builds CalendarQueryService.
func NewCalendarQueryService(repo domain.EventRepository) *CalendarQueryService {
	return &CalendarQueryService{repo: repo}
}

// Get returns event if actor can read.
func (s *CalendarQueryService) Get(ctx context.Context, actorID, eventID string, isAdmin bool) (EventDTO, error) {
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return EventDTO{}, err
	}
	eid, err := domain.ParseEventID(eventID)
	if err != nil {
		return EventDTO{}, err
	}
	e, err := s.repo.GetByID(ctx, eid)
	if err != nil {
		return EventDTO{}, err
	}
	if !e.CanRead(actor, isAdmin) {
		return EventDTO{}, domain.ErrForbidden
	}
	return toDTO(e), nil
}

// List lists events for actor with filters.
func (s *CalendarQueryService) List(ctx context.Context, actorID string, q ListEventsQuery, isAdmin bool) (ListResult, error) {
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(q.Page, q.PageSize)
	filter := domain.EventFilter{
		From:             q.From,
		To:               q.To,
		IncludeCancelled: q.IncludeCancelled,
		IncludeDeleted:   q.IncludeDeleted && isAdmin,
	}
	if q.RelatedType != nil && *q.RelatedType != "" {
		rt, _ := domain.ParseRelatedType(*q.RelatedType)
		filter.RelatedType = &rt
	}
	items, total, err := s.repo.ListForUser(ctx, actor, filter, p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}

// ListUpcoming lists upcoming active events.
func (s *CalendarQueryService) ListUpcoming(ctx context.Context, actorID string, page, pageSize int) (ListResult, error) {
	actor, err := domain.ParseUserID(actorID)
	if err != nil {
		return ListResult{}, err
	}
	p := pagination.Normalize(page, pageSize)
	from := time.Now().UTC()
	items, total, err := s.repo.ListUpcoming(ctx, actor, from, p)
	if err != nil {
		return ListResult{}, err
	}
	return listDTO(items, p, total), nil
}
