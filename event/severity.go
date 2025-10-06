package event

// Severity levels for events
const (
	SeverityDebug = "debug"
	SeverityInfo  = "info"
	SeverityWarn  = "warn"
	SeverityError = "error"
	SeverityFatal = "fatal"
	SeverityTrace = "trace"
)

// SeverityEvent represents an event with severity information
type SeverityEvent struct {
	*Event
	Severity string
}

// NewSeverityEvent creates a new event with severity
func NewSeverityEvent(eventType EventType, message string,
	severity string) *SeverityEvent {
	return &SeverityEvent{
		Event: &Event{
			Type:    eventType,
			Message: message,
			Data:    nil,
		},
		Severity: severity,
	}
}

// WithSeverity sets the severity of the event
func (e *SeverityEvent) WithSeverity(severity string) *SeverityEvent {
	new := *e
	new.Severity = severity
	return &new
}

// SeverityEmitter is an interface that can emit events with severity
type SeverityEmitter interface {
	EventEmitter
	EmitDebug(eventType EventType, message string)
	EmitInfo(eventType EventType, message string)
	EmitWarn(eventType EventType, message string)
	EmitError(eventType EventType, message string)
	EmitFatal(eventType EventType, message string)
	EmitTrace(eventType EventType, message string)
}

// DefaultSeverityEmitter implements SeverityEmitter
type DefaultSeverityEmitter struct {
	EventEmitter
}

// NewDefaultSeverityEmitter creates a new default severity emitter
func NewDefaultSeverityEmitter(emitter EventEmitter) SeverityEmitter {
	return &DefaultSeverityEmitter{
		EventEmitter: emitter,
	}
}

// EmitDebug emits a debug level event
func (e *DefaultSeverityEmitter) EmitDebug(eventType EventType,
	message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityDebug)
	e.Emit(severityEvent.Event)
}

// EmitInfo emits an info level event
func (e *DefaultSeverityEmitter) EmitInfo(eventType EventType, message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityInfo)
	e.Emit(severityEvent.Event)
}

// EmitWarn emits a warning level event
func (e *DefaultSeverityEmitter) EmitWarn(eventType EventType, message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityWarn)
	e.Emit(severityEvent.Event)
}

// EmitError emits an error level event
func (e *DefaultSeverityEmitter) EmitError(eventType EventType,
	message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityError)
	e.Emit(severityEvent.Event)
}

// EmitFatal emits a fatal level event
func (e *DefaultSeverityEmitter) EmitFatal(eventType EventType,
	message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityFatal)
	e.Emit(severityEvent.Event)
}

// EmitTrace emits a trace level event
func (e *DefaultSeverityEmitter) EmitTrace(eventType EventType,
	message string) {
	severityEvent := NewSeverityEvent(eventType, message, SeverityTrace)
	e.Emit(severityEvent.Event)
}
