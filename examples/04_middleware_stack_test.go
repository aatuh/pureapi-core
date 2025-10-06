package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
)

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// pretend to log
		next.ServeHTTP(w, r)
	})
}

func authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Building middleware stacks and per-endpoint customization.
func Test_MiddlewareStack(t *testing.T) {
	server := pureapi.NewServer()

	shared := pureapi.NewStack().AddWrapper(
		pureapi.NewWrapperFromHandler("logging", logMW),
	)

	secure := shared.Clone()
	secure, _ = secure.InsertAfter(
		"logging", pureapi.NewWrapperFromHandler("auth", authMW),
	)

	server.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).WithMiddlewares(shared.Middlewares())

	server.Get("/secure", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).WithMiddlewares(secure.Middlewares())

	// Public ok
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("public expected 200, got %d", rr.Code)
	}

	// Secure unauthorized without header
	req = httptest.NewRequest(http.MethodGet, "/secure", nil)
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("secure expected 401, got %d", rr.Code)
	}

	// Secure ok with header
	req = httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("X-Auth", "token")
	rr = httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("secure expected 200, got %d", rr.Code)
	}
}
