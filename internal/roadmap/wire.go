package roadmap

import (
	identityhttp "github.com/go-mentorship-platform/backend/internal/identity/adapter/http"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	roadmaphttp "github.com/go-mentorship-platform/backend/internal/roadmap/adapter/http"
	roadmappersistence "github.com/go-mentorship-platform/backend/internal/roadmap/adapter/persistence"
	"github.com/go-mentorship-platform/backend/internal/roadmap/application"
)

// Module wires the roadmap bounded context.
type Module struct {
	HTTP *roadmaphttp.Module
}

// NewModule constructs roadmap services and HTTP module.
func NewModule(pool *postgres.Pool, identityHTTP *identityhttp.Module) *Module {
	blocks := roadmappersistence.NewBlockRepo(pool)
	materials := roadmappersistence.NewMaterialRepo(pool)
	progress := roadmappersistence.NewProgressReader(pool)
	outbox := roadmappersistence.NewOutboxRecorder(pool)

	admin := application.NewAdminService(blocks, materials, progress, outbox)
	catalog := application.NewCatalogService(blocks, materials)
	student := application.NewStudentRoadmapService(blocks, materials)

	httpModule := &roadmaphttp.Module{
		Admin:             roadmaphttp.NewAdminHandler(admin),
		Catalog:           roadmaphttp.NewCatalogHandler(catalog),
		Student:           roadmaphttp.NewStudentHandler(student),
		Authenticate:      identityHTTP.Authenticate,
		RequireRole:       identityHTTP.RequireRole,
		RequireActiveRole: identityHTTP.RequireActiveRole,
	}

	return &Module{HTTP: httpModule}
}
