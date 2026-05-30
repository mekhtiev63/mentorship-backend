package domain

// ActivityType is normalized feed filter type.
type ActivityType string

const (
	TypeMaterialViewed       ActivityType = "material_viewed"
	TypeBlockApproved        ActivityType = "block_approved"
	TypeAchievementGranted   ActivityType = "achievement_granted"
	TypeBonusTransaction     ActivityType = "bonus_transaction"
	TypeInterviewCreated     ActivityType = "interview_created"
	TypeInterviewUpdated     ActivityType = "interview_updated"
	TypeFinalCheckCompleted  ActivityType = "final_check_completed"
	TypeOneOnOneApproved     ActivityType = "one_on_one_approved"
	TypeCalendarEventCreated ActivityType = "calendar_event_created"
)
