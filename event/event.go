package event

// EventType represents the type of event.
type EventType string

// Event represents an emitted event.
type Event struct {
	Type    EventType
	Message string
	Data    any
}

// WithData sets the data of the event. It returns a new event with the data
// set.
//
// Parameters:
//   - data: The data to set.
//
// Returns:
//   - *Event: A new Event instance with the data set.
func (event *Event) WithData(data any) *Event {
	new := *event
	new.Data = data
	return &new
}

// EventCallback is a function that handles an event.
type EventCallback func(event *Event)

// EventEmitter is responsible for emitting events.
type EventEmitter interface {
	RegisterListener(eventType EventType, callback EventCallback) EventEmitter
	RemoveListener(eventType EventType, id string)
	Emit(event *Event)
	RegisterGlobalListener(callback EventCallback) EventEmitter
	RemoveGlobalListener(id string)
}

// NewEvent creates a new event.
//
// Parameters:
//   - eventType: The type of the event.
//   - message: The message of the event.
//
// Returns:
//   - *Event: A new Event instance.
func NewEvent(eventType EventType, message string) *Event {
	return &Event{
		Type:    eventType,
		Message: message,
		Data:    nil,
	}
}

// NoopEventEmitter is a no-op implementation of EventEmitter.
type NoopEventEmitter struct{}

// NewNoopEventEmitter creates a new noop event emitter.
func NewNoopEventEmitter() *NoopEventEmitter {
	return &NoopEventEmitter{}
}

// RegisterListener does nothing.
func (n *NoopEventEmitter) RegisterListener(eventType EventType,
	callback EventCallback) EventEmitter {
	return n
}

// RemoveListener does nothing.
func (n *NoopEventEmitter) RemoveListener(eventType EventType, id string) {}

// Emit does nothing.
func (n *NoopEventEmitter) Emit(event *Event) {}

// RegisterGlobalListener does nothing.
func (n *NoopEventEmitter) RegisterGlobalListener(
	callback EventCallback) EventEmitter {
	return n
}

// RemoveGlobalListener does nothing.
func (n *NoopEventEmitter) RemoveGlobalListener(id string) {}

// NewEmitterLogger creates a new event emitter.
// This is a placeholder function that returns a noop emitter.
// In a real implementation, this would create a proper event emitter.
func NewEmitterLogger(eventEmitter EventEmitter,
	loggerFactoryFn func(params ...any) any) EventEmitter {
	return NewNoopEventEmitter()
}

// NewEventEmitter creates a new event emitter.
// This is a placeholder function that returns a noop emitter.
// In a real implementation, this would create a proper event emitter.
func NewEventEmitter() EventEmitter {
	return NewNoopEventEmitter()
}
