package detection

import (
	"sync"
	"logmind/internal/domain"
)

type Engine struct {
	mu        sync.RWMutex
	metricMap map[string]*domain.ServiceMetrics
	windowSec int
}

func NewEngine(windowSec int) *Engine {
	return &Engine{
		metricMap: make(map[string]*domain.ServiceMetrics),
		windowSec: windowSec,
	}
}

func (e *Engine) RecordError(service string, timestamp int64) {
	e.mu.RLock()
	m, exists := e.metricMap[service]
	e.mu.RUnlock()

	if !exists {
		e.mu.Lock()
		m, exists = e.metricMap[service]
		if !exists {
			m = domain.NewServiceMetrics(e.windowSec)
			e.metricMap[service] = m
		}
		e.mu.Unlock()
	}

	m.RecordError(timestamp)
}

func (e *Engine) GetMetricMap() map[string]*domain.ServiceMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a shallow copy of the map to prevent concurrent map iteration issues
	snapshot := make(map[string]*domain.ServiceMetrics, len(e.metricMap))
	for k, v := range e.metricMap {
		snapshot[k] = v
	}
	return snapshot
}
