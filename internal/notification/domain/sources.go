package domain

const (
	SourceProgressBlockApproved   = "progress.block.approved"
	SourceAchievementGranted      = "achievement.granted"
	SourceBonusCredited           = "bonus.credited"
	SourceBonusConverted          = "bonus.converted"
	SourceInterviewMockScheduled  = "interview.mock.scheduled"
	SourceFinalCheckEligibility   = "final_check.eligibility.granted"
	SourceOneOnOneApproved        = "one_on_one.request.approved"
	SourceOneOnOneRejected        = "one_on_one.request.rejected"
	SourceOneOnOneCompleted       = "one_on_one.request.completed"
)

var notificationSourceEvents = []string{
	SourceProgressBlockApproved,
	SourceAchievementGranted,
	SourceBonusCredited,
	SourceBonusConverted,
	SourceInterviewMockScheduled,
	SourceFinalCheckEligibility,
	SourceOneOnOneApproved,
	SourceOneOnOneRejected,
	SourceOneOnOneCompleted,
}

// NotificationSourceEventNames returns whitelisted outbox event names.
func NotificationSourceEventNames() []string {
	out := make([]string, len(notificationSourceEvents))
	copy(out, notificationSourceEvents)
	return out
}
