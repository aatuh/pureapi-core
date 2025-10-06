package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Builtin router with path params and 405 behavior.
func Test_Router_ParamsAnd405(t *testing.T) {
	server := pureapi.NewServer(
		pureapi.WithRouter(pureapi.NewBuiltinRouter()),
	)

	server.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := pureapi.RouteParams(r)
		if params["id"] == "" {
			t.Fatalf("missing route param id")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Allowed method
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Disallowed method on existing path => 405
	req = httptest.NewRequest(http.MethodPost, "/users/999", nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
