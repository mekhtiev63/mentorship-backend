package domain

// Outbox event names consumed by activity (mirror producer BCs).
const (
	SourceProgressMaterialViewed = "progress.material.viewed"
	SourceProgressBlockApproved  = "progress.block.approved"
	SourceAchievementGranted     = "achievement.granted"
	SourceBonusCredited          = "bonus.credited"
	SourceBonusConverted         = "bonus.converted"
	SourceInterviewRealCreated   = "interview.real.created"
	SourceInterviewRealUpdated   = "interview.real.updated"
	SourceInterviewMockScheduled = "interview.mock.scheduled"
	SourceFinalCheckBothCompleted = "final_check.both_completed"
	SourceOneOnOneApproved       = "one_on_one.request.approved"
	SourceCalendarEventCreated   = "calendar.event.created"
)

var activitySourceEvents = []string{
	SourceProgressMaterialViewed,
	SourceProgressBlockApproved,
	SourceAchievementGranted,
	SourceBonusCredited,
	SourceBonusConverted,
	SourceInterviewRealCreated,
	SourceInterviewRealUpdated,
	SourceInterviewMockScheduled,
	SourceFinalCheckBothCompleted,
	SourceOneOnOneApproved,
	SourceCalendarEventCreated,
}

// ActivitySourceEventNames returns event names for SQL IN clause.
func ActivitySourceEventNames() []string {
	out := make([]string, len(activitySourceEvents))
	copy(out, activitySourceEvents)
	return out
}
