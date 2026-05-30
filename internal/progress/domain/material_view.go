package domain

import "time"

// MaterialView records first view of a material by a student.
type MaterialView struct {
	ID             string
	StudentID      StudentID
	MaterialID     MaterialID
	FirstViewedAt  time.Time
	IdempotencyKey *string
}
