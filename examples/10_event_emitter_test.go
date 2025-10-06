package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aatuh/pureapi-core"
	"github.com/aatuh/pureapi-core/event"
	"github.com/aatuh/pureapi-core/server"
)

type capturingEmitter struct {
	events []*event.Event
}

func (c *capturingEmitter) RegisterListener(event.EventType, event.EventCallback) event.EventEmitter {
	return c
}

func (c *capturingEmitter) RemoveListener(event.EventType, string) {}

func (c *capturingEmitter) Emit(ev *event.Event) {
	c.events = append(c.events, ev)
}

func (c *capturingEmitter) RegisterGlobalListener(event.EventCallback) event.EventEmitter {
	return c
}

func (c *capturingEmitter) RemoveGlobalListener(string) {}

// Show how to plug a custom event emitter and observe panic recovery.
func Test_CustomEventEmitter(t *testing.T) {
	emitter := &capturingEmitter{}
	srv := pureapi.NewServer(pureapi.WithEventEmitter(emitter))

	srv.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("bang")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for panic route, got %d", rr.Code)
	}

	var sawPanic bool
	for _, ev := range emitter.events {
		if ev.Type == server.EventPanic {
			sawPanic = true
			break
		}
	}

	if !sawPanic {
		t.Fatalf("expected to capture %q event, got %#v", server.EventPanic, emitter.events)
	}
}
