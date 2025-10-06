package examples

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Demonstrate automatic HEAD fallback and synthesized OPTIONS responses.
func Test_AutoHEADAndOPTIONS(t *testing.T) {
	server := pureapi.NewServer()

	server.Get("/resources/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("resource"))
	})

	server.Post("/resources/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// HEAD should reuse GET logic but discard the body.
	req := httptest.NewRequest(http.MethodHead, "/resources/42", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for HEAD, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body for HEAD, got %q", rr.Body.String())
	}

	// OPTIONS should advertise the union of registered methods.
	req = httptest.NewRequest(http.MethodOptions, "/resources/42", nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", rr.Code)
	}

	allow := rr.Header().Get("Allow")
	if allow == "" {
		t.Fatalf("expected Allow header to be set")
	}

	expected := []string{"OPTIONS", "GET", "HEAD", "POST"}
	got := strings.Split(allow, ",")
	if len(got) != len(expected) {
		t.Fatalf("expected %d methods, got %d (%q)", len(expected), len(got), allow)
	}

	for i, method := range got {
		method = strings.TrimSpace(method)
		if method != expected[i] {
			t.Fatalf("Allow[%d] = %q, want %q (full header %q)", i, method, expected[i], allow)
		}
	}
}
