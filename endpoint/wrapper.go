package endpoint

// DefaultWrapper encapsulates a middleware with an identifier and optional
// metadata. ID can be used to identify the middleware type (e.g. for reordering
// or documentation). Data can carry any type of additional information.
type DefaultWrapper struct {
	id         string
	middleware Middleware
	data       any
}

// DefaultWrapper implements the Wrapper interface.
var _ Wrapper = (*DefaultWrapper)(nil)

// NewWrapper creates a new middleware DefaultWrapper.
//
// Parameters:
//   - id: The ID of the wrapper.
//   - middleware: The middleware to wrap.
//
// Returns:
//   - *DefaultWrapper: A new DefaultWrapper instance.
func NewWrapper(
	id string, middleware Middleware,
) *DefaultWrapper {
	w := &DefaultWrapper{
		id:         id,
		middleware: middleware,
		data:       nil,
	}
	return w
}

// WithData returns a new DefaultWrapper with the given data.
//
// Parameters:
//   - data: The data to attach to the wrapper.
//
// Returns:
//   - *DefaultWrapper: A new DefaultWrapper instance.
func (m *DefaultWrapper) WithData(data any) *DefaultWrapper {
	new := *m
	new.data = data
	return &new
}

// Middleware returns the middleware wrapped by this wrapper.
//
// Returns:
//   - Middleware: The wrapped middleware.
func (m *DefaultWrapper) Middleware() Middleware {
	return m.middleware
}

// ID returns the ID of the wrapper.
//
// Returns:
//   - string: The ID of the wrapper.
func (m *DefaultWrapper) ID() string {
	return m.id
}

// Data returns the data attached to the wrapper.
//
// Returns:
//   - any: The data attached to the wrapper.
func (m *DefaultWrapper) Data() any {
	return m.data
}
