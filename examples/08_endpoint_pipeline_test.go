package examples

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aatuh/pureapi-core"
)

// Show endpoint.DefaultHandler pipeline: input -> logic -> error/output.
func Test_EndpointPipeline_JSON(t *testing.T) {
	type In struct{ Name string }

	// JSON input handler
	ih := inputJSON[In]()

	// Business logic: error if empty name
	logic := pureapi.HandlerLogicFn[In](
		func(_ http.ResponseWriter, _ *http.Request, in *In) (any, error) {
			if strings.TrimSpace(in.Name) == "" {
				return nil, errors.New("empty name")
			}
			return map[string]any{"hello": in.Name}, nil
		},
	)

	// Map errors to APIError and status
	eh := errorMapper(func(err error) (int, pureapi.APIError) {
		return http.StatusBadRequest, pureapi.NewAPIError("invalid_input").WithMessage(err.Error())
	})

	// JSON output
	oh := outputJSON()

	h := pureapi.NewHandler(ih, logic, eh, oh)

	// Valid request
	req := httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(`{"Name":"Go"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Handle(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Invalid request
	req = httptest.NewRequest(
		http.MethodPost, "/", strings.NewReader(`{"Name":""}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	h.Handle(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Helpers for the pipeline example.

func inputJSON[T any]() pureapi.InputHandler[T] {
	return inputHandlerFunc[T](func(
		w http.ResponseWriter, r *http.Request,
	) (*T, error) {
		var v T
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			return nil, err
		}
		return &v, nil
	})
}

type inputHandlerFunc[T any] func(
	w http.ResponseWriter, r *http.Request,
) (*T, error)

func (f inputHandlerFunc[T]) Handle(
	w http.ResponseWriter, r *http.Request,
) (*T, error) {
	return f(w, r)
}

func errorMapper(
	mapper func(error) (int, pureapi.APIError),
) pureapi.ErrorHandler {
	return errorHandlerFunc(func(err error) (int, pureapi.APIError) {
		return mapper(err)
	})
}

type errorHandlerFunc func(err error) (int, pureapi.APIError)

func (f errorHandlerFunc) Handle(
	err error,
) (int, pureapi.APIError) {
	return f(err)
}

func outputJSON() pureapi.OutputHandler {
	return outputHandlerFunc(
		func(
			w http.ResponseWriter,
			_ *http.Request,
			out any,
			outErr error,
			status int,
		) error {
			if outErr != nil {
				w.WriteHeader(status)
				_ = json.NewEncoder(w).
					Encode(pureapi.APIErrorFrom(outErr.(pureapi.APIError)))
				return nil
			}
			w.WriteHeader(status)
			return json.NewEncoder(w).Encode(out)
		})
}

type outputHandlerFunc func(
	w http.ResponseWriter, r *http.Request, out any, outErr error, status int,
) error

func (f outputHandlerFunc) Handle(
	w http.ResponseWriter, r *http.Request, out any, outErr error, status int,
) error {
	return f(w, r, out, outErr, status)
}
