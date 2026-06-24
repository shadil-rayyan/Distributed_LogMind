package unit_test

import (
	"testing"
	"time"

	"logmind/internal/domain"
)

func TestServiceMetrics_RecordError(t *testing.T) {
	sm := domain.NewServiceMetrics(60)
	now := time.Now().Unix()

	sm.RecordError(now)
	sm.RecordError(now)

	count := sm.GetActiveCount(now)
	if count != 2 {
		t.Errorf("expected 2 active errors, got %d", count)
	}
}

func TestServiceMetrics_SlidingWindowExpiration(t *testing.T) {
	sm := domain.NewServiceMetrics(60)
	now := time.Now().Unix()

	sm.RecordError(now - 61) // Outside the window
	sm.RecordError(now)      // Inside the window

	count := sm.GetActiveCount(now)
	if count != 1 {
		t.Errorf("expected 1 active error, got %d", count)
	}
}
