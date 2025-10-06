package endpoint

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core/apierror"
	"github.com/aatuh/pureapi-core/event"
	"github.com/stretchr/testify/suite"
)

// dummyInputHandler implements InputHandler.
type dummyInputHandler struct {
	result *string
	err    error
}

func (d *dummyInputHandler) Handle(
	w http.ResponseWriter, r *http.Request,
) (*string, error) {
	return d.result, d.err
}

// dummyErrorHandler implements ErrorHandler.
type dummyErrorHandler struct {
	capturedErr error
	retStatus   int
	retAPIError apierror.APIError
}

func (d *dummyErrorHandler) Handle(err error) (int, apierror.APIError) {
	d.capturedErr = err
	return d.retStatus, d.retAPIError
}

// dummyOutputHandler implements OutputHandler.
type dummyOutputHandler struct {
	called     bool
	statusCode int
	out        any
	outErr     error
	retErr     error
}

func (d *dummyOutputHandler) Handle(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outputError error,
	statusCode int,
) error {
	d.called = true
	d.statusCode = statusCode
	d.out = out
	d.outErr = outputError
	w.WriteHeader(statusCode)
	if out != nil {
		fmt.Fprint(w, out)
	}
	return d.retErr
}

// dummyOutputHandlerNoWrite is used to simulate an output failure without
// writing a header.
type dummyOutputHandlerNoWrite struct {
	called bool
	retErr error
}

func (d *dummyOutputHandlerNoWrite) Handle(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outputError error,
	statusCode int,
) error {
	d.called = true
	// Do not write header or output; simply return the error.
	return d.retErr
}

// dummyEventEmitter implements types.EventEmitter.
type dummyEventEmitter struct {
	events []*event.Event
}

func (d *dummyEventEmitter) RegisterListener(eventType event.EventType, callback event.EventCallback) event.EventEmitter {
	return d
}

func (d *dummyEventEmitter) RemoveListener(eventType event.EventType, id string) {}

func (d *dummyEventEmitter) Emit(event *event.Event) {
	d.events = append(d.events, event)
}

func (d *dummyEventEmitter) RegisterGlobalListener(callback event.EventCallback) event.EventEmitter {
	return d
}

func (d *dummyEventEmitter) RemoveGlobalListener(id string) {}

// TableTestCase defines parameters for testing the Handle method.
type TableTestCase struct {
	name               string
	inHandlerErr       error
	logicErr           error
	outputRetErr       error
	errorHandlerStatus int
	expectedStatus     int
	expectedBody       string
	useNoWrite         bool // Use non-writing output handler?
}

// HandlerTestSuite tests the Handler.
type HandlerTestSuite struct {
	suite.Suite
	systemID *string
}

// TestHandlerTestSuite runs the test suite.
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// SetupTest sets up the test suite.
func (s *HandlerTestSuite) SetupTest() {
	id := "SYS123"
	s.systemID = &id
}

// Test_Handle tests various cases for the Handle method.
func (s *HandlerTestSuite) Test_Handle() {
	testCases := []TableTestCase{
		{
			name:               "InputError",
			inHandlerErr:       errors.New("input error"),
			logicErr:           nil,
			outputRetErr:       nil,
			errorHandlerStatus: 400,
			expectedStatus:     400,
		},
		{
			name:               "LogicError",
			inHandlerErr:       nil,
			logicErr:           errors.New("logic error"),
			outputRetErr:       nil,
			errorHandlerStatus: 422,
			expectedStatus:     422,
		},
		{
			name:               "OutputError",
			inHandlerErr:       nil,
			logicErr:           nil,
			outputRetErr:       errors.New("output error"),
			errorHandlerStatus: 0, // Not used in this case.
			expectedStatus:     http.StatusInternalServerError,
			useNoWrite:         true,
		},
		{
			name:               "Success",
			inHandlerErr:       nil,
			logicErr:           nil,
			outputRetErr:       nil,
			errorHandlerStatus: 0, // Not used in success case.
			expectedStatus:     http.StatusOK,
			expectedBody:       "logicOutput",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup input handler.
			var result *string
			if tc.inHandlerErr == nil {
				str := "valid"
				result = &str
			}
			inHandler := &dummyInputHandler{
				result: result,
				err:    tc.inHandlerErr,
			}

			// Setup logic function.
			logicFn := func(
				w http.ResponseWriter, r *http.Request, i *string,
			) (any, error) {
				return "logicOutput", tc.logicErr
			}

			// Setup output handler.
			var outHandler OutputHandler
			if tc.useNoWrite {
				outHandler = &dummyOutputHandlerNoWrite{retErr: tc.outputRetErr}
			} else {
				outHandler = &dummyOutputHandler{retErr: tc.outputRetErr}
			}

			// Setup error handler.
			errHandler := &dummyErrorHandler{
				retStatus:   tc.errorHandlerStatus,
				retAPIError: nil,
			}

			emitter := &dummyEventEmitter{}

			handler := NewHandler(
				inHandler, logicFn, errHandler, outHandler,
			).WithEmitterLogger(emitter)

			// Create request and invoke handler.
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/"+tc.name, nil)
			handler.Handle(rr, req)

			s.Equal(
				tc.expectedStatus, rr.Result().StatusCode,
				tc.name+" status code mismatch",
			)
			if tc.expectedBody != "" {
				s.Equal(
					tc.expectedBody, rr.Body.String(),
					tc.name+" response body mismatch",
				)
			}
		})
	}
}

// Test_Handle_NilEmitterLogger verifies that passing a nil emitter logger
// defaults correctly.
func (s *HandlerTestSuite) Test_Handle_NilEmitterLogger() {
	inputVal := "input"
	inHandler := &dummyInputHandler{
		result: &inputVal,
		err:    nil,
	}
	logicFn := func(
		w http.ResponseWriter, r *http.Request, i *string,
	) (any, error) {
		return "logic", nil
	}
	outHandler := &dummyOutputHandler{}
	errHandler := &dummyErrorHandler{}

	// Create handler without emitter. Should use a noop emitter.
	handler := NewHandler(inHandler, logicFn, errHandler, outHandler)

	// Create request and invoke handler.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nil-logger", nil)
	handler.Handle(rr, req)

	s.True(outHandler.called, "Output handler should be called")
	s.Equal("logic", rr.Body.String(), "Expected output 'logic'")
}
