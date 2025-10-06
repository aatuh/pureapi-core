package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/router"
)

func TestHandler_BodyLimit(t *testing.T) {
	// Test body size limit
	handler := NewHandler(
		event.NewNoopEventEmitter(),
		WithBodyLimit(100), // 100 bytes limit
	)

	testRouter := router.NewBuiltinRouter()
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	testRouter.Register("POST", "/test", testHandler)
	handler.router = testRouter

	// Test with small body (should pass)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("small")))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for small body, got %d", w.Code)
	}

	// Test with large body (should fail)
	largeBody := make([]byte, 200) // 200 bytes, exceeds 100 byte limit
	req = httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req.ContentLength = 200
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("Expected status 413 for large body, got %d", w.Code)
	}
}

func TestHandler_BodyLimit_Zero(t *testing.T) {
	// Test with zero body limit (should allow any size)
	handler := NewHandler(
		event.NewNoopEventEmitter(),
		WithBodyLimit(0), // No limit
	)

	testRouter := router.NewBuiltinRouter()
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	testRouter.Register("POST", "/test", testHandler)
	handler.router = testRouter

	// Test with large body (should pass with no limit)
	largeBody := make([]byte, 1000)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req.ContentLength = 1000
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 with no body limit, got %d", w.Code)
	}
}

func TestHandler_DefaultBodyLimit(t *testing.T) {
	// Test default body limit (2MB)
	handler := NewHandler(event.NewNoopEventEmitter())

	testRouter := router.NewBuiltinRouter()
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	testRouter.Register("POST", "/test", testHandler)
	handler.router = testRouter

	// Test with body within 2MB limit
	body := make([]byte, 1024*1024) // 1MB
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 within default limit, got %d", w.Code)
	}
}
