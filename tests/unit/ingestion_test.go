package unit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"logmind/internal/domain"
	"logmind/internal/ingestion"
)

func TestHandleLogsWaitsForQueueCapacity(t *testing.T) {
	logCh := make(chan domain.Log)
	handler := ingestion.NewLogHandler(logCh)

	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(`{"service":"payment","level":"error","message":"db down"}`))
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.HandleLogs(rr, req)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("handler returned before queue capacity was available")
	case <-time.After(50 * time.Millisecond):
	}

	received := make(chan domain.Log, 1)
	go func() {
		received <- <-logCh
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not complete after queue drained")
	}

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d", rr.Code)
	}

	select {
	case logged := <-received:
		if logged.Service != "payment" || logged.Level != "error" {
			t.Fatalf("unexpected log payload: %+v", logged)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive queued log")
	}
}
