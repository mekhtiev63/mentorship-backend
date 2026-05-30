package domain

// NotificationType is a normalized in-app notification kind.
type NotificationType string

const (
	TypeAchievementGranted    NotificationType = "achievement_granted"
	TypeBlockApproved           NotificationType = "block_approved"
	TypeMockInterviewAssigned   NotificationType = "mock_interview_assigned"
	TypeFinalCheckAssigned      NotificationType = "final_check_assigned"
	TypeOneOnOneApproved        NotificationType = "one_on_one_approved"
	TypeOneOnOneRejected        NotificationType = "one_on_one_rejected"
	TypeOneOnOneCompleted       NotificationType = "one_on_one_completed"
	TypeBonusCredited           NotificationType = "bonus_credited"
	TypeBonusDebited            NotificationType = "bonus_debited"
)
