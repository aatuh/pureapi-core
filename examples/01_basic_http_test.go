package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Basic server with a single endpoint.
func Test_BasicHTTP(t *testing.T) {
	server := pureapi.NewServer()

	server.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}
