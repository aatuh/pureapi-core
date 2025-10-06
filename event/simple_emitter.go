package event

// SimpleSeverityEmitter provides a simple way to emit events with severity
type SimpleSeverityEmitter struct {
	emitter EventEmitter
}

// NewSimpleSeverityEmitter creates a new simple severity emitter
func NewSimpleSeverityEmitter(emitter EventEmitter) *SimpleSeverityEmitter {
	return &SimpleSeverityEmitter{
		emitter: emitter,
	}
}

// EmitWithSeverity emits an event with severity information in the data
func (e *SimpleSeverityEmitter) EmitWithSeverity(eventType EventType,
	message string, severity string) {
	event := &Event{
		Type:    eventType,
		Message: message,
		Data: map[string]any{
			"severity":  severity,
			"timestamp": "now", // You can add actual timestamp here
		},
	}
	e.emitter.Emit(event)
}

// EmitDebug emits a debug level event
func (e *SimpleSeverityEmitter) EmitDebug(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "debug")
}

// EmitInfo emits an info level event
func (e *SimpleSeverityEmitter) EmitInfo(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "info")
}

// EmitWarn emits a warning level event
func (e *SimpleSeverityEmitter) EmitWarn(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "warn")
}

// EmitError emits an error level event
func (e *SimpleSeverityEmitter) EmitError(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "error")
}

// EmitFatal emits a fatal level event
func (e *SimpleSeverityEmitter) EmitFatal(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "fatal")
}

// EmitTrace emits a trace level event
func (e *SimpleSeverityEmitter) EmitTrace(eventType EventType, message string) {
	e.EmitWithSeverity(eventType, message, "trace")
}
