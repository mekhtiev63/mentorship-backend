package application

import (
	"encoding/json"
	"time"

	"github.com/go-mentorship-platform/backend/internal/activity/domain"
	"github.com/go-mentorship-platform/backend/pkg/pagination"
)

// EntryDTO is API journal row.
type EntryDTO struct {
	ID              string          `json:"id"`
	SubjectUserID   string          `json:"subject_user_id"`
	ActorID         *string         `json:"actor_id,omitempty"`
	ActivityType    string          `json:"activity_type"`
	Verb            string          `json:"verb"`
	ObjectType      string          `json:"object_type"`
	ObjectID        *string         `json:"object_id,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	OccurredAt      time.Time       `json:"occurred_at"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ListResult paginated feed.
type ListResult struct {
	Items []EntryDTO      `json:"items"`
	Meta  pagination.Meta `json:"meta"`
}

// ListQuery HTTP list filters.
type ListQuery struct {
	ActivityType *string
	Verb         *string
	ObjectType   *string
	From         *time.Time
	To           *time.Time
	SubjectUserID *string
	ActorID      *string
	Page         int
	PageSize     int
}

func toDTO(e domain.ActivityEntry) EntryDTO {
	d := EntryDTO{
		ID:            string(e.ID),
		SubjectUserID: string(e.SubjectUserID),
		ActivityType:  string(e.ActivityType),
		Verb:          e.Verb,
		ObjectType:    e.ObjectType,
		ObjectID:      e.ObjectID,
		Payload:       e.Payload,
		OccurredAt:    e.OccurredAt,
		CreatedAt:     e.CreatedAt,
	}
	if e.ActorID != nil {
		s := string(*e.ActorID)
		d.ActorID = &s
	}
	if len(d.Payload) == 0 {
		d.Payload = json.RawMessage(`{}`)
	}
	return d
}

func listDTO(items []domain.ActivityEntry, p pagination.Params, total int64) ListResult {
	dtos := make([]EntryDTO, 0, len(items))
	for _, e := range items {
		dtos = append(dtos, toDTO(e))
	}
	return ListResult{Items: dtos, Meta: pagination.NewMeta(p.Page, p.PageSize, total)}
}

func toFilter(q ListQuery) domain.ActivityFilter {
	f := domain.ActivityFilter{
		Verb:       q.Verb,
		ObjectType: q.ObjectType,
		From:       q.From,
		To:         q.To,
	}
	if q.ActivityType != nil && *q.ActivityType != "" {
		t := domain.ActivityType(*q.ActivityType)
		f.ActivityType = &t
	}
	if q.SubjectUserID != nil && *q.SubjectUserID != "" {
		u := domain.UserID(*q.SubjectUserID)
		f.SubjectUserID = &u
	}
	if q.ActorID != nil && *q.ActorID != "" {
		u := domain.UserID(*q.ActorID)
		f.ActorID = &u
	}
	return f
}
