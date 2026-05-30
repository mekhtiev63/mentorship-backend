package stub

import (
	"net/http"

	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/response"
)

// Handler responds with HTTP 501 until the module route is implemented.
func Handler(module string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response.Error(w, http.StatusNotImplemented, "not_implemented", module+" API is not implemented yet")
	}
}
