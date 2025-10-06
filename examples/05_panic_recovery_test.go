package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Demonstrate default panic recovery returning 500.
func Test_PanicRecovery_Default(t *testing.T) {
	server := pureapi.NewServer()

	server.Get("/boom", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
