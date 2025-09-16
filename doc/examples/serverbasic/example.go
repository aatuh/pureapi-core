package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aatuh/pureapi-core/endpoint"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/server"
)

// This example demonstrates how to use the server package to start a server.
func main() {
	// Create the endpoints.
	endpoints := []endpoint.Endpoint{
		endpoint.NewEndpoint("/hello", http.MethodGet).WithHandler(
			func(w http.ResponseWriter, r *http.Request) {
				log.Println("Incoming request")
				fmt.Fprintf(w, "Hello, PureAPI!")
			}),
	}

	// Create the server handler.
	port := 8080
	eventEmitter := SetupEventEmitter(port)
	emitterLogger := event.NewEmitterLogger(eventEmitter, nil)
	handler := server.NewHandler(emitterLogger)

	// Create a HTTP server.
	httpServer := server.DefaultHTTPServer(handler, port, endpoints)

	// Start the server.
	if err := server.StartServer(handler, httpServer, nil); err != nil {
		panic(fmt.Errorf("server panic: %w", err))
	}
}

// SetupEventEmitter sets up an event emitter for the server. It demonstrates
// how to register event listeners. For server there are more events available.
// See the server package for more information.
//
// Parameters:
//   - port: Port for the server.
//
// Returns:
//   - util.EventEmitter: The event emitter.
func SetupEventEmitter(port int) event.EventEmitter {
	eventEmitter := event.NewEventEmitter()
	eventEmitter.
		RegisterListener(
			server.EventStart,
			func(event *event.Event) {
				// Using event message directly for logging.
				log.Printf("Event: %s, port: %d\n", event.Message, port)
			},
		).
		RegisterListener(
			server.EventRegisterURL,
			func(event *event.Event) {
				// Using event data to log the path and methods.
				data := event.Data.(map[string]any)
				log.Printf(
					"Event: Registering URL %s %v\n",
					data["path"],
					data["methods"],
				)
			},
		)
	return eventEmitter
}
