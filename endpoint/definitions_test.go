package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// DefinitionsTestSuite tests the collection of endpoint definitions.
type DefinitionsTestSuite struct {
	suite.Suite
}

// TestDefinitionsTestSuite runs the test suite.
func TestDefinitionsTestSuite(t *testing.T) {
	suite.Run(t, new(DefinitionsTestSuite))
}

// Test_ToEndpoints tests the ToEndpoints helper function.
func (s *DefinitionsTestSuite) Test_ToEndpoints() {
	// Create two endpoint specifications with simple handlers.
	spec1 := NewEndpointSpec("/one", "GET", nil,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("one"))
			if err != nil {
				panic(err)
			}
		}))
	spec2 := NewEndpointSpec("/two", "POST", nil,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("two"))
			if err != nil {
				panic(err)
			}
		}))

	// Test with single specification
	endpoints1 := ToEndpoints(spec1)
	s.Equal(1, len(endpoints1))

	// Test with multiple specifications
	endpoints2 := ToEndpoints(spec1, spec2)
	s.Equal(2, len(endpoints2))
}

// Test_ToEndpointsDetailed tests that ToEndpoints returns a slice of endpoints with
// the correct URL, method and handler.
func (s *DefinitionsTestSuite) Test_ToEndpointsDetailed() {
	// Create two endpoint specifications.
	spec1 := NewEndpointSpec("/one", "GET", nil,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("one"))
			if err != nil {
				panic(err)
			}
		}))
	spec2 := NewEndpointSpec("/two", "POST", nil,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("two"))
			if err != nil {
				panic(err)
			}
		}))

	endpoints := ToEndpoints(spec1, spec2)
	s.Equal(2, len(endpoints))

	// For each endpoint, verify URL, method and handler output.
	for i, ep := range endpoints {
		var expectedURL, expectedMethod, expectedBody string
		if i == 0 {
			expectedURL = "/one"
			expectedMethod = "GET"
			expectedBody = "one"
		} else {
			expectedURL = "/two"
			expectedMethod = "POST"
			expectedBody = "two"
		}
		// Assume that the endpoint implements the following methods.
		epAccessor, ok := ep.(interface {
			URL() string
			Method() string
			Handler() http.HandlerFunc
		})
		s.True(ok, "endpoint should implement URL, Method, and Handler")
		s.Equal(expectedURL, epAccessor.URL())
		s.Equal(expectedMethod, epAccessor.Method())

		rr := httptest.NewRecorder()
		epAccessor.Handler()(rr, httptest.NewRequest(expectedMethod,
			expectedURL, nil))
		s.Equal(expectedBody, rr.Body.String())
	}
}
