package domain

const (
	EventEligibilityGranted = "final_check.eligibility.granted"
	EventTechScheduled      = "final_check.tech.scheduled"
	EventTechCompleted      = "final_check.tech.completed"
	EventTechFailed         = "final_check.tech.failed"
	EventRoastAvailable     = "final_check.roast.available"
	EventRoastScheduled     = "final_check.roast.scheduled"
	EventRoastCompleted     = "final_check.roast.completed"
	EventRoastFailed        = "final_check.roast.failed"
	EventBothCompleted      = "final_check.both_completed"
)

// EventBothCompleted is consumed by Achievement module for code "finalist".
const AchievementFinalistCode = "finalist"
