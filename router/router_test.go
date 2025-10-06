package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuiltinRouter_Register(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	err := router.Register("GET", "/test", handler)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestBuiltinRouter_Match_Exact(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("exact"))
	})

	router.Register("GET", "/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	matched := router.Match(req)

	if matched == nil {
		t.Fatal("Expected match, got nil")
	}

	if matched.Handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	if len(matched.Params) != 0 {
		t.Fatalf("Expected no params, got %v", matched.Params)
	}
}

func TestBuiltinRouter_Match_WithParams(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("param"))
	})

	router.Register("GET", "/user/:id", handler)

	req := httptest.NewRequest("GET", "/user/123", nil)
	matched := router.Match(req)

	if matched == nil {
		t.Fatal("Expected match, got nil")
	}

	if matched.Handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	expectedParams := Params{"id": "123"}
	if len(matched.Params) != len(expectedParams) {
		t.Fatalf("Expected %d params, got %d", len(expectedParams), len(matched.Params))
	}

	if matched.Params["id"] != "123" {
		t.Fatalf("Expected param 'id' to be '123', got '%s'", matched.Params["id"])
	}
}

func TestBuiltinRouter_Match_NoMatch(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	router.Register("GET", "/test", handler)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	matched := router.Match(req)

	if matched != nil {
		t.Fatal("Expected no match, got match")
	}
}

func TestBuiltinRouter_Match_WrongMethod(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	router.Register("GET", "/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	matched := router.Match(req)

	if matched != nil {
		t.Fatal("Expected no match for wrong method, got match")
	}
}

func TestBuiltinRouter_Match_MultipleParams(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("multi"))
	})

	router.Register("GET", "/user/:id/post/:postId", handler)

	req := httptest.NewRequest("GET", "/user/123/post/456", nil)
	matched := router.Match(req)

	if matched == nil {
		t.Fatal("Expected match, got nil")
	}

	expectedParams := Params{"id": "123", "postId": "456"}
	if len(matched.Params) != len(expectedParams) {
		t.Fatalf("Expected %d params, got %d", len(expectedParams), len(matched.Params))
	}

	if matched.Params["id"] != "123" {
		t.Fatalf("Expected param 'id' to be '123', got '%s'", matched.Params["id"])
	}

	if matched.Params["postId"] != "456" {
		t.Fatalf("Expected param 'postId' to be '456', got '%s'", matched.Params["postId"])
	}
}

func TestBuiltinRouter_Match_WithBracesParams(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("braces"))
	})

	router.Register("GET", "/user/{id}", handler)

	req := httptest.NewRequest("GET", "/user/123", nil)
	matched := router.Match(req)

	if matched == nil {
		t.Fatal("Expected match, got nil")
	}

	if matched.Handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	expectedParams := Params{"id": "123"}
	if len(matched.Params) != len(expectedParams) {
		t.Fatalf("Expected %d params, got %d", len(expectedParams), len(matched.Params))
	}

	if matched.Params["id"] != "123" {
		t.Fatalf("Expected param 'id' to be '123', got '%s'", matched.Params["id"])
	}
}

func TestBuiltinRouter_Match_MixedParamSyntax(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mixed"))
	})

	router.Register("GET", "/user/{id}/post/:postId", handler)

	req := httptest.NewRequest("GET", "/user/123/post/456", nil)
	matched := router.Match(req)

	if matched == nil {
		t.Fatal("Expected match, got nil")
	}

	expectedParams := Params{"id": "123", "postId": "456"}
	if len(matched.Params) != len(expectedParams) {
		t.Fatalf("Expected %d params, got %d", len(expectedParams), len(matched.Params))
	}

	if matched.Params["id"] != "123" {
		t.Fatalf("Expected param 'id' to be '123', got '%s'", matched.Params["id"])
	}

	if matched.Params["postId"] != "456" {
		t.Fatalf("Expected param 'postId' to be '456', got '%s'", matched.Params["postId"])
	}
}

func TestBuiltinRouter_Unregister(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Register a route
	err := router.Register("GET", "/test", handler)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it matches
	req := httptest.NewRequest("GET", "/test", nil)
	matched := router.Match(req)
	if matched == nil {
		t.Fatal("Expected match before unregister, got nil")
	}

	// Unregister the route
	err = router.Unregister("GET", "/test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it no longer matches
	matched = router.Match(req)
	if matched != nil {
		t.Fatal("Expected no match after unregister, got match")
	}
}

func TestBuiltinRouter_Unregister_WithParams(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("param"))
	})

	// Register a route with parameters
	err := router.Register("GET", "/user/:id", handler)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it matches
	req := httptest.NewRequest("GET", "/user/123", nil)
	matched := router.Match(req)
	if matched == nil {
		t.Fatal("Expected match before unregister, got nil")
	}

	// Unregister the route
	err = router.Unregister("GET", "/user/:id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it no longer matches
	matched = router.Match(req)
	if matched != nil {
		t.Fatal("Expected no match after unregister, got match")
	}
}

func TestBuiltinRouter_Unregister_WithBraces(t *testing.T) {
	router := NewBuiltinRouter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("braces"))
	})

	// Register a route with braces parameters
	err := router.Register("GET", "/user/{id}", handler)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it matches
	req := httptest.NewRequest("GET", "/user/123", nil)
	matched := router.Match(req)
	if matched == nil {
		t.Fatal("Expected match before unregister, got nil")
	}

	// Unregister the route
	err = router.Unregister("GET", "/user/{id}")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it no longer matches
	matched = router.Match(req)
	if matched != nil {
		t.Fatal("Expected no match after unregister, got match")
	}
}
