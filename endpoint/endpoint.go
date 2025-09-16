package endpoint

import (
	"net/http"
)

// Endpoint represents an API endpoint with middlewares.
type Endpoint interface {
	URL() string
	Method() string
	Middlewares() Middlewares
	Handler() http.HandlerFunc
	WithURL(string) Endpoint
	WithMethod(string) Endpoint
	WithMiddlewares(Middlewares) Endpoint
	WithHandler(http.HandlerFunc) Endpoint
}

// DefaultEndpoint represents an API endpoint with middlewares.
type DefaultEndpoint struct {
	URLVal         string
	MethodVal      string
	MiddlewaresVal Middlewares
	HandlerVal     http.HandlerFunc // Optional handler for the endpoint.
}

// defaultEndpoint implements the Endpoint interface.
var _ Endpoint = (*DefaultEndpoint)(nil)

// NewEndpoint creates a new defaultEndpoint with the given details.
//
// Parameters:
//   - url: The URL of the endpoint.
//   - method: The HTTP method of the endpoint.
//   - middlewares: The middlewares to apply to the endpoint.
//
// Returns:
//   - *defaultEndpoint: A new defaultEndpoint instance.
func NewEndpoint(url string, method string) *DefaultEndpoint {
	return &DefaultEndpoint{
		URLVal:         url,
		MethodVal:      method,
		MiddlewaresVal: nil,
		HandlerVal:     nil,
	}
}

// URL returns the URL of the endpoint.
//
// Returns:
//   - string: The URL of the endpoint.
func (e *DefaultEndpoint) URL() string {
	return e.URLVal
}

// Method returns the HTTP method of the endpoint.
//
// Returns:
//   - string: The HTTP method of the endpoint.
func (e *DefaultEndpoint) Method() string {
	return e.MethodVal
}

// Middlewares returns the middlewares of the endpoint. If no middlewares are
// set, it returns an empty Middlewares instance.
//
// Returns:
//   - Middlewares: The middlewares of the endpoint.
func (e *DefaultEndpoint) Middlewares() Middlewares {
	if e.MiddlewaresVal == nil {
		return NewMiddlewares()
	}
	return e.MiddlewaresVal
}

// Handler returns the handler of the endpoint.
//
// Returns:
//   - http.HandlerFunc: The handler of the endpoint.
func (e *DefaultEndpoint) Handler() http.HandlerFunc {
	return e.HandlerVal
}

// WithURL sets the URL of the endpoint. It returns a new endpoint.
//
// Parameters:
//   - url: The URL of the endpoint.
//
// Returns:
//   - Endpoint: A new Endpoint.
func (e *DefaultEndpoint) WithURL(url string) Endpoint {
	new := *e
	new.URLVal = url
	return &new
}

// WithMethod sets the method of the endpoint. It returns a new endpoint.
//
// Parameters:
//   - method: The method of the endpoint.
//
// Returns:
//   - Endpoint: A new Endpoint.
func (e *DefaultEndpoint) WithMethod(method string) Endpoint {
	new := *e
	new.MethodVal = method
	return &new
}

// WithMiddlewares sets the middlewares for the endpoint. It returns a new
// endpoint.
//
// Parameters:
//   - middlewares: The middlewares to apply to the endpoint.
//
// Returns:
//   - Endpoint: A new Endpoint.
func (e *DefaultEndpoint) WithMiddlewares(middlewares Middlewares) Endpoint {
	new := *e
	new.MiddlewaresVal = middlewares
	return &new
}

// WithHandler sets the handler for the endpoint. It returns a new endpoint.
//
// Parameters:
//   - handler: The handler for the endpoint.
//
// Returns:
//   - Endpoint: A new Endpoint.
func (e *DefaultEndpoint) WithHandler(handler http.HandlerFunc) Endpoint {
	new := *e
	new.HandlerVal = handler
	return &new
}
