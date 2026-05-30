package domain

import "time"

// ActivityFilter lists query filters.
type ActivityFilter struct {
	ActivityType *ActivityType
	Verb         *string
	ObjectType   *string
	From         *time.Time
	To           *time.Time
	SubjectUserID *UserID
	ActorID      *UserID
}
