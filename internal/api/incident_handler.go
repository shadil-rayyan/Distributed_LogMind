package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/observability"
)

type IncidentHandler struct {
	engine *detection.Engine
	cfg    *config.Config
}

func NewIncidentHandler(engine *detection.Engine, cfg *config.Config) *IncidentHandler {
	return &IncidentHandler{engine: engine, cfg: cfg}
}

func (h *IncidentHandler) HandleIncidents(w http.ResponseWriter, r *http.Request) {
	var incidents []domain.Incident
	now := time.Now().Unix()

	metricMap := h.engine.GetMetricMap()

	for service, metrics := range metricMap {
		count := metrics.GetActiveCount(now)
		if count >= h.cfg.ErrorThreshold {
			incidents = append(incidents, domain.Incident{
				Service:   service,
				Type:      "error_spike_realtime",
				Severity:  "high",
				Message:   fmt.Sprintf("Real-time alert anomaly: detected %d faults within rolling window.", count),
				Count:     count,
				UpdatedAt: now,
			})
		}
	}

	observability.ActiveIncidents.Set(float64(len(incidents)))

	respData, err := json.Marshal(map[string]interface{}{
		"window_seconds": h.cfg.SlidingWindowSec,
		"incidents":      incidents,
	})
	if err != nil {
		http.Error(w, "Internal Server Error: failed to marshal incidents JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}
