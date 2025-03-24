package endpoint

import (
	"fmt"
	"net/http"

	"github.com/pureapi/pureapi-core/endpoint/types"
)

// DefaultDefinition represents an endpoint definition.
type DefaultDefinition struct {
	url     string
	method  string
	stack   types.Stack
	handler http.HandlerFunc
}

var _ types.Definition = (*DefaultDefinition)(nil)

// NewDefinition creates a new endpoint definition.
//
// Parameters:
//   - url: The URL of the endpoint. Defaults to "/" if empty.
//   - method: The HTTP method of the endpoint.
//   - stack: The middleware stack for the endpoint.
//   - handler: The handler for the endpoint.
//
// Returns:
//   - *DefaultDefinition: A new DefaultDefinition instance.
func NewDefinition(
	url string,
	method string,
	stack types.Stack,
	handler http.HandlerFunc,
) *DefaultDefinition {
	return &DefaultDefinition{
		url:     defaultURL(url),
		method:  method,
		stack:   stack,
		handler: handler,
	}
}

// URL returns the URL of the endpoint.
//
// Returns:
//   - string: The URL of the endpoint.
func (d *DefaultDefinition) URL() string {
	return d.url
}

// Method returns the HTTP method of the endpoint.
//
// Returns:
//   - string: The HTTP method of the endpoint.
func (d *DefaultDefinition) Method() string {
	return d.method
}

// Stack returns the middleware stack of the endpoint.
//
// Returns:
//   - Stack: The middleware stack of the endpoint.
func (d *DefaultDefinition) Stack() types.Stack {
	return d.stack
}

// Handler returns the handler of the endpoint.
//
// Returns:
//   - http.HandlerFunc: The handler of the endpoint.
func (d *DefaultDefinition) Handler() http.HandlerFunc {
	return d.handler
}

// Clone creates a deep copy of an endpoint definition and returns the cloned
// endpoint definition.
//
// Parameters:
//   - opts: Options to apply to the cloned definition.
//
// Returns:
//   - *Definition: the cloned definition.
func (d *DefaultDefinition) Clone() *DefaultDefinition {
	cloned := *d
	if d.stack != nil {
		cloned.stack = d.stack.Clone()
	}
	return &cloned
}

// WithURL sets the URL of the endpoint. Defaults to "/" if empty. It returns a
// new endpoint definition.
//
// Parameters:
//   - url: The URL of the endpoint.
//
// Returns:
//   - *DefaultDefinition: A new DefaultDefinition instance.
func (d *DefaultDefinition) WithURL(url string) *DefaultDefinition {
	new := *d
	new.url = defaultURL(url)
	return &new
}

// WithMethod sets the method of the endpoint. It returns a new endpoint
// definition.
//
// Parameters:
//   - method: The method of the endpoint.
//
// Returns:
//   - *DefaultDefinition: A new DefaultDefinition instance.
func (d *DefaultDefinition) WithMethod(method string) *DefaultDefinition {
	new := *d
	new.method = method
	return &new
}

// WithHandler sets the handler of the endpoint. It returns a new endpoint
// definition.
//
// Parameters:
//   - handler: The handler of the endpoint.
//
// Returns:
//   - *DefaultDefinition: A new DefaultDefinition instance.
func (d *DefaultDefinition) WithHandler(
	handler http.HandlerFunc,
) *DefaultDefinition {
	new := *d
	new.handler = handler
	return &new
}

// WithMiddlewareStack sets the middleware stack of the endpoint. It returns a
// new endpoint definition.
//
// Parameters:
//   - stack: The middleware stack.
//
// Returns:
//   - *DefaultDefinition: A new DefaultDefinition instance.
func (d *DefaultDefinition) WithMiddlewareStack(
	stack types.Stack,
) *DefaultDefinition {
	new := *d
	new.stack = stack
	return &new
}

// defaultDefinitions is a new list of endpoint definitions.
type defaultDefinitions struct {
	definitions []types.Definition
}

// DefaultDefinition implements the Definitions interface.
var _ types.Definitions = (*defaultDefinitions)(nil)

func NewDefinitions(
	definitions ...types.Definition,
) (*defaultDefinitions, error) {
	for _, definition := range definitions {
		if definition == nil {
			return nil, fmt.Errorf(
				"NewDefinitions: endpoint definition cannot be nil",
			)
		}
	}
	return &defaultDefinitions{
		definitions: definitions,
	}, nil
}

// Add adds new endpoint definitions to the list of endpoint definitions.
//
// Parameters:
//   - definitions: The new endpoint definitions.
//
// Returns:
//   - *Definitions: A new list of endpoint definitions.
func (d defaultDefinitions) Add(
	definitions ...types.Definition,
) (*defaultDefinitions, error) {
	defs := append([]types.Definition{}, d.definitions...)
	defs = append(defs, definitions...)
	return NewDefinitions(defs...)
}

// ToEndpoints converts a list of endpoint definitions to a list of API
// endpoints.
//
// Returns:
//   - []api.Endpoint: a list of API endpoints.
func (d defaultDefinitions) ToEndpoints() []types.Endpoint {
	endpoints := []types.Endpoint{}
	for _, definition := range d.definitions {
		middlewares := []types.Middleware{}
		if definition.Stack() != nil {
			for _, mw := range definition.Stack().Wrappers() {
				middlewares = append(middlewares, mw.Middleware())
			}
		}
		endpoints = append(
			endpoints,
			NewEndpoint(definition.URL(), definition.Method()).
				WithMiddlewares(NewMiddlewares(middlewares...)).
				WithHandler(definition.Handler()),
		)
	}
	return endpoints
}

// defaultURL returns the default URL if the URL is empty.
func defaultURL(url string) string {
	if url == "" {
		return "/"
	} else {
		return url
	}
}
