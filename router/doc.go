// Package router provides pluggable HTTP routing with path parameter support.
//
// This package defines a flexible routing interface that supports both exact
// matches and path parameters. It includes a built-in implementation with
// colon-style path parameters and can be extended with custom routing logic.
//
// Route Mutation: The builtin router is not thread-safe for concurrent route
// mutations. Register or unregister routes during startup, or guard runtime
// changes with your own synchronization.
package router
