package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// Limit of 2 requests/sec, burst of 2
	limiter := NewLimiter(2.0, 2)
	handler := limiter.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request 1: Allowed (uses 1 burst token)
	req1 := httptest.NewRequest("POST", "/logs", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr1.Code)
	}

	// Request 2: Allowed (uses 2nd burst token)
	req2 := httptest.NewRequest("POST", "/logs", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr2.Code)
	}

	// Request 3: Blocked (tokens exhausted)
	req3 := httptest.NewRequest("POST", "/logs", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests, got %d", rr3.Code)
	}

	// Request 4 from a different IP: Allowed
	req4 := httptest.NewRequest("POST", "/logs", nil)
	req4.RemoteAddr = "192.168.1.2:12345"
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for different IP, got %d", rr4.Code)
	}
}
