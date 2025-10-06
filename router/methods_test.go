package router

import (
	"net/http"
	"reflect"
	"testing"
)

func TestBuiltinRouter_MethodsFor(t *testing.T) {
	r := NewBuiltinRouter()

	register := func(method, pattern string) {
		t.Helper()
		if err := r.Register(method, pattern, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})); err != nil {
			t.Fatalf("Register(%s %s) returned error: %v", method, pattern, err)
		}
	}

	register("GET", "/users")
	register("POST", "/users")
	register("POST", "/forms")
	register("GET", "/users/:id")
	register("PATCH", "/users/:id")
	register("LINK", "/users/:id")
	register("GET", "/teams/{teamId}")

	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "exact",
			path: "/users",
			want: []string{"OPTIONS", "GET", "HEAD", "POST"},
		},
		{
			name: "param colon",
			path: "/users/123",
			want: []string{"OPTIONS", "GET", "HEAD", "PATCH", "LINK"},
		},
		{
			name: "braces param",
			path: "/teams/abc",
			want: []string{"OPTIONS", "GET", "HEAD"},
		},
		{
			name: "post only",
			path: "/forms",
			want: []string{"OPTIONS", "POST"},
		},
		{
			name: "no routes",
			path: "/missing",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.MethodsFor(tt.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("MethodsFor(%s) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
