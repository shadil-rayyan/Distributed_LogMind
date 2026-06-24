package ingestion

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"logmind/internal/domain"
	"logmind/internal/observability"
)

type LogHandler struct {
	logChannel chan<- domain.Log
}

func NewLogHandler(ch chan<- domain.Log) *LogHandler {
	return &LogHandler{logChannel: ch}
}

func (h *LogHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	var l domain.Log
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	if l.Service == "" || l.Level == "" {
		http.Error(w, "Missing required criteria fields", http.StatusUnprocessableEntity)
		return
	}

	l.Timestamp = time.Now().Unix()

	select {
	case h.logChannel <- l:
		observability.LogsIngested.WithLabelValues(l.Service, l.Level).Inc()
		observability.QueueDepth.Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"queued"}`))
	default:
		slog.Warn("queue saturated, dropping log", "service", l.Service)
		http.Error(w, "Engine processing boundaries saturated. Dropping package.", http.StatusServiceUnavailable)
	}
}
