package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	"github.com/go-mentorship-platform/backend/internal/activity/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// JournalRepo implements ActivityJournalRepository.
type JournalRepo struct {
	pool *postgres.Pool
}

// NewJournalRepo creates JournalRepo.
func NewJournalRepo(pool *postgres.Pool) *JournalRepo {
	return &JournalRepo{pool: pool}
}

// Append inserts journal entry (idempotent by source_outbox_id).
func (r *JournalRepo) Append(ctx context.Context, e domain.ActivityEntry) error {
	id := string(e.ID)
	if id == "" {
		id = uuid.NewString()
	}
	var actor *string
	if e.ActorID != nil {
		s := string(*e.ActorID)
		actor = &s
	}
	var objID *string
	if e.ObjectID != nil && uuidValid(*e.ObjectID) {
		objID = e.ObjectID
	}
	const q = `
		INSERT INTO activity_events (
			id, actor_id, verb, object_type, object_id, payload,
			subject_user_id, activity_type, source_outbox_id, source_event_name, occurred_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12)
	`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		id, actor, e.Verb, e.ObjectType, objID, e.Payload,
		string(e.SubjectUserID), string(e.ActivityType), e.SourceOutboxID, e.SourceEventName,
		e.OccurredAt, e.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrDuplicate
		}
		return err
	}
	return nil
}

// GetByID loads entry.
func (r *JournalRepo) GetByID(ctx context.Context, id domain.ActivityID) (domain.ActivityEntry, error) {
	const q = selectEntry + ` WHERE id = $1`
	return scanEntry(r.pool.QueryRow(ctx, q, string(id)))
}

// ListBySubject lists subject timeline.
func (r *JournalRepo) ListBySubject(ctx context.Context, subject domain.UserID, filter domain.ActivityFilter, page pagination.Params) ([]domain.ActivityEntry, int64, error) {
	filter.SubjectUserID = &subject
	return r.list(ctx, filter, page)
}

// ListAll lists all entries (admin).
func (r *JournalRepo) ListAll(ctx context.Context, filter domain.ActivityFilter, page pagination.Params) ([]domain.ActivityEntry, int64, error) {
	return r.list(ctx, filter, page)
}

func (r *JournalRepo) list(ctx context.Context, filter domain.ActivityFilter, page pagination.Params) ([]domain.ActivityEntry, int64, error) {
	where := `WHERE 1=1`
	args := []any{}
	n := 1
	if filter.SubjectUserID != nil {
		where += fmt.Sprintf(` AND subject_user_id = $%d`, n)
		args = append(args, string(*filter.SubjectUserID))
		n++
	}
	if filter.ActorID != nil {
		where += fmt.Sprintf(` AND actor_id = $%d`, n)
		args = append(args, string(*filter.ActorID))
		n++
	}
	if filter.ActivityType != nil {
		where += fmt.Sprintf(` AND activity_type = $%d`, n)
		args = append(args, string(*filter.ActivityType))
		n++
	}
	if filter.Verb != nil && *filter.Verb != "" {
		where += fmt.Sprintf(` AND verb = $%d`, n)
		args = append(args, *filter.Verb)
		n++
	}
	if filter.ObjectType != nil && *filter.ObjectType != "" {
		where += fmt.Sprintf(` AND object_type = $%d`, n)
		args = append(args, *filter.ObjectType)
		n++
	}
	if filter.From != nil {
		where += fmt.Sprintf(` AND occurred_at >= $%d`, n)
		args = append(args, *filter.From)
		n++
	}
	if filter.To != nil {
		where += fmt.Sprintf(` AND occurred_at <= $%d`, n)
		args = append(args, *filter.To)
		n++
	}
	countQ := `SELECT COUNT(*) FROM activity_events ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitPh := fmt.Sprintf(`$%d`, n)
	offsetPh := fmt.Sprintf(`$%d`, n+1)
	args = append(args, page.PageSize, page.Offset)
	q := selectEntry + ` ` + where + ` ORDER BY occurred_at DESC LIMIT ` + limitPh + ` OFFSET ` + offsetPh
	rows, err := db(ctx, r.pool).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.ActivityEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

const selectEntry = `
	SELECT id, subject_user_id, actor_id, activity_type, verb, object_type, object_id, payload,
	       source_outbox_id, source_event_name, occurred_at, created_at
	FROM activity_events
`

func scanEntry(row pgx.Row) (domain.ActivityEntry, error) {
	var e domain.ActivityEntry
	var id, subject, actType, verb, objType string
	var actor, objID, sourceOutbox *string
	err := row.Scan(
		&id, &subject, &actor, &actType, &verb, &objType, &objID, &e.Payload,
		&sourceOutbox, &e.SourceEventName, &e.OccurredAt, &e.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ActivityEntry{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ActivityEntry{}, err
	}
	e.ID = domain.ActivityID(id)
	e.SubjectUserID = domain.UserID(subject)
	e.ActivityType = domain.ActivityType(actType)
	e.Verb = verb
	e.ObjectType = objType
	e.ObjectID = objID
	if actor != nil {
		u := domain.UserID(*actor)
		e.ActorID = &u
	}
	if sourceOutbox != nil {
		e.SourceOutboxID = *sourceOutbox
	}
	return e, nil
}

func uuidValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
