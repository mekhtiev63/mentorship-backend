package progress

import (
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	progresshttp "github.com/go-mentorship-platform/backend/internal/progress/adapter/http"
	progresspersistence "github.com/go-mentorship-platform/backend/internal/progress/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/progress/application"
)

// Module wires the progress bounded context.
type Module struct {
	HTTP *progresshttp.Module
}

// NewModule constructs progress services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	tx := progresspersistence.NewTransactor(pool)
	progress := progresspersistence.NewBlockProgressRepo(pool)
	views := progresspersistence.NewMaterialViewRepo(pool)
	roadmap := progresspersistence.NewRoadmapPolicyReader(pool)
	buddy := progresspersistence.NewBuddyScopeReader(pool)
	outbox := progresspersistence.NewOutboxRecorder(pool)

	materialSvc := application.NewMaterialProgressService(tx, progress, views, roadmap, outbox)
	approvalSvc := application.NewBlockApprovalService(tx, progress, buddy, outbox)
	studentQuery := application.NewStudentProgressQueryService(progress, views, roadmap)
	buddyQuery := application.NewBuddyStudentsQueryService(progress, buddy, studentQuery)

	httpModule := &progresshttp.Module{
		Student: progresshttp.NewStudentHandler(materialSvc, studentQuery),
		Buddy:   progresshttp.NewBuddyHandler(approvalSvc, buddyQuery),
		Admin:   progresshttp.NewAdminHandler(approvalSvc, studentQuery),
		Authenticate:      identityHTTP.Authenticate,
		RequireRole:       identityHTTP.RequireRole,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}

	return &Module{HTTP: httpModule}
}
