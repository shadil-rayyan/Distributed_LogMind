package domain

import (
	"sync"
)

// ServiceMetrics implements an allocation-free rolling 60-second window
type ServiceMetrics struct {
	mu        sync.RWMutex
	windowSec int
	slots     []int   // Array index corresponds to (timestamp % windowSec)
	epochs    []int64 // Tracks which exact second a slot belongs to
}

func NewServiceMetrics(windowSec int) *ServiceMetrics {
	return &ServiceMetrics{
		windowSec: windowSec,
		slots:     make([]int, windowSec),
		epochs:    make([]int64, windowSec),
	}
}

func (sm *ServiceMetrics) RecordError(timestamp int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	slot := timestamp % int64(sm.windowSec)
	if sm.epochs[slot] != timestamp {
		// Slot belongs to an older minute window, overwrite it
		sm.epochs[slot] = timestamp
		sm.slots[slot] = 1
	} else {
		sm.slots[slot]++
	}
}

func (sm *ServiceMetrics) GetActiveCount(now int64) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total := 0
	for i := 0; i < sm.windowSec; i++ {
		// Only sum counts that fall within the current sliding window boundary
		if now-sm.epochs[i] < int64(sm.windowSec) {
			total += sm.slots[i]
		}
	}
	return total
}
