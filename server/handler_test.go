package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/querydec"
	"github.com/aatuh/pureapi-core/router"
)

func TestHandler_ServeHTTP_WithRouter(t *testing.T) {
	// Create a custom router for testing
	testRouter := router.NewBuiltinRouter()

	// Create handler with custom router
	handler := NewHandler(
		event.NewNoopEventEmitter(),
		WithRouter(testRouter),
	)

	// Register a test endpoint
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	testRouter.Register("GET", "/test", testHandler)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test response" {
		t.Fatalf("Expected 'test response', got '%s'", w.Body.String())
	}
}

func TestHandler_ServeHTTP_WithQueryDecoder(t *testing.T) {
	// Create handler with explicit query decoder
	handler := NewHandler(
		event.NewNoopEventEmitter(),
		WithQueryDecoder(querydec.PlainDecoder{}),
	)

	// Create a test router and register endpoint
	testRouter := router.NewBuiltinRouter()
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that query params are available in context
		queryMap := QueryMap(r)
		if queryMap == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("no query map"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("query decoded"))
	})

	testRouter.Register("GET", "/test", testHandler)
	handler.router = testRouter

	// Create request with query params
	req := httptest.NewRequest("GET", "/test?x=1&y=a", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "query decoded" {
		t.Fatalf("Expected 'query decoded', got '%s'", w.Body.String())
	}
}

func TestHandler_ServeHTTP_WithRouteParams(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create a test router and register endpoint with params
	testRouter := router.NewBuiltinRouter()
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that route params are available in context
		params := RouteParams(r)
		if params == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("no route params"))
			return
		}

		if params["id"] != "123" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("wrong param value: " + params["id"]))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("params ok"))
	})

	// Use colon parameter pattern
	testRouter.Register("GET", "/user/:id", testHandler)
	handler.router = testRouter

	// Create request
	req := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "params ok" {
		t.Fatalf("Expected 'params ok', got '%s'", w.Body.String())
	}
}

func TestHandler_ServeHTTP_NotFound(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create request for non-existent endpoint
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}
}

func TestHandler_Register(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create test endpoint
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("registered"))
	})

	ep := endpoint.NewEndpoint("/test", "GET").
		WithHandler(testHandler)

	// Register endpoint
	handler.Register([]endpoint.Endpoint{ep})

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "registered" {
		t.Fatalf("Expected 'registered', got '%s'", w.Body.String())
	}
}

func TestQueryMap(t *testing.T) {
	// Test with no query map in context
	req := httptest.NewRequest("GET", "/test", nil)

	queryMap := QueryMap(req)
	if queryMap != nil {
		t.Fatalf("Expected nil query map, got %v", queryMap)
	}
}

func TestRouteParams(t *testing.T) {
	// Test with no route params in context
	req := httptest.NewRequest("GET", "/test", nil)

	params := RouteParams(req)
	if params != nil {
		t.Fatalf("Expected nil route params, got %v", params)
	}
}

func TestHandler_HEAD_Fallback(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create test endpoint for GET
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	ep := endpoint.NewEndpoint("/test", "GET").
		WithHandler(testHandler)

	// Register endpoint
	handler.Register([]endpoint.Endpoint{ep})

	// Test HEAD request - should return 200 with empty body
	req := httptest.NewRequest("HEAD", "/test", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "" {
		t.Fatalf("Expected empty body for HEAD request, got '%s'", w.Body.String())
	}
}

func TestHandler_OPTIONS_Synthesized(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create test endpoints for GET and POST
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get response"))
	})

	postHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("post response"))
	})

	// Register endpoints
	handler.Register([]endpoint.Endpoint{
		endpoint.NewEndpoint("/test", "GET").WithHandler(getHandler),
		endpoint.NewEndpoint("/test", "POST").WithHandler(postHandler),
	})

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response - OPTIONS typically returns 204 No Content
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", w.Code)
	}

	// Check Allow header includes GET, POST, HEAD, and OPTIONS
	allow := w.Header().Get("Allow")
	if allow == "" {
		t.Fatalf("Expected Allow header to be set")
	}

	// Parse Allow header and check for expected methods
	expectedMethods := []string{"GET", "POST", "HEAD", "OPTIONS"}
	for _, method := range expectedMethods {
		if !containsMethod(allow, method) {
			t.Fatalf("Expected Allow header to contain %s, got %s", method, allow)
		}
	}
}

func TestHandler_OPTIONS_WithBraces(t *testing.T) {
	// Create handler
	handler := NewHandler(event.NewNoopEventEmitter())

	// Create test endpoint with braces pattern
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("braces response"))
	})

	// Register endpoint with braces pattern
	handler.Register([]endpoint.Endpoint{
		endpoint.NewEndpoint("/user/{id}", "GET").WithHandler(testHandler),
	})

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/user/123", nil)
	w := httptest.NewRecorder()

	// Serve request
	handler.ServeHTTP(w, req)

	// Check response - OPTIONS typically returns 204 No Content
	if w.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", w.Code)
	}

	// Check Allow header includes GET, HEAD, and OPTIONS
	allow := w.Header().Get("Allow")
	if allow == "" {
		t.Fatalf("Expected Allow header to be set")
	}

	// Parse Allow header and check for expected methods
	expectedMethods := []string{"GET", "HEAD", "OPTIONS"}
	for _, method := range expectedMethods {
		if !containsMethod(allow, method) {
			t.Fatalf("Expected Allow header to contain %s, got %s", method, allow)
		}
	}
}

// containsMethod checks if a method is present in the Allow header
func containsMethod(allow, method string) bool {
	// Parse the Allow header by splitting on commas and checking each method
	methods := strings.Split(allow, ",")
	for _, m := range methods {
		if strings.TrimSpace(m) == method {
			return true
		}
	}
	return false
}
