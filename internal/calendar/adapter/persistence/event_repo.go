package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/calendar/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EventRepo implements EventRepository.
type EventRepo struct {
	pool *postgres.Pool
}

// NewEventRepo creates EventRepo.
func NewEventRepo(pool *postgres.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

const selectEvent = `
	SELECT id, organizer_id, title, description, starts_at, ends_at, related_type, related_id,
	       cancelled_at, deleted_at, created_at, updated_at
	FROM calendar_events
`

// Insert creates event and attendees.
func (r *EventRepo) Insert(ctx context.Context, e domain.CalendarEvent) error {
	id := string(e.ID)
	if id == "" {
		id = uuid.NewString()
	}
	const q = `
		INSERT INTO calendar_events (id, organizer_id, title, description, starts_at, ends_at, related_type, related_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, string(e.OrganizerID), e.Title, e.Description, e.StartsAt, e.EndsAt, string(e.RelatedType), e.RelatedID,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return r.replaceAttendees(ctx, domain.EventID(id), e.AttendeeIDs)
}

// GetByID loads event with attendees.
func (r *EventRepo) GetByID(ctx context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	e, err := scanEvent(db(ctx, r.pool).QueryRow(ctx, selectEvent+` WHERE id = $1`, string(id)))
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	att, err := r.loadAttendees(ctx, id)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	e.AttendeeIDs = att
	return e, nil
}

// GetForUpdate loads with lock.
func (r *EventRepo) GetForUpdate(ctx context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	e, err := scanEvent(db(ctx, r.pool).QueryRow(ctx, selectEvent+` WHERE id = $1 FOR UPDATE`, string(id)))
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	att, err := r.loadAttendees(ctx, id)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	e.AttendeeIDs = att
	return e, nil
}

// Save updates event row and attendees.
func (r *EventRepo) Save(ctx context.Context, e domain.CalendarEvent) error {
	const q = `
		UPDATE calendar_events SET
			title = $2, description = $3, starts_at = $4, ends_at = $5,
			related_type = $6, related_id = $7, cancelled_at = $8, deleted_at = $9, updated_at = now()
		WHERE id = $1
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		string(e.ID), e.Title, e.Description, e.StartsAt, e.EndsAt, string(e.RelatedType), e.RelatedID,
		e.CancelledAt, e.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("save event: %w", err)
	}
	return r.replaceAttendees(ctx, e.ID, e.AttendeeIDs)
}

// ListForUser lists events visible to user.
func (r *EventRepo) ListForUser(ctx context.Context, userID domain.UserID, filter domain.EventFilter, page pagination.Params) ([]domain.CalendarEvent, int64, error) {
	where, args := r.visibilityWhere(userID, filter)
	return r.list(ctx, where, args, page)
}

// ListUpcoming lists future active events for user.
func (r *EventRepo) ListUpcoming(ctx context.Context, userID domain.UserID, from time.Time, page pagination.Params) ([]domain.CalendarEvent, int64, error) {
	filter := domain.EventFilter{From: &from}
	where, args := r.visibilityWhere(userID, filter)
	where += ` AND e.cancelled_at IS NULL AND e.deleted_at IS NULL AND e.starts_at >= $` + fmt.Sprint(len(args)+1)
	args = append(args, from)
	return r.list(ctx, where, args, page)
}

func (r *EventRepo) visibilityWhere(userID domain.UserID, filter domain.EventFilter) (string, []any) {
	args := []any{string(userID)}
	where := `
		FROM calendar_events e
		WHERE (
			e.organizer_id = $1
			OR EXISTS (SELECT 1 FROM calendar_event_attendees a WHERE a.event_id = e.id AND a.user_id = $1)
		)
	`
	n := 2
	if !filter.IncludeDeleted {
		where += ` AND e.deleted_at IS NULL`
	}
	if !filter.IncludeCancelled {
		where += ` AND e.cancelled_at IS NULL`
	}
	if filter.From != nil {
		where += fmt.Sprintf(` AND e.ends_at >= $%d`, n)
		args = append(args, *filter.From)
		n++
	}
	if filter.To != nil {
		where += fmt.Sprintf(` AND e.starts_at <= $%d`, n)
		args = append(args, *filter.To)
		n++
	}
	if filter.RelatedType != nil {
		where += fmt.Sprintf(` AND e.related_type = $%d`, n)
		args = append(args, string(*filter.RelatedType))
		n++
	}
	_ = n
	return where, args
}

func (r *EventRepo) list(ctx context.Context, where string, args []any, page pagination.Params) ([]domain.CalendarEvent, int64, error) {
	countQ := `SELECT COUNT(*) ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, page.PageSize, page.Offset)
	q := `SELECT e.id, e.organizer_id, e.title, e.description, e.starts_at, e.ends_at, e.related_type, e.related_id,
	       e.cancelled_at, e.deleted_at, e.created_at, e.updated_at ` + where +
		fmt.Sprintf(` ORDER BY e.starts_at ASC LIMIT $%d OFFSET $%d`, limitArg, offsetArg)
	rows, err := db(ctx, r.pool).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.CalendarEvent
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, 0, err
		}
		att, err := r.loadAttendees(ctx, e.ID)
		if err != nil {
			return nil, 0, err
		}
		e.AttendeeIDs = att
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *EventRepo) loadAttendees(ctx context.Context, id domain.EventID) ([]domain.UserID, error) {
	const q = `SELECT user_id FROM calendar_event_attendees WHERE event_id = $1 ORDER BY user_id`
	rows, err := db(ctx, r.pool).Query(ctx, q, string(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.UserID
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		out = append(out, domain.UserID(uid))
	}
	return out, rows.Err()
}

func (r *EventRepo) replaceAttendees(ctx context.Context, id domain.EventID, attendees []domain.UserID) error {
	_, err := db(ctx, r.pool).Exec(ctx, `DELETE FROM calendar_event_attendees WHERE event_id = $1`, string(id))
	if err != nil {
		return err
	}
	for _, a := range attendees {
		if string(a) == "" {
			continue
		}
		_, err := db(ctx, r.pool).Exec(ctx,
			`INSERT INTO calendar_event_attendees (event_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			string(id), string(a),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanEvent(row pgx.Row) (domain.CalendarEvent, error) {
	var e domain.CalendarEvent
	var id, org, title, desc, relType string
	err := row.Scan(
		&id, &org, &title, &desc, &e.StartsAt, &e.EndsAt, &relType, &e.RelatedID,
		&e.CancelledAt, &e.DeletedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CalendarEvent{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	e.ID = domain.EventID(id)
	e.OrganizerID = domain.UserID(org)
	e.Title = title
	e.Description = desc
	e.RelatedType = domain.RelatedType(relType)
	return e, nil
}
