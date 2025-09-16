package endpoint

import (
	"fmt"
	"net/http"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/aatuh/pureapi-core/event"
)

// Constants for event types.
const (
	// EventError is emitted when an error occurs during request processing.
	EventError event.EventType = "event_error"

	// EventOutputError event is emitted when an output error occurs.
	EventOutputError event.EventType = "event_output_error"
)

// InputHandler defines how to process the request input.
type InputHandler[Input any] interface {
	Handle(w http.ResponseWriter, r *http.Request) (*Input, error)
}

// Handler is an interface for handling endpoints.
type Handler[Input any] interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// ErrorHandler handles apierror and maps them to appropriate HTTP responses.
type ErrorHandler interface {
	Handle(err error) (int, apierror.APIError)
}

// OutputHandler processes and writes the endpoint response.
type OutputHandler interface {
	Handle(
		w http.ResponseWriter,
		r *http.Request,
		out any,
		outputError error,
		statusCode int,
	) error
}

// HandlerLogicFn is a function for handling endpoint logic.
type HandlerLogicFn[Input any] func(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error)

// DefaultHandler represents an endpoint with input, business logic, and
// output.
type DefaultHandler[Input any] struct {
	inputHandler   InputHandler[Input]
	handlerLogicFn HandlerLogicFn[Input]
	errorHandler   ErrorHandler
	outputHandler  OutputHandler
	emitterLogger  event.EmitterLogger
}

// NewHandler creates a new handler. During requst handling it
// executes common endpoints logic. It calls the input handler, handler
// logic, and output handler. If an error occurs during output handling, it
// will write a 500 error.
//
// Parameters:
//   - inputHandler: The input handler.
//   - handlerLogicFn: The handler logic function.
//   - errorHandler: The error handler.
//   - outputHandler: The output handler.
//
// Returns:
//   - *DefaultHandler: A new DefaultHandler instance.
func NewHandler[Input any](
	inputHandler InputHandler[Input],
	handlerLogicFn HandlerLogicFn[Input],
	errorHandler ErrorHandler,
	outputHandler OutputHandler,
) *DefaultHandler[Input] {
	return &DefaultHandler[Input]{
		inputHandler:   inputHandler,
		handlerLogicFn: handlerLogicFn,
		errorHandler:   errorHandler,
		outputHandler:  outputHandler,
		emitterLogger:  defaultEmitterLogger(),
	}
}

// WithInputHandler adds an input handler to the handler and returns a new
// handler instance.
//
// Parameters:
//   - inputHandler: The input handler.
//
// Returns:
//   - *DefaultHandler: A new handler instance.
func (h *DefaultHandler[Input]) WithInputHandler(
	inputHandler InputHandler[Input],
) *DefaultHandler[Input] {
	new := *h
	new.inputHandler = inputHandler
	return &new
}

// WithHandlerLogicFn adds a handler logic function to the handler and returns a
// new handler instance.
//
// Parameters:
//   - handlerLogicFn: The handler logic function.
//
// Returns:
//   - *DefaultHandler: A new handler instance.
func (h *DefaultHandler[Input]) WithHandlerLogicFn(
	handlerLogicFn HandlerLogicFn[Input],
) *DefaultHandler[Input] {
	new := *h
	new.handlerLogicFn = handlerLogicFn
	return &new
}

// WithErrorHandler adds an error handler to the handler and returns a new
// handler instance.
//
// Parameters:
//   - errorHandler: The error handler.
//
// Returns:
//   - *DefaultHandler: A new handler instance.
func (h *DefaultHandler[Input]) WithErrorHandler(
	errorHandler ErrorHandler,
) *DefaultHandler[Input] {
	new := *h
	new.errorHandler = errorHandler
	return &new
}

// WithEmitterLogger adds an emitter logger to the handler and returns a new
// handler instance.
//
// Parameters:
//   - emitterLogger: The emitter logger.
//
// Returns:
//   - *DefaultHandler: A new handler instance.
func (h *DefaultHandler[Input]) WithEmitterLogger(
	emitterLogger event.EmitterLogger,
) *DefaultHandler[Input] {
	new := *h
	if new.emitterLogger == nil {
		new.emitterLogger = defaultEmitterLogger()
	} else {
		new.emitterLogger = emitterLogger
	}
	return &new
}

// Handle executes common endpoints logic. It calls the input handler, handler
// logic, and output handler.
//
// Parameters:
//   - w: The HTTP response writer.
//   - r: The HTTP request.
func (h *DefaultHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request,
) {
	// Handle input.
	input, err := h.inputHandler.Handle(w, r)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	// Call handler logic.
	out, err := h.handlerLogicFn(w, r, input)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	// Write output.
	h.handleOutput(w, r, out, nil, http.StatusOK)
}

// handleError maps apierror and writes the error response.
func (h *DefaultHandler[Input]) handleError(
	w http.ResponseWriter, r *http.Request, err error,
) {
	// Handle error.
	statusCode, outError := h.errorHandler.Handle(err)
	h.emitterLogger.Trace(
		event.NewEvent(
			EventError,
			fmt.Sprintf(
				"Error, status: %d, err: %s, out: %s",
				statusCode,
				err,
				outError,
			),
		).WithData(
			map[string]any{"status": statusCode, "err": err, "out": outError},
		),
		r.Context(),
	)
	// Handle and write output.
	h.handleOutput(w, r, nil, outError, statusCode)
}

// handleOutput processes and writes the endpoint response.
func (h *DefaultHandler[Input]) handleOutput(
	w http.ResponseWriter, r *http.Request, out any, outError error, status int,
) {
	if err := h.outputHandler.Handle(
		w, r, out, outError, status,
	); err != nil {
		// If error handling output, write 500 error.
		h.emitterLogger.Trace(
			event.NewEvent(
				EventOutputError,
				fmt.Sprintf("Error handling output: %+v", err),
			).WithData(map[string]any{"err": err}),
			r.Context(),
		)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// defaultEmitterLogger returns a noop emitter logger.
func defaultEmitterLogger() event.EmitterLogger {
	return event.NewNoopEmitterLogger()
}
