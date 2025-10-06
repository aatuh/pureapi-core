package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Using default PlainDecoder and reading QueryMap from context.
func Test_QueryDecoder_Plain(t *testing.T) {
	server := pureapi.NewServer()

	server.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		qm := pureapi.QueryMap(r)
		if qm == nil {
			t.Fatalf("expected query map in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Plain decoder turns single values into strings and multi-values into
	// []string
	req := httptest.NewRequest(
		http.MethodGet, "/search?q=go&tags=go&tags=api", nil,
	)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
