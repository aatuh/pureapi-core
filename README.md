# pureapi-core

Robust, type-safe HTTP APIs in Go with minimal boilerplate.

pureapi-core provides the essential building blocks for creating production-ready
HTTP services. Whether you're building microservices, REST APIs, or web
applications, this library gives you the power of modern Go patterns with a
clean, intuitive API.

## Key Features

- Type-Safe Endpoints: Generic endpoint pipeline with compile-time type safety
- Pluggable Architecture: Swap routers, query decoders, and middleware as needed  
- Zero Dependencies: Built on Go standard library only
- Production Ready: Built-in body limits, panic recovery, and automatic HEAD/OPTIONS handling
- Composable: Mix and match components to fit your architecture
- Event-Driven: Built-in event emitter for observability and monitoring
- Clean API: Express-like syntax with Go's type system benefits

**Generic Endpoint Pipeline**: Transform your HTTP handlers from error-prone manual
parsing to type-safe, composable functions:

```go
// Input → Logic → Error Handling → Output
handler := endpoint.NewHandler(
    inputJSON[UserRequest](),     // Parse JSON input
    businessLogic,                // Your typed business logic  
    errorMapper,                  // Map errors to HTTP responses
    outputJSON(),                 // Serialize JSON output
)
```

**Smart Routing**: Built-in router with path parameters, method validation, and
automatic 405 responses.

**Middleware Stack**: Compose cross-cutting concerns like authentication,
logging, and rate limiting.

**Event System**: Built-in event emitter for metrics, logging, and inter-service
communication. Wire your own emitter with `pureapi.WithEventEmitter` to stream
events into your observability stack.

**Swappability**: Pluggable architecture lets you swap components:

```go
import (
    "github.com/aatuh/pureapi-core"
    brackets "github.com/aatuh/pureapi-querydec-brackets"
)

// Custom router and query decoder
server := pureapi.NewServer(
    pureapi.WithRouter(pureapi.NewBuiltinRouter()),
    pureapi.WithQueryDecoder(brackets.NewDecoder()),
)

server.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    // Access route params and query params
    params := pureapi.RouteParams(r)
    query := pureapi.QueryMap(r)
    
    // Handle request...
})
```

## Install

```bash
go get github.com/aatuh/pureapi-core
```

## Quick start

Check the examples from the [examples package](./examples).

Run examples with:

```bash
go test -v -count 1 ./examples

# Run specific example.
go test -v -count 1 ./examples -run Test_BasicHTTP
```
