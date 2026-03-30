package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

// SetupMockServer creates an httptest server and wires it into ClientFactory.
// Returns the server (for custom assertions) and cleans up on test completion.
func SetupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(func() {
		srv.Close()
		ClientFactory = nil
	})
	ClientFactory = func() (*api.Client, error) {
		return api.NewTestClient(srv.URL, "test-key", "test-client-key"), nil
	}
	return srv
}
