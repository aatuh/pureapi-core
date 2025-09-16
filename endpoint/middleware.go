package endpoint

import (
	"net/http"
)

// Middleware represents a function that wraps an http.Handler with additional
// behavior. A Middleware typically performs actions before and/or after calling
// the next handler.
//
// Example:
//
//	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    // final processing
//	})
//	wrappedHandler := Apply(finalHandler, middleware1, middleware2)
type Middleware func(http.Handler) http.Handler

// Middlewares is a collection of Middleware functions.
type Middlewares interface {
	Chain(h http.Handler) http.Handler
}

// Wrapper is an interface for a middleware wrapper. It encapsulates a
// Middleware with an identifier and optional metadata.
type Wrapper interface {
	ID() string
	Middleware() Middleware
	Data() any
}

// Stack is an interface for managing a list of middleware wrappers.
type Stack interface {
	Wrappers() []Wrapper
	Middlewares() Middlewares
	Clone() Stack
	AddWrapper(w Wrapper) Stack
	InsertBefore(id string, w Wrapper) (Stack, bool)
	InsertAfter(id string, w Wrapper) (Stack, bool)
	Remove(id string) (Stack, bool)
}

// DefaultMiddlewares is a slice of Middleware functions.
type DefaultMiddlewares struct {
	middlewares []Middleware
}

// DefaultMiddlewares implements the Middlewares interface.
var _ Middlewares = (*DefaultMiddlewares)(nil)

// NewMiddlewares creates a new DefaultMiddlewares instance with the provided
// middlewares.
//
// Parameters:
//   - middlewares: The middlewares to add to the list.
//
// Returns:
//   - *DefaultMiddlewares: A new DefaultMiddlewares instance.
func NewMiddlewares(middlewares ...Middleware) *DefaultMiddlewares {
	return &DefaultMiddlewares{
		middlewares: middlewares,
	}
}

// Middlewares returns the middlewares in the list.
//
// Returns:
//   - []Middleware: The middlewares in the list.
func (m DefaultMiddlewares) Middlewares() []Middleware {
	return m.middlewares
}

// Chain applies a sequence of middlewares to an http.Handler. During a request
// the middlewaress are applied in the order they are provided.
// The middlewares are applied so that the first middleware in the list becomes
// the outermost wrapper.
//
// Example with middlewares m1, m2
//
//	Chain(finalHandler) yields m1(m2(finalHandler)).
//
// Parameters:
//   - h: The http.Handler to wrap.
//
// Returns:
//   - http.Handler: The wrapped http.Handler.
func (m DefaultMiddlewares) Chain(h http.Handler) http.Handler {
	wrapped := h
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		wrapped = m.middlewares[i](wrapped)
	}
	return wrapped
}

// Add adds one or more middlewares to the list and returns a new
// DefaultMiddlewares instance.
//
// Parameters:
//   - middlewares: The middlewares to add to the list.
func (m DefaultMiddlewares) Add(
	middlewares ...Middleware,
) *DefaultMiddlewares {
	allMiddlewares := append([]Middleware{}, m.middlewares...)
	allMiddlewares = append(allMiddlewares, middlewares...)
	return NewMiddlewares(allMiddlewares...)
}
