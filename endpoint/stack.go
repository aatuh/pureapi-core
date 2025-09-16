package endpoint

import (
	"sync"
)

// DefaultStack manages a list of middleware wrappers with concurrency safety
// for editing the list.
type DefaultStack struct {
	mu       sync.RWMutex
	wrappers []Wrapper
}

// DefaultStack implements the Stack interface.
var _ Stack = (*DefaultStack)(nil)

// NewStack creates and returns an initialized DefaultStack.
//
// Parameters:
//   - wrappers: The initial list of middleware wrappers.
//
// Returns:
//   - *DefaultStack: A new DefaultStack instance.
func NewStack(wrappers ...Wrapper) *DefaultStack {
	return &DefaultStack{
		mu:       sync.RWMutex{},
		wrappers: wrappers,
	}
}

// Wrappers returns the list of middleware wrappers in the stack.
//
// Returns:
//   - []Wrapper: The list of middleware wrappers in the stack.
func (s *DefaultStack) Wrappers() []Wrapper {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wrappers
}

// Middlewares returns the middlewares in the stack.
//
// Returns:
//   - Middlewares: The list of middlewares in the stack.
func (s *DefaultStack) Middlewares() Middlewares {
	s.mu.RLock()
	defer s.mu.RUnlock()
	middlewares := []Middleware{}
	for _, wrapper := range s.wrappers {
		middlewares = append(middlewares, wrapper.Middleware())
	}
	return NewMiddlewares(middlewares...)
}

// Clone creates a deep copy of the Stack.
//
// Returns:
//   - *Stack: The cloned middleware stack.
func (s *DefaultStack) Clone() Stack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	newStack := &DefaultStack{}
	newStack.wrappers = make([]Wrapper, len(s.wrappers))
	copy(newStack.wrappers, s.wrappers)
	return newStack
}

// Add appends a new middleware Wrapper to the stack and returns the stack for
// chaining.
//
// Parameters:
//   - w: The wrapper to add.
//
// Returns:
//   - *Stack: The updated middleware stack.
func (s *DefaultStack) AddWrapper(w Wrapper) Stack {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wrappers = append(s.wrappers, w)
	return s
}

// InsertBefore inserts a middleware Wrapper before the one with the specified
// ID. Returns true if a matching wrapper was found and insertion happened
// before it; if no match is found, the new wrapper is appended and false is
// returned.
//
// Parameters:
//   - id: The ID of the wrapper to insert before.
//   - w: The wrapper to insert.
//
// Returns:
//   - *Stack: The updated middleware stack.
//   - bool: True if a matching wrapper was found and insertion succeeded.
func (s *DefaultStack) InsertBefore(
	id string, w Wrapper,
) (Stack, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, wrapper := range s.wrappers {
		if wrapper.ID() == id {
			s.wrappers = append(
				s.wrappers[:i],
				append([]Wrapper{w}, s.wrappers[i:]...)...,
			)
			return s, true
		}
	}
	s.wrappers = append(s.wrappers, w)
	return s, false
}

// InsertAfter inserts a middleware Wrapper after the one with the specified ID.
// Returns true if a matching wrapper was found and insertion happened after it.
// If no match is found, the new wrapper is appended and false is returned.
//
// Parameters:
//   - id: The ID of the wrapper to insert after.
//   - w: The wrapper to insert.
//
// Returns:
//   - *Stack: The updated middleware stack.
//   - bool: True if a matching wrapper was found and insertion succeeded.
func (s *DefaultStack) InsertAfter(
	id string, w Wrapper,
) (Stack, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, wrapper := range s.wrappers {
		if wrapper.ID() == id {
			pos := i + 1
			s.wrappers = append(
				s.wrappers[:pos],
				append([]Wrapper{w}, s.wrappers[pos:]...)...,
			)
			return s, true
		}
	}
	s.wrappers = append(s.wrappers, w)
	return s, false
}

// Remove deletes the middleware Wrapper with the specified ID from the stack.
// Returns true if the middleware was found and removed; false otherwise.
//
// Parameters:
//   - id: The ID of the wrapper to remove.
//
// Returns:
//   - *Stack: The updated middleware stack.
//   - bool: True if the middleware was found and removed; false otherwise.
func (s *DefaultStack) Remove(id string) (Stack, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, wrapper := range s.wrappers {
		if wrapper.ID() == id {
			s.wrappers = append(s.wrappers[:i], s.wrappers[i+1:]...)
			return s, true
		}
	}
	return s, false
}
