package endpoint

import (
	"net/http"
)

// ToEndpoints converts a list of endpoint specifications to a list of API endpoints.
// This is a simple helper function that replaces the Definition interface.
//
// Parameters:
//   - specs: The endpoint specifications to convert.
//
// Returns:
//   - []Endpoint: A list of API endpoints.
func ToEndpoints(specs ...EndpointSpec) []Endpoint {
	endpoints := []Endpoint{}
	for _, spec := range specs {
		if spec == nil {
			continue // Skip nil specifications
		}
		endpoints = append(endpoints, spec.ToEndpoint())
	}
	return endpoints
}

// EndpointSpec represents a specification for creating an endpoint.
type EndpointSpec interface {
	ToEndpoint() Endpoint
}

// SimpleEndpointSpec is a simple implementation of EndpointSpec.
type SimpleEndpointSpec struct {
	URLVal         string
	MethodVal      string
	MiddlewaresVal Middlewares
	HandlerVal     http.HandlerFunc
}

// ToEndpoint converts the specification to an Endpoint.
func (s *SimpleEndpointSpec) ToEndpoint() Endpoint {
	return NewEndpoint(s.URLVal, s.MethodVal).
		WithMiddlewares(s.MiddlewaresVal).
		WithHandler(s.HandlerVal)
}

// NewEndpointSpec creates a new endpoint specification.
func NewEndpointSpec(url, method string, middlewares Middlewares, handler http.HandlerFunc) *SimpleEndpointSpec {
	return &SimpleEndpointSpec{
		URLVal:         defaultURL(url),
		MethodVal:      method,
		MiddlewaresVal: middlewares,
		HandlerVal:     handler,
	}
}

// defaultURL returns the default URL if the URL is empty.
func defaultURL(url string) string {
	if url == "" {
		return "/"
	}
	return url
}
