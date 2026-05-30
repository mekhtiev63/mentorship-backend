package routes

import (
	"github.com/go-chi/chi/v5"
	adminhttp "github.com/go-mentorship-platform/backend/internal/admin/adapter/http"
	achievementhttp "github.com/go-mentorship-platform/backend/internal/achievement/adapter/http"
	bonushttp "github.com/go-mentorship-platform/backend/internal/bonus/adapter/http"
	activityhttp "github.com/go-mentorship-platform/backend/internal/activity/adapter/http"
	notificationhttp "github.com/go-mentorship-platform/backend/internal/notification/adapter/http"
	calendarhttp "github.com/go-mentorship-platform/backend/internal/calendar/adapter/http"
	finalcheckhttp "github.com/go-mentorship-platform/backend/internal/finalcheck/adapter/http"
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	interviewhttp "github.com/go-mentorship-platform/backend/internal/interview/adapter/http"
	oneononehttp "github.com/go-mentorship-platform/backend/internal/oneonone/adapter/http"
	profilehttp "github.com/go-mentorship-platform/backend/internal/profile/adapter/http"
	progresshttp "github.com/go-mentorship-platform/backend/internal/progress/adapter/http"
	roadmaphttp "github.com/go-mentorship-platform/backend/internal/roadmap/adapter/http"
	userhttp "github.com/go-mentorship-platform/backend/internal/user/adapter/http"
)

// V1Deps bundles modules for API v1 route registration.
type V1Deps struct {
	Identity *identityhttp.Module
	User     *userhttp.Module
	Profile  *profilehttp.Module
	Roadmap  *roadmaphttp.Module
	Progress    *progresshttp.Module
	Achievement *achievementhttp.Module
	Bonus       *bonushttp.Module
	OneOnOne    *oneononehttp.Module
	Interview   *interviewhttp.Module
	FinalCheck  *finalcheckhttp.Module
	Calendar    *calendarhttp.Module
	Activity    *activityhttp.Module
	Notification *notificationhttp.Module
}

// RegisterV1 mounts versioned API route groups.
func RegisterV1(r chi.Router, deps V1Deps) {
	if deps.Identity != nil {
		deps.Identity.Register(r)
	}
	if deps.User != nil {
		deps.User.Register(r)
	}
	if deps.Profile != nil {
		deps.Profile.Register(r)
	}
	adminhttp.Register(r, adminhttp.RegisterDeps{
		User:       deps.User,
		Roadmap:    deps.Roadmap,
		Progress:   deps.Progress,
		OneOnOne:   deps.OneOnOne,
		FinalCheck: deps.FinalCheck,
		Activity:   deps.Activity,
	})
	if deps.Roadmap != nil {
		deps.Roadmap.Register(r)
	}
	if deps.Progress != nil {
		deps.Progress.Register(r)
	}
	if deps.OneOnOne != nil {
		deps.OneOnOne.Register(r)
	}
	if deps.Interview != nil {
		deps.Interview.Register(r)
	}
	if deps.FinalCheck != nil {
		deps.FinalCheck.Register(r)
	}
	if deps.Calendar != nil {
		deps.Calendar.Register(r)
	}
	if deps.Activity != nil {
		deps.Activity.Register(r)
	}
	if deps.Notification != nil {
		deps.Notification.Register(r)
	}
	if deps.Achievement != nil {
		deps.Achievement.Register(r)
	}
	if deps.Bonus != nil {
		deps.Bonus.Register(r)
	}
}
