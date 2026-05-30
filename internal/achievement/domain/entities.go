package domain

import "time"

// AchievementDefinition is catalog entry.
type AchievementDefinition struct {
	Code        AchievementCode
	Title       string
	Description string
	Rule        AchievementRule
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserAchievement is granted achievement (aggregate root).
type UserAchievement struct {
	UserID         UserID
	AchievementCode AchievementCode
	GrantedAt      time.Time
	SourceEventID  SourceEventID
}
