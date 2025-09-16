package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
)

// Define events.
const (
	EventRegisterURL      event.EventType = "event_register_url"
	EventNotFound         event.EventType = "event_not_found"
	EventMethodNotAllowed event.EventType = "event_method_not_allowed"
	EventPanic            event.EventType = "event_panic"
	EventStart            event.EventType = "event_start"
	EventErrorStart       event.EventType = "event_error_start"
	EventShutDownStarted  event.EventType = "event_shutdown_started"
	EventShutDown         event.EventType = "event_shutdown"
	EventShutDownError    event.EventType = "event_shutdown_error"
)

// HTTPServer represents an HTTP server.
type HTTPServer interface {
	ListenAndServe() error              // Start the server.
	Shutdown(ctx context.Context) error // Shut down the server.
}

// DefaultHTTPServer returns the default HTTP server implementation. It sets
// default request read and write timeouts of 10 seconds, idle timeout of 60
// seconds, and a max header size of 64KB.
//
// Parameters:
//   - handler: HTTP server handler.
//   - port: Port for the HTTP server.
//   - endpoints: Endpoints to register.
//
// Returns:
//   - *http.Server: http.Server instance.
func DefaultHTTPServer(
	handler *Handler, port int, endpoints []endpoint.Endpoint,
) *http.Server {
	return &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler.setupMux(endpoints),
		ReadTimeout:    10 * time.Second, // Limits slow clients.
		WriteTimeout:   10 * time.Second, // Ensures fast responses.
		IdleTimeout:    60 * time.Second, // Keeps alive long enough.
		MaxHeaderBytes: 1 << 16,          // 64KB to prevent excessive memory use.
	}
}

// StartServer sets up an HTTP server with the specified port and endpoints,
// using optional event emitter. The handler listens for OS interrupt signals to
// gracefully shut down. If no shutdown timeout is provided, 60 seconds will be
// used by default.
//
// Parameters:
//   - handler: HTTP server handler.
//   - server: Server implementation to use.
//   - shutdownTimeout: Optional shutdown timeout.
//
// Returns:
//   - error: Error starting the server.
func StartServer(
	handler *Handler,
	server HTTPServer,
	shutdownTimeout *time.Duration,
) error {
	var useShutdownTimeout time.Duration
	if shutdownTimeout == nil {
		useShutdownTimeout = 60 * time.Second
	} else {
		useShutdownTimeout = *shutdownTimeout
	}
	return handler.startServer(
		make(chan os.Signal, 1), server, useShutdownTimeout,
	)
}

// Handler represents an HTTP server handler.
type Handler struct {
	emitterLogger event.EmitterLogger
}

// NewHandler creates a new HTTPServer.
// If an event emitter is provided, it will be used to emit events. Otherwise,
// logging will be used. If no logger is provided, log.Default() will be used.
// If no event emitter is provided, no events will be emitted or logged.
//
// Parameters:
//   - eventEmitter: Optional event emitter.
//
// Returns:
//   - *Handler: A new Handler instance.
func NewHandler(emitterLogger event.EmitterLogger) *Handler {
	var useEmitterLogger event.EmitterLogger
	if emitterLogger == nil {
		useEmitterLogger = event.NewNoopEmitterLogger()
	} else {
		useEmitterLogger = emitterLogger
	}
	return &Handler{
		emitterLogger: useEmitterLogger,
	}
}

// startServer starts the HTTP server and listens for shutdown signals.
func (s *Handler) startServer(
	stopChan chan os.Signal,
	server HTTPServer,
	shutdownTimeout time.Duration,
) error {
	// Prepare channel for shutdown signal.
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	errChan := make(chan error, 1)

	go func() {
		s.listenAndServe(server, errChan, stopChan)
	}()

	// Wait for shutdown signal.
	<-stopChan

	// Give the server some time to shut down.
	s.emitterLogger.Info(
		event.NewEvent(EventShutDownStarted, "Shutting down HTTP server"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.emitterLogger.Error(
			event.NewEvent(
				EventShutDownError,
				"HTTP server shutdown error",
			).WithData(map[string]any{"error": err}),
		)
		return fmt.Errorf("startServer: shutdown error: %w", err)
	}
	s.emitterLogger.Info(
		event.NewEvent(EventShutDown, "HTTP server shut down"),
	)
	return <-errChan
}

// listenAndServe listens and serves the HTTP server.
func (s *Handler) listenAndServe(
	server HTTPServer, errChan chan error, stopChan chan os.Signal,
) {
	s.emitterLogger.Info(
		event.NewEvent(EventStart, "Starting HTTP server"),
	)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.emitterLogger.Error(
			event.NewEvent(
				EventErrorStart,
				fmt.Sprintf("Error starting HTTP server: %v", err),
			).WithData(map[string]any{"error": err}),
		)
		errChan <- err
		stopChan <- os.Interrupt
	} else {
		errChan <- nil
	}
}

// setupMux sets up the HTTP mux with the specified endpoints.
func (s *Handler) setupMux(
	httpEndpoints []endpoint.Endpoint,
) *http.ServeMux {
	mux := http.NewServeMux()
	endpoints := s.multiplexEndpoints(httpEndpoints)

	for url := range endpoints {
		methods := mapKeys(endpoints[url])
		s.emitterLogger.Info(
			event.NewEvent(
				EventRegisterURL,
				fmt.Sprintf("Registering URL: %s %v", url, methods),
			).WithData(map[string]any{"path": url, "methods": methods}),
		)
		iterURL := url
		mux.Handle(iterURL, s.createEndpointHandler(endpoints[iterURL]))
	}

	// Only register the not found handler if "/" is not already an endpoint.
	if _, exists := endpoints["/"]; !exists {
		mux.Handle("/", s.createNotFoundHandler())
	}

	return mux
}

// createEndpointHandler creates an HTTP handler for the specified endpoints.
func (s *Handler) createEndpointHandler(
	endpoints map[string]http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := endpoints[r.Method]; ok {
			handler.ServeHTTP(w, r)
			return
		}
		s.emitterLogger.Info(
			event.NewEvent(
				EventMethodNotAllowed,
				fmt.Sprintf(
					"Method not allowed: %s (%v)", r.URL.Path, r.Method,
				),
			).WithData(map[string]any{"path": r.URL.Path, "method": r.Method}),
		)
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
	}
}

// createNotFoundHandler creates an HTTP handler for not found requests.
func (s *Handler) createNotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.emitterLogger.Info(
			event.NewEvent(
				EventNotFound,
				fmt.Sprintf("Not found: %s (%v)", r.URL.Path, r.Method),
			).WithData(map[string]any{"path": r.URL.Path, "method": r.Method}),
		)
		http.Error(
			w,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	}
}

// multiplexEndpoints multiplexes endpoints by URL and method.
func (s *Handler) multiplexEndpoints(
	endpoints []endpoint.Endpoint,
) map[string]map[string]http.Handler {
	multiplexed := make(map[string]map[string]http.Handler)
	for _, endpoint := range endpoints {
		s.multiplexEndpoint(endpoint, multiplexed)
	}
	return multiplexed
}

// multiplexEndpoint multiplexes an endpoint by URL and method.
func (s *Handler) multiplexEndpoint(
	endpoint endpoint.Endpoint,
	multiplexed map[string]map[string]http.Handler,
) {
	if multiplexed[endpoint.URL()] == nil {
		multiplexed[endpoint.URL()] = make(map[string]http.Handler)
	}
	middlewares := endpoint.Middlewares()
	multiplexed[endpoint.URL()][endpoint.Method()] = s.serverPanicHandler(
		middlewares.Chain(emptyOrCustomHandler(endpoint)),
	)
}

// emptyOrCustomHandler determines the HTTP handler for the endpoint.
func emptyOrCustomHandler(endpoint endpoint.Endpoint) http.Handler {
	if endpoint.Handler() != nil {
		// Use the provided handler.
		return http.HandlerFunc(endpoint.Handler())
	}
	// Fallback to a default no-op handler.
	return http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) {},
	)
}

// serverPanicHandler returns an HTTP handler that recovers from panics.
func (s *Handler) serverPanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.panicRecovery(w, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// panicRecovery handles recovery from panics.
func (s *Handler) panicRecovery(w http.ResponseWriter, err any) {
	s.emitterLogger.Error(
		event.NewEvent(
			EventPanic,
			fmt.Sprintf("Server panic: %v", err),
		).WithData(map[string]any{"stack": stackTraceSlice()}),
	)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// stackTraceSlice returns the stack trace as a slice of strings.
func stackTraceSlice() []string {
	var trace []string
	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			return trace
		}
		fn := runtime.FuncForPC(pc)
		trace = append(trace, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
}

// mapKeys returns the keys of a map.
func mapKeys(m map[string]http.Handler) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
