package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/querydec"
	"github.com/aatuh/pureapi-core/router"
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
//   - *http.Server: A configured http.Server instance.
func DefaultHTTPServer(
	handler *Handler, port int, endpoints []endpoint.Endpoint,
) *http.Server {
	// Register endpoints with the handler
	handler.Register(endpoints)

	return &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second, // Limits slow clients.
		WriteTimeout:   10 * time.Second, // Ensures fast responses.
		IdleTimeout:    60 * time.Second, // Keeps alive long enough.
		MaxHeaderBytes: 1 << 16,          // 64KB to prevent excessive memory use.
		// Security hardening
		ReadHeaderTimeout: 5 * time.Second, // Prevent slow header attacks
		// Connection handling
		ConnState: func(conn net.Conn, state http.ConnState) {
			// Close connections on parser errors or oversized headers
			if state == http.StateActive {
				// This is where we could add connection monitoring
			}
		},
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
//   - error: An error if starting the server fails.
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
	emitter      event.EventEmitter
	router       router.Router
	queryDecoder querydec.Decoder
	notFound     http.Handler
	recoverer    func(http.Handler) http.Handler
	bodyLimit    int64 // Maximum request body size in bytes
	// Store registered routes for method not allowed checking
	registeredRoutes map[string]map[string]bool // path -> method -> exists
	routesMu         sync.RWMutex
}

// HandlerOption configures a Handler.
type HandlerOption func(*Handler)

// WithRouter sets the router for the handler.
//
// Parameters:
//   - r: The router implementation to use.
//
// Returns:
//   - HandlerOption: A handler option function.
func WithRouter(r router.Router) HandlerOption {
	return func(h *Handler) { h.router = r }
}

// WithQueryDecoder sets the query decoder for the handler.
//
// Parameters:
//   - d: The query decoder to use.
//
// Returns:
//   - HandlerOption: A handler option function.
func WithQueryDecoder(d querydec.Decoder) HandlerOption {
	return func(h *Handler) { h.queryDecoder = d }
}

// WithEventEmitter overrides the handler event emitter.
func WithEventEmitter(em event.EventEmitter) HandlerOption {
	return func(h *Handler) {
		if em != nil {
			h.emitter = em
		}
	}
}

// WithNotFound sets the not found handler.
//
// Parameters:
//   - nf: The not found handler to use.
//
// Returns:
//   - HandlerOption: A handler option function.
func WithNotFound(nf http.Handler) HandlerOption {
	return func(h *Handler) { h.notFound = nf }
}

// WithRecoverer sets the recoverer function.
//
// Parameters:
//   - wrap: The recoverer function to use.
//
// Returns:
//   - HandlerOption: A handler option function.
func WithRecoverer(wrap func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) { h.recoverer = wrap }
}

// WithBodyLimit sets the maximum request body size in bytes.
//
// Parameters:
//   - limit: The maximum request body size in bytes.
//
// Returns:
//   - HandlerOption: A handler option function.
func WithBodyLimit(limit int64) HandlerOption {
	return func(h *Handler) { h.bodyLimit = limit }
}

// NewHandler creates a new HTTPServer.
// If an event emitter is provided, it will be used to emit events. Otherwise,
// logging will be used. If no logger is provided, log.Default() will be used.
// If no event emitter is provided, no events will be emitted or logged.
//
// Parameters:
//   - emitter: Event emitter logger.
//   - opts: Optional handler options.
//
// Returns:
//   - *Handler: A new Handler instance.
func NewHandler(
	emitter event.EventEmitter,
	opts ...HandlerOption,
) *Handler {
	h := &Handler{
		emitter:          emitter,
		notFound:         http.NotFoundHandler(),
		queryDecoder:     querydec.PlainDecoder{},
		bodyLimit:        2 * 1024 * 1024, // 2MB default
		registeredRoutes: make(map[string]map[string]bool),
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.router == nil {
		// Provide a tiny built-in router for zero deps.
		h.router = router.NewBuiltinRouter()
	}
	if h.recoverer == nil {
		h.recoverer = h.createRecoverer()
	}
	return h
}

// startServer starts the HTTP server and listens for shutdown signals.
func (s *Handler) startServer(
	stopChan chan os.Signal,
	server HTTPServer,
	shutdownTimeout time.Duration,
) error {
	// Prepare channel for shutdown signal.
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopChan)
	errChan := make(chan error, 1)

	go func() {
		s.listenAndServe(server, errChan, stopChan)
	}()

	// Wait for shutdown signal.
	<-stopChan

	// Give the server some time to shut down.
	s.emitter.Emit(
		event.NewEvent(EventShutDownStarted, "Shutting down HTTP server"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.emitter.Emit(
			event.NewEvent(
				EventShutDownError,
				"HTTP server shutdown error",
			).WithData(map[string]any{"error": err}),
		)
		return fmt.Errorf("startServer: shutdown error: %w", err)
	}
	s.emitter.Emit(
		event.NewEvent(EventShutDown, "HTTP server shut down"),
	)
	return <-errChan
}

// listenAndServe listens and serves the HTTP server.
func (s *Handler) listenAndServe(
	server HTTPServer, errChan chan error, stopChan chan os.Signal,
) {
	s.emitter.Emit(
		event.NewEvent(EventStart, "Starting HTTP server"),
	)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		s.emitter.Emit(
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

// Register registers endpoints with the handler.
//
// Parameters:
//   - endpoints: The endpoints to register.
//
// Returns:
//   - error: An error if the endpoint registration fails.
func (h *Handler) Register(endpoints []endpoint.Endpoint) {
	for _, ep := range endpoints {
		// Compose middleware stack.
		middlewares := ep.Middlewares()
		var handler http.Handler
		if ep.Handler() != nil {
			handler = http.HandlerFunc(ep.Handler())
		} else {
			handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "Not Implemented", http.StatusNotImplemented)
			})
		}

		if middlewares != nil {
			handler = middlewares.Chain(handler)
		}

		// Register to router with method+pattern.
		h.router.Register(ep.Method(), ep.URL(), handler)

		// Track registered routes for method not allowed checking
		h.routesMu.Lock()
		if h.registeredRoutes[ep.URL()] == nil {
			h.registeredRoutes[ep.URL()] = make(map[string]bool)
		}
		h.registeredRoutes[ep.URL()][ep.Method()] = true
		h.routesMu.Unlock()

		h.emitter.Emit(
			event.NewEvent(
				EventRegisterURL,
				fmt.Sprintf("Registering URL: %s %s", ep.URL(), ep.Method()),
			).WithData(map[string]any{"path": ep.URL(), "method": ep.Method()}),
		)
	}
}

// Unregister removes a method+path route from the router and tracking map.
//
// Parameters:
//   - method: The HTTP method of the route.
//   - path: The path of the route.
//
// Returns:
//   - error: An error if the endpoint unregistration fails.
func (h *Handler) Unregister(method, path string) {
	if h.router != nil {
		_ = h.router.Unregister(method, path)
	}
	h.routesMu.Lock()
	if mm, ok := h.registeredRoutes[path]; ok {
		delete(mm, method)
		if len(mm) == 0 {
			delete(h.registeredRoutes, path)
		}
	}
	h.routesMu.Unlock()
}

// ServeHTTP implements http.Handler.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//
// Returns:
//   - error: An error if the request serving fails.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Wrap with tracking response writer to prevent double WriteHeader
	tw := newTrackingResponseWriter(w)
	// Body limits as you have them...
	if h.bodyLimit > 0 && r.ContentLength > h.bodyLimit {
		http.Error(tw, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	if h.bodyLimit > 0 {
		r.Body = http.MaxBytesReader(tw, r.Body, h.bodyLimit)
	}

	// Auto OPTIONS: check for explicit handler first, then synthesize
	if r.Method == http.MethodOptions {
		// Check if there's an explicit OPTIONS handler
		m := h.router.Match(r)
		if m != nil {
			// Use explicit OPTIONS handler
			qm, _ := h.queryDecoder.Decode(r.URL.Query())
			ctx := context.WithValue(r.Context(), ctxKeyQueryMapVal, qm)
			if len(m.Params) > 0 {
				ctx = context.WithValue(ctx, ctxKeyRouteParamsVal, m.Params)
			}
			r = r.WithContext(ctx)
			h.recoverer(m.Handler).ServeHTTP(tw, r)
			return
		}
		// No explicit OPTIONS handler, synthesize response
		if allow := h.allowedMethods(r.URL.Path); len(allow) > 0 {
			tw.Header().Set("Allow", strings.Join(allow, ", "))
			tw.WriteHeader(http.StatusNoContent)
			return
		}
	}

	m := h.router.Match(r)

	// HEAD fallback: if GET exists but no direct HEAD handler.
	if m == nil && r.Method == http.MethodHead {
		r2 := r.Clone(r.Context())
		r2.Method = http.MethodGet
		if m2 := h.router.Match(r2); m2 != nil {
			// Decode query + params same as below
			qm, _ := h.queryDecoder.Decode(r2.URL.Query())
			ctx := context.WithValue(r2.Context(), ctxKeyQueryMapVal, qm)
			if len(m2.Params) > 0 {
				ctx = context.WithValue(ctx, ctxKeyRouteParamsVal, m2.Params)
			}
			r2 = r2.WithContext(ctx)

			h.recoverer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				// Discard body writes.
				dw := &discardingWriter{ResponseWriter: w}
				m2.Handler.ServeHTTP(dw, r2) // use r2 (GET)
			})).ServeHTTP(tw, r2)
			return
		}
	}

	if m == nil {
		if h.isMethodNotAllowed(r) {
			if allow := h.allowedMethods(r.URL.Path); len(allow) > 0 {
				tw.Header().Set("Allow", strings.Join(allow, ", "))
			}
			http.Error(tw, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}
		h.notFound.ServeHTTP(tw, r)
		return
	}

	// Decode query + params into context.
	qm, _ := h.queryDecoder.Decode(r.URL.Query())
	ctx := context.WithValue(r.Context(), ctxKeyQueryMapVal, qm)
	if len(m.Params) > 0 {
		ctx = context.WithValue(ctx, ctxKeyRouteParamsVal, m.Params)
	}
	r = r.WithContext(ctx)

	h.recoverer(m.Handler).ServeHTTP(tw, r)
}

func (h *Handler) allowedMethods(path string) []string {
	// Prefer router introspection if available.
	type methodsFor interface{ MethodsFor(string) []string }
	if mf, ok := h.router.(methodsFor); ok {
		return mf.MethodsFor(path)
	}

	// Fallback to registeredRoutes map + stableAllow
	h.routesMu.RLock()
	defer h.routesMu.RUnlock()

	set := map[string]struct{}{}
	// Exact path methods
	if mm, ok := h.registeredRoutes[path]; ok {
		for m := range mm {
			set[m] = struct{}{}
		}
	}
	// Colon/braces pattern methods
	for pat, mm := range h.registeredRoutes {
		if pat == path || h.matchesPattern(pat, path) {
			for m := range mm {
				set[m] = struct{}{}
			}
		}
	}
	return stableAllow(set)
}

// isMethodNotAllowed checks if the request path exists but with a different
// method.
func (h *Handler) isMethodNotAllowed(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	h.routesMu.RLock()
	defer h.routesMu.RUnlock()

	// Check if this path exists with any method
	if methods, exists := h.registeredRoutes[path]; exists {
		// Check if the current method is not in the allowed methods
		if !methods[method] {
			return true
		}
	}

	// Also check for colon parameter patterns
	for registeredPath := range h.registeredRoutes {
		if h.matchesPattern(registeredPath, path) {
			methods := h.registeredRoutes[registeredPath]
			if !methods[method] {
				return true
			}
		}
	}

	return false
}

type discardingWriter struct{ http.ResponseWriter }

// Write discards the written data.
//
// Parameters:
//   - p: The data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if the write fails.
func (d *discardingWriter) Write(p []byte) (int, error) { return len(p), nil }

// matchesPattern checks if a pattern matches a path (for colon and brace parameters).
func (h *Handler) matchesPattern(pattern, path string) bool {
	// Simple colon and brace parameter matching
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		// treat ":id" and "{id}" as params
		if strings.HasPrefix(patternPart, ":") ||
			(strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}")) {
			// This is a parameter, so it matches any value
			if pathPart == "" { // require a real segment for params
				return false
			}
			continue
		} else if patternPart != pathPart {
			// Exact match required for non-parameter parts
			return false
		}
	}

	return true
}

// Access helpers for handlers.
type ctxKeyQueryMap struct{}
type ctxKeyRouteParams struct{}

var (
	ctxKeyQueryMapVal    = ctxKeyQueryMap{}
	ctxKeyRouteParamsVal = ctxKeyRouteParams{}
)

// QueryMap extracts the query map from the request context.
func QueryMap(r *http.Request) map[string]any {
	if v := r.Context().Value(ctxKeyQueryMapVal); v != nil {
		return v.(map[string]any)
	}
	return nil
}

// RouteParams extracts the route parameters from the request context.
func RouteParams(r *http.Request) map[string]string {
	if v := r.Context().Value(ctxKeyRouteParamsVal); v != nil {
		// Handle both router.Params and map[string]string
		if params, ok := v.(router.Params); ok {
			return map[string]string(params)
		}
		if params, ok := v.(map[string]string); ok {
			return params
		}
	}
	return nil
}

// serverPanicHandler returns an HTTP handler that recovers from panics.
//
// Parameters:
//   - next: The next handler in the chain.
//
// Returns:
//   - http.Handler: A handler that recovers from panics.
func (s *Handler) serverPanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				panicRecovery(w, err, s.emitter)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// createRecoverer creates a panic recovery middleware with emitter access.
//
// Returns:
//   - func(http.Handler) http.Handler: A panic recovery middleware function.
func (h *Handler) createRecoverer() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					panicRecovery(w, err, h.emitter)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// panicRecovery handles recovery from panics.
//
// Parameters:
//   - w: The HTTP response writer.
//   - err: The panic error.
//   - emitter: The event emitter for logging.
func panicRecovery(w http.ResponseWriter, err any, emitter event.EventEmitter) {
	emitter.Emit(
		event.NewEvent(
			EventPanic,
			fmt.Sprintf("Panic recovered: %v", err),
		).WithData(map[string]any{"panic": err}),
	)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// stableAllow returns a deterministic, RFC-friendly Allow list.
func stableAllow(methods map[string]struct{}) []string {
	// Always advertise HEAD with GET, and OPTIONS when anything exists.
	h := map[string]struct{}{}
	for m := range methods {
		h[m] = struct{}{}
	}
	if _, ok := h["GET"]; ok {
		h["HEAD"] = struct{}{}
	}
	if len(h) > 0 {
		h["OPTIONS"] = struct{}{}
	}

	order := []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	out := make([]string, 0, len(h))
	// stable order first
	for _, m := range order {
		if _, ok := h[m]; ok {
			out = append(out, m)
			delete(h, m)
		}
	}
	// include any custom methods in sorted order for determinism
	rest := make([]string, 0, len(h))
	for m := range h {
		rest = append(rest, m)
	}
	slices.Sort(rest)
	out = append(out, rest...)
	return out
}
