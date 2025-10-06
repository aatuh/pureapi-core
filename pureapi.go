// Package pureapi provides a unified API for building HTTP applications.
//
// Example:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"net/http"
//		"log"
//
//		"github.com/aatuh/pureapi-core"
//	)
//
//	func main() {
//		// Create a new server
//		server := pureapi.NewServer()
//
//		// Register a simple GET route
//		server.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//			w.Write([]byte("Hello, World!"))
//		})
//
//		// Register a route with parameters
//		server.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//			params := pureapi.RouteParams(r)
//			userID := params["id"]
//
//			w.WriteHeader(http.StatusOK)
//			fmt.Fprintf(w, "User ID: %s", userID)
//		})
//
//		// Start the server
//		log.Println("Server starting on :8080")
//		if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
//			log.Fatal(err)
//		}
//	}
package pureapi

import (
	"net/http"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/querydec"
	"github.com/aatuh/pureapi-core/router"
	"github.com/aatuh/pureapi-core/server"
)

// Server is a small facade over server.Handler with route helpers.
type Server struct {
	h *server.Handler
}

// registeredEndpoint tracks registration updates when mutating endpoint settings.
type registeredEndpoint struct {
	s  *server.Handler
	ep endpoint.Endpoint
}

// URL returns the URL of the registered endpoint.
//
// Returns:
//   - string: The URL of the endpoint.
func (r *registeredEndpoint) URL() string { return r.ep.URL() }

// Method returns the HTTP method of the registered endpoint.
//
// Returns:
//   - string: The HTTP method of the endpoint.
func (r *registeredEndpoint) Method() string { return r.ep.Method() }

// Middlewares returns the middlewares of the registered endpoint.
//
// Returns:
//   - Middlewares: The middlewares of the endpoint.
func (r *registeredEndpoint) Middlewares() Middlewares { return r.ep.Middlewares() }

// Handler returns the handler of the registered endpoint.
//
// Returns:
//   - http.HandlerFunc: The handler of the endpoint.
func (r *registeredEndpoint) Handler() http.HandlerFunc { return r.ep.Handler() }

// WithURL updates the URL of the registered endpoint.
//
// Parameters:
//   - u: The new URL for the endpoint.
//
// Returns:
//   - endpoint.Endpoint: The updated endpoint for chaining.
func (r *registeredEndpoint) WithURL(u string) endpoint.Endpoint {
	oldURL, oldMethod := r.ep.URL(), r.ep.Method()
	r.ep = r.ep.WithURL(u)
	r.s.Unregister(oldMethod, oldURL)
	r.s.Register([]endpoint.Endpoint{r.ep})
	return r
}

// WithMethod updates the HTTP method of the registered endpoint.
//
// Parameters:
//   - m: The new HTTP method for the endpoint.
//
// Returns:
//   - endpoint.Endpoint: The updated endpoint for chaining.
func (r *registeredEndpoint) WithMethod(m string) endpoint.Endpoint {
	oldURL, oldMethod := r.ep.URL(), r.ep.Method()
	r.ep = r.ep.WithMethod(m)
	r.s.Unregister(oldMethod, oldURL)
	r.s.Register([]endpoint.Endpoint{r.ep})
	return r
}

// WithMiddlewares updates the middlewares of the registered endpoint.
//
// Parameters:
//   - m: The new middlewares for the endpoint.
//
// Returns:
//   - endpoint.Endpoint: The updated endpoint for chaining.
func (r *registeredEndpoint) WithMiddlewares(m Middlewares) endpoint.Endpoint {
	oldURL, oldMethod := r.ep.URL(), r.ep.Method()
	r.ep = r.ep.WithMiddlewares(m)
	r.s.Unregister(oldMethod, oldURL)
	r.s.Register([]endpoint.Endpoint{r.ep})
	return r
}

// WithHandler updates the handler of the registered endpoint.
//
// Parameters:
//   - h: The new handler for the endpoint.
//
// Returns:
//   - endpoint.Endpoint: The updated endpoint for chaining.
func (r *registeredEndpoint) WithHandler(h http.HandlerFunc) endpoint.Endpoint {
	oldURL, oldMethod := r.ep.URL(), r.ep.Method()
	r.ep = r.ep.WithHandler(h)
	r.s.Unregister(oldMethod, oldURL)
	r.s.Register([]endpoint.Endpoint{r.ep})
	return r
}

// ServerOption configures the underlying HTTP handler.
type ServerOption = server.HandlerOption

// NewServer constructs a server with optional configuration.
//
// Parameters:
//   - opts: Optional server configuration options.
//
// Returns:
//   - *Server: A new Server instance with the specified configuration.
func NewServer(opts ...ServerOption) *Server {
	h := server.NewHandler(event.NewNoopEventEmitter(), opts...)
	return &Server{h: h}
}

// Handler returns the underlying http.Handler.
//
// Returns:
//   - http.Handler: The underlying HTTP handler.
func (s *Server) Handler() http.Handler { return s.h }

// Get registers a GET route and returns the created endpoint for chaining.
//
// Parameters:
//   - path: The URL path for the route.
//   - fn: The handler function for the route.
//
// Returns:
//   - endpoint.Endpoint: The created endpoint for method chaining.
func (s *Server) Get(path string, fn http.HandlerFunc) endpoint.Endpoint {
	ep := endpoint.NewEndpoint(path, http.MethodGet).WithHandler(fn)
	s.h.Register([]endpoint.Endpoint{ep})
	return &registeredEndpoint{s: s.h, ep: ep}
}

// Post registers a POST route and returns the created endpoint for chaining.
//
// Parameters:
//   - path: The URL path for the route.
//   - fn: The handler function for the route.
//
// Returns:
//   - endpoint.Endpoint: The created endpoint for method chaining.
func (s *Server) Post(path string, fn http.HandlerFunc) endpoint.Endpoint {
	ep := endpoint.NewEndpoint(path, http.MethodPost).WithHandler(fn)
	s.h.Register([]endpoint.Endpoint{ep})
	return &registeredEndpoint{s: s.h, ep: ep}
}

// Put registers a PUT route and returns the created endpoint for chaining.
//
// Parameters:
//   - path: The URL path for the route.
//   - fn: The handler function for the route.
//
// Returns:
//   - endpoint.Endpoint: The created endpoint for method chaining.
func (s *Server) Put(path string, fn http.HandlerFunc) endpoint.Endpoint {
	ep := endpoint.NewEndpoint(path, http.MethodPut).WithHandler(fn)
	s.h.Register([]endpoint.Endpoint{ep})
	return &registeredEndpoint{s: s.h, ep: ep}
}

// Patch registers a PATCH route and returns the created endpoint for chaining.
//
// Parameters:
//   - path: The URL path for the route.
//   - fn: The handler function for the route.
//
// Returns:
//   - endpoint.Endpoint: The created endpoint for method chaining.
func (s *Server) Patch(path string, fn http.HandlerFunc) endpoint.Endpoint {
	ep := endpoint.NewEndpoint(path, http.MethodPatch).WithHandler(fn)
	s.h.Register([]endpoint.Endpoint{ep})
	return &registeredEndpoint{s: s.h, ep: ep}
}

// Delete registers a DELETE route and returns the created endpoint for chaining.
//
// Parameters:
//   - path: The URL path for the route.
//   - fn: The handler function for the route.
//
// Returns:
//   - endpoint.Endpoint: The created endpoint for method chaining.
func (s *Server) Delete(path string, fn http.HandlerFunc) endpoint.Endpoint {
	ep := endpoint.NewEndpoint(path, http.MethodDelete).WithHandler(fn)
	s.h.Register([]endpoint.Endpoint{ep})
	return &registeredEndpoint{s: s.h, ep: ep}
}

// WithRouter sets the router to use.
//
// Parameters:
//   - r: The router implementation to use.
//
// Returns:
//   - ServerOption: A server option function.
func WithRouter(r router.Router) ServerOption { return server.WithRouter(r) }

// WithCustomNotFound sets a custom 404 handler.
//
// Parameters:
//   - h: The custom 404 handler.
//
// Returns:
//   - ServerOption: A server option function.
func WithCustomNotFound(h http.Handler) ServerOption { return server.WithNotFound(h) }

// WithBodyLimit sets maximum request body size in bytes.
//
// Parameters:
//   - limit: The maximum request body size in bytes.
//
// Returns:
//   - ServerOption: A server option function.
func WithBodyLimit(limit int64) ServerOption { return server.WithBodyLimit(limit) }

// WithQueryDecoder sets the query decoder to use.
//
// Parameters:
//   - d: The query decoder to use.
//
// Returns:
//   - ServerOption: A server option function.
func WithQueryDecoder(d querydec.Decoder) ServerOption { return server.WithQueryDecoder(d) }

// WithEventEmitter sets a custom event emitter for the server.
func WithEventEmitter(em event.EventEmitter) ServerOption { return server.WithEventEmitter(em) }

// NewBuiltinRouter exposes the tiny built-in router.
//
// Returns:
//   - router.Router: A new built-in router instance.
func NewBuiltinRouter() router.Router { return router.NewBuiltinRouter() }

// QueryMap exposes query parameters decoded into a map from request context.
//
// Parameters:
//   - r: The HTTP request.
//
// Returns:
//   - map[string]any: The decoded query parameters.
func QueryMap(r *http.Request) map[string]any { return server.QueryMap(r) }

// RouteParams exposes route parameters extracted by the router.
//
// Parameters:
//   - r: The HTTP request.
//
// Returns:
//   - map[string]string: The extracted route parameters.
func RouteParams(r *http.Request) map[string]string { return server.RouteParams(r) }

// InputHandler processes request input into a typed value.
type InputHandler[T any] interface {
	endpoint.InputHandler[T]
}

// HandlerLogicFn performs business logic for an endpoint.
type HandlerLogicFn[T any] func(http.ResponseWriter, *http.Request, *T) (any, error)

// ErrorHandler maps errors to API errors and status codes.
type ErrorHandler = endpoint.ErrorHandler

// OutputHandler writes the response.
type OutputHandler = endpoint.OutputHandler

// NewHandler constructs the default endpoint handler pipeline.
//
// Parameters:
//   - ih: The input handler for processing request input.
//   - lf: The handler logic function for business logic.
//   - eh: The error handler for mapping errors to API responses.
//   - oh: The output handler for writing responses.
//
// Returns:
//   - endpoint.Handler[T]: A new handler instance.
func NewHandler[T any](
	ih InputHandler[T], lf HandlerLogicFn[T], eh ErrorHandler, oh OutputHandler,
) endpoint.Handler[T] {
	return endpoint.NewHandler(
		asEndpointInputHandler(ih),
		endpoint.HandlerLogicFn[T](lf),
		eh,
		oh,
	)
}

func asEndpointInputHandler[T any](ih InputHandler[T]) endpoint.InputHandler[T] {
	if ih == nil {
		return nil
	}
	v, ok := ih.(endpoint.InputHandler[T])
	if !ok {
		panic("pureapi: InputHandler must satisfy endpoint.InputHandler")
	}
	return v
}

// Middleware wraps an http.Handler.
type Middleware = endpoint.Middleware

// Middlewares is a collection of middleware.
type Middlewares = endpoint.Middlewares

// Wrapper wraps a middleware with an ID and optional metadata.
type Wrapper = endpoint.Wrapper

// Stack manages an ordered list of wrappers.
type Stack = endpoint.Stack

// NewStack creates a new middleware stack.
//
// Parameters:
//   - wrappers: The initial middleware wrappers.
//
// Returns:
//   - Stack: A new middleware stack.
func NewStack(wrappers ...Wrapper) Stack { return endpoint.NewStack(wrappers...) }

// NewWrapperFromHandler constructs a wrapper from an http middleware.
//
// Parameters:
//   - id: The identifier for the wrapper.
//   - mw: The HTTP middleware function.
//
// Returns:
//   - Wrapper: A new middleware wrapper.
func NewWrapperFromHandler(id string, mw func(http.Handler) http.Handler) Wrapper {
	return endpoint.NewWrapper(id, Middleware(mw))
}

// APIError represents a structured API error.
type APIError = apierror.APIError

// NewAPIError constructs a new API error with an ID.
//
// Parameters:
//   - id: The error identifier.
//
// Returns:
//   - *apierror.DefaultAPIError: A new API error instance.
func NewAPIError(id string) *apierror.DefaultAPIError { return apierror.NewAPIError(id) }

// APIErrorFrom converts an APIError to its default concrete type.
//
// Parameters:
//   - err: The API error to convert.
//
// Returns:
//   - *apierror.DefaultAPIError: The converted API error.
func APIErrorFrom(err APIError) *apierror.DefaultAPIError { return apierror.APIErrorFrom(err) }
