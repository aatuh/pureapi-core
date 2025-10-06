package endpoint

import (
	"fmt"
	"net/http"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/aatuh/pureapi-core/event"
)

// Constants for event event.
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

// DefaultErrorHandler provides a sensible default error mapping.
type DefaultErrorHandler struct{}

// Handle maps errors to appropriate HTTP responses.
// Returns 400 for validation errors, 404 for not found, 500 for others.
func (d DefaultErrorHandler) Handle(err error) (int, apierror.APIError) {
	// Check for specific error types
	if apiErr, ok := err.(apierror.APIError); ok {
		switch apiErr.ID() {
		case "validation_error", "invalid_input":
			return http.StatusBadRequest, apiErr
		case "not_found", "resource_not_found":
			return http.StatusNotFound, apiErr
		case "unauthorized":
			return http.StatusUnauthorized, apiErr
		case "forbidden":
			return http.StatusForbidden, apiErr
		case "conflict":
			return http.StatusConflict, apiErr
		default:
			return http.StatusInternalServerError, apierror.NewAPIError("internal_error").WithMessage("Internal server error")
		}
	}

	// Default to 500 for unknown errors
	return http.StatusInternalServerError, apierror.NewAPIError("internal_error").WithMessage("Internal server error")
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
	emitterLogger  event.EventEmitter
}

// NewHandler creates a new handler. During request handling it
// executes common endpoints logic. It calls the input handler, handler
// logic, and output handler. If an error occurs during output handling, it
// will write a 500 error.
//
// Parameters:
//   - inputHandler: The input handler for processing request input.
//   - handlerLogicFn: The handler logic function for business logic.
//   - errorHandler: The error handler for mapping errors to API responses.
//   - outputHandler: The output handler for writing responses.
//
// Returns:
//   - *DefaultHandler[Input]: A new DefaultHandler instance.
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
//   - inputHandler: The input handler to set.
//
// Returns:
//   - *DefaultHandler[Input]: A new handler instance.
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
//   - handlerLogicFn: The handler logic function to set.
//
// Returns:
//   - *DefaultHandler[Input]: A new handler instance.
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
//   - errorHandler: The error handler to set.
//
// Returns:
//   - *DefaultHandler[Input]: A new handler instance.
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
//   - emitterLogger: The emitter logger to set.
//
// Returns:
//   - *DefaultHandler[Input]: A new handler instance.
func (h *DefaultHandler[Input]) WithEmitterLogger(
	emitterLogger event.EventEmitter,
) *DefaultHandler[Input] {
	new := *h
	if emitterLogger == nil {
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
	h.emitterLogger.Emit(
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
	)
	// Handle and write output.
	h.handleOutput(w, r, nil, outError, statusCode)
}

// trackingWriter wraps http.ResponseWriter to track if headers have been written.
type trackingWriter struct {
	http.ResponseWriter
	wrote bool
}

func (tw *trackingWriter) WriteHeader(code int) {
	if !tw.wrote {
		tw.wrote = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

func (tw *trackingWriter) Write(p []byte) (int, error) {
	if !tw.wrote {
		tw.wrote = true
	}
	return tw.ResponseWriter.Write(p)
}

// handleOutput processes and writes the endpoint response.
func (h *DefaultHandler[Input]) handleOutput(
	w http.ResponseWriter, r *http.Request, out any, outError error, status int,
) {
	tw := &trackingWriter{ResponseWriter: w}
	if err := h.outputHandler.Handle(tw, r, out, outError, status); err != nil {
		h.emitterLogger.Emit(
			event.NewEvent(
				EventOutputError, fmt.Sprintf("Error handling output: %+v", err),
			).WithData(map[string]any{"err": err}),
		)
		if !tw.wrote {
			tw.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// defaultEmitterLogger returns a noop emitter logger.
func defaultEmitterLogger() event.EventEmitter {
	return event.NewNoopEventEmitter()
}
