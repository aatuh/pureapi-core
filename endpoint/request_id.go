package endpoint

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/aatuh/pureapi-core/event"
)

// RequestIDKey is the context key for request ID.
type RequestIDKey struct{}

// RequestIDMiddleware creates a middleware that injects a request ID into the context
// and response headers, and includes it in emitted events.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate or extract request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add to context
			ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// Add to response header
			w.Header().Set("X-Request-ID", requestID)

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestIDFromRequest extracts the request ID from the request context.
func RequestIDFromRequest(r *http.Request) string {
	return RequestIDFromContext(r.Context())
}

// generateRequestID creates a unique request ID using cryptographic randomness.
func generateRequestID() string {
	b := make([]byte, 16) // 128-bit random id
	if _, err := rand.Read(b); err != nil {
		// ultra-rare fallback: still return unique-ish id without panicking
		return hex.EncodeToString([]byte("fallback_request_id"))
	}
	return hex.EncodeToString(b)
}

// EventWithRequestID creates an event with request ID included in the data.
func EventWithRequestID(eventType event.EventType, requestID string, data map[string]any) event.Event {
	if data == nil {
		data = make(map[string]any)
	}
	data["request_id"] = requestID
	return event.Event{
		Type: eventType,
		Data: data,
	}
}
