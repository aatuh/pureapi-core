package event

import "github.com/aatuh/pureapi-core/logging"

// EmitterLogger is an interface that can emit events and log messages.
type EmitterLogger interface {
	Debug(event *Event, factoryParams ...any)
	Info(event *Event, factoryParams ...any)
	Warn(event *Event, factoryParams ...any)
	Error(event *Event, factoryParams ...any)
	Fatal(event *Event, factoryParams ...any)
	Trace(event *Event, factoryParams ...any)
}

// DefaultEmitterLogger is a struct that can emit events and log messages.
type DefaultEmitterLogger struct {
	eventEmitter    EventEmitter
	loggerFactoryFn logging.LoggerFactoryFn
}

// NewEmitterLogger creates a new EmitterLogger.
//
// Parameters:
//   - eventEmitter: An EventEmitter.
//   - loggerFactoryFn: A LoggerFactoryFn.
//
// Returns:
//   - *DefaultEmitterLogger: A new DefaultEmitterLogger instance.
func NewEmitterLogger(
	eventEmitter EventEmitter, loggerFactoryFn logging.LoggerFactoryFn,
) *DefaultEmitterLogger {
	return &DefaultEmitterLogger{
		eventEmitter:    eventEmitter,
		loggerFactoryFn: loggerFactoryFn,
	}
}

// NewNoopEmitterLogger creates a new EmitterLogger that does nothing.
//
// Returns:
//   - *DefaultEmitterLogger: A new DefaultEmitterLogger instance.
func NewNoopEmitterLogger() *DefaultEmitterLogger {
	return &DefaultEmitterLogger{}
}

// Debug emits an event and logs at the Debug level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Debug(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Debug(event.Message)
	}
}

// Trace emits an event and logs at the Trace level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Trace(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Trace(event.Message)
	}
}

// Info emits an event and logs at the Info level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Info(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Info(event.Message)
	}
}

// Warn emits an event and logs at the Warn level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Warn(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Warn(event.Message)
	}
}

// Error emits an event and logs at the Error level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Error(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Error(event.Message)
	}
}

// Fatal emits an event and logs at the Fatal level.
//
// Parameters:
//   - event The event to emit and log.
//   - factoryParams: The parameters to pass to the logger factory function.
func (e *DefaultEmitterLogger) Fatal(event *Event, factoryParams ...any) {
	e.emitIfCan(event)
	if e.loggerFactoryFn != nil {
		e.loggerFactoryFn(factoryParams...).Fatal(event.Message)
	}
}

// emitIfCan emits the event if the event emitter is not nil.
func (e *DefaultEmitterLogger) emitIfCan(event *Event) {
	if e.eventEmitter != nil {
		e.eventEmitter.Emit(event)
	}
}
