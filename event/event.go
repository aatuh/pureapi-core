package event

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// EventType represents the type of event.
type EventType string

// Event represents an emitted event.
type Event struct {
	Type    EventType
	Message string
	Data    any
}

// NewEvent creates a new event.
//
// Parameters:
//   - eventType: The type of the event.
//   - message: The message of the event.
//   - data: The optional data of the event.
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
}

// eventListener wraps a listener callback with an ID.
type eventListener struct {
	id       string
	callback func(*Event)
}

// EventEmitter is responsible for emitting events.
type DefaultEventEmitter struct {
	listeners       map[EventType][]eventListener
	globalListeners []eventListener
	mu              sync.RWMutex   // Mutex for thread safety when emitting events.
	counter         int            // Used to generate unique IDs for listeners.
	timeout         *time.Duration // Optional timeout for each callback.
}

// DefaultEventEmitter implements the EventEmitter interface.
var _ EventEmitter = (*DefaultEventEmitter)(nil)

// NewEventEmitter creates a new DefaultEventEmitter.
//
// Parameters:
//   - opts: Options to configure the DefaultEventEmitter.
//
// Returns:
//   - *DefaultEventEmitter: A new DefaultEventEmitter.
func NewEventEmitter() *DefaultEventEmitter {
	eventEmitter := &DefaultEventEmitter{
		listeners:       make(map[EventType][]eventListener),
		globalListeners: []eventListener{},
		mu:              sync.RWMutex{},
		counter:         0,
		timeout:         nil,
	}
	return eventEmitter
}

// WithTimeout sets the timeout for each callback. If the timeout is exceeded,
// an error message will be printed to stderr. It will return a new
// eventEmitterOption.
//
// Parameters:
//   - timeout: The timeout duration.
//
// Returns:
//   - *DefaultEventEmitter: A new DefaultEventEmitter.
func (e *DefaultEventEmitter) WithTimeout(
	timeout *time.Duration,
) *DefaultEventEmitter {
	new := NewEventEmitter()
	new.timeout = timeout
	return new
}

// RegisterListener registers a listener for a specific event type.
//
// Parameters:
//   - eventType: The type of the event.
//   - callback: The function to call when the event is emitted.
//
// Returns:
//   - *eventEmitter: The eventEmitter.
func (e *DefaultEventEmitter) RegisterListener(
	eventType EventType, callback EventCallback,
) EventEmitter {
	// Generate a unique ID for the listener.
	e.mu.Lock()
	defer e.mu.Unlock()
	e.counter++
	id := fmt.Sprintf("%s-%d", eventType, e.counter)

	// Add the listener to the list.
	e.listeners[eventType] = append(e.listeners[eventType], eventListener{
		id:       id,
		callback: callback,
	})
	return e
}

// RemoveListener removes a listener for a specific event type.
//
// Parameters:
//   - eventType: The type of the event.
//   - listener: The listener function.
//
// Returns:
//   - *eventEmitter: The eventEmitter.
func (e *DefaultEventEmitter) RemoveListener(
	eventType EventType, id string,
) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if list, found := e.listeners[eventType]; found {
		for i, l := range list {
			if l.id == id {
				// Remove the listener with the matching ID.
				e.listeners[eventType] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}

// RegisterGlobalListener registers a callback to listen to all events.
//
// Parameters:
//   - callback: The function to call when an event is emitted.
//
// Returns:
//   - *eventEmitter: The eventEmitter.
func (e *DefaultEventEmitter) RegisterGlobalListener(
	callback EventCallback,
) EventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.counter++
	id := fmt.Sprintf("global-%d", e.counter)
	e.globalListeners = append(e.globalListeners, eventListener{
		id:       id,
		callback: callback,
	})
	return e
}

// RemoveGlobalListener removes a global listener based on its ID.
//
// Parameters:
//   - id: The ID of the listener to remove.
//
// Returns:
//   - *eventEmitter: The eventEmitter.
func (e *DefaultEventEmitter) RemoveGlobalListener(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.globalListeners {
		if l.id == id {
			e.globalListeners = append(e.globalListeners[:i],
				e.globalListeners[i+1:]...)
			break
		}
	}
}

// Emit emits an event to all registered listeners. It runs each callback in a
// separate goroutine. If timeout is set for the eventEmitter, the callbacks
// will be run with the specified timeout. If the timeout is exceeded, an error
// message will be printed to stderr.
//
// Parameters:
//   - event: The event to emit.
//
// Returns:
//   - *eventEmitter: The eventEmitter.
func (e *DefaultEventEmitter) Emit(event *Event) {
	e.mu.RLock()
	listeners := e.listeners[event.Type]
	globalListeners := e.globalListeners
	e.mu.RUnlock()
	// Determine the timeout for each callback.
	var timeout *time.Duration
	if e.timeout != nil {
		timeout = new(time.Duration)
		*timeout = *e.timeout
	}
	// Emit to type-specific listeners.
	for _, l := range listeners {
		go func(cb EventCallback, to *time.Duration) {
			runCallback(event, cb, to)
		}(l.callback, timeout)
	}
	// Emit to global listeners.
	for _, l := range globalListeners {
		go func(cb EventCallback, to *time.Duration) {
			runCallback(event, cb, to)
		}(l.callback, timeout)
	}
}

// runCallback runs a callback with an optional timeout.
func runCallback(
	event *Event, cb EventCallback, timeout *time.Duration,
) {
	done := make(chan struct{})
	go func() {
		cb(event)
		close(done)
	}()
	if timeout != nil {
		select {
		case <-done:
			// Callback completed within the timeout.
		case <-time.After(*timeout):
			// Timeout reached; the callback might still be in the background.
			fmt.Fprintf(
				os.Stderr,
				"Callback for event %v timed out after %v, event type: %v\n",
				event.Type,
				*timeout,
				event.Type,
			)
		}
	}
}
