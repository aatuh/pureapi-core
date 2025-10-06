package examples

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Demonstrate WithBodyLimit rejecting large request bodies.
func Test_BodyLimit(t *testing.T) {
	server := pureapi.NewServer(
		pureapi.WithBodyLimit(4),
	)

	server.Post("/echo", func(w http.ResponseWriter, r *http.Request) {
		// If body is allowed, just read it.
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	// Send body larger than limit
	req := httptest.NewRequest(
		http.MethodPost, "/echo", strings.NewReader("12345"),
	)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}
