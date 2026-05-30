package authorization

// Action identifies an operation checked by Authorize.
type Action string

const (
	ActionProfileRead       Action = "profile:read"
	ActionProfileWrite      Action = "profile:write"
	ActionRoadmapRead       Action = "roadmap:read"
	ActionRoadmapWrite      Action = "roadmap:write"
	ActionProgressRead      Action = "progress:read"
	ActionProgressWrite     Action = "progress:write"
	ActionProgressApprove   Action = "progress:approve"
	ActionOneOnOneManage    Action = "one_on_one:manage"
	ActionCalendarManage    Action = "calendar:manage"
	ActionInterviewManage   Action = "interview:manage"
	ActionFinalCheckManage  Action = "final_check:manage"
	ActionAchievementRead   Action = "achievement:read"
	ActionActivityRead      Action = "activity:read"
	ActionBonusRead         Action = "bonus:read"
	ActionBonusConvert      Action = "bonus:convert"
	ActionAdminManage       Action = "admin:manage"
)

// Resource is the target of an authorization check.
type Resource struct {
	Type       string
	ID         string
	OwnerID    string
	StudentID  string
}

// Authorize reports whether the actor may perform the action on the resource.
// Business rules are implemented in later iterations.
func (s *Service) Authorize(_ Action, _ Resource, _ string, _ []Role) bool {
	return false
}
