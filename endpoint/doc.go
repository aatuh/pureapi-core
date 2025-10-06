// Package endpoint provides type-safe HTTP endpoint definitions and handlers.
//
// This package is the core of pureapi-core, offering generic endpoint
// handlers with compile-time type safety, middleware composition, and
// structured error handling. It provides a generic endpoint pipeline
// that transforms HTTP handlers from error-prone manual parsing to
// type-safe, composable functions.
//
// Example:
//
//	type UserRequest struct {
//		Name  string `json:"name"`
//		Email string `json:"email"`
//	}
//
//	type UserResponse struct {
//		ID    int    `json:"id"`
//		Name  string `json:"name"`
//		Email string `json:"email"`
//	}
//
//	// Input handler: parse JSON request body
//	inputHandler := inputJSON[UserRequest]()
//
//	// Business logic: create user
//	businessLogic := func(ctx context.Context, req UserRequest) (UserResponse, error) {
//		// Your business logic here
//		return UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
//	}
//
//	// Error handler: map errors to HTTP responses
//	errorHandler := func(err error) (int, string) {
//		return http.StatusInternalServerError, "Internal server error"
//	}
//
//	// Output handler: write JSON response
//	outputHandler := outputJSON[UserResponse]()
//
//	// Create the endpoint handler
//	handler := NewHandler(inputHandler, businessLogic, errorHandler, outputHandler)
//
//	// Use with your HTTP server
//	http.HandleFunc("/users", handler.ServeHTTP)
package endpoint
