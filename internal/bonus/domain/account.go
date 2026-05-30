package domain

import "time"

// UserID identifies account owner.
type UserID string

// BonusAccount holds materialized balance.
type BonusAccount struct {
	UserID    UserID
	Balance   BonusAmount
	UpdatedAt time.Time
}
