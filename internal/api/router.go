package api

import (
	"net/http"
	"os"

	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/ingestion"
	"logmind/internal/observability"
	"logmind/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(cfg *config.Config, logHandler *ingestion.LogHandler, engine *detection.Engine) *http.ServeMux {
	mux := http.NewServeMux()

	incidentHandler := NewIncidentHandler(engine, cfg)

	limiter := ratelimit.NewLimiter(100, 200) // 100 req/sec limit, 200 burst limit

	// Apply Rate Limiting and CORS to endpoints
	mux.Handle("/logs", corsMiddleware(limiter.Limit(http.HandlerFunc(logHandler.HandleLogs))))
	mux.Handle("/incidents", corsMiddleware(limiter.Limit(http.HandlerFunc(incidentHandler.HandleIncidents))))

	// Serve the static frontend dashboard with robust path fallbacks
	mux.Handle("/", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		paths := []string{"index.html", "../index.html", "../../index.html", "/app/index.html"}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				http.ServeFile(w, r, p)
				return
			}
		}
		http.Error(w, "Dashboard frontend not found", http.StatusNotFound)
	})))

	// Observability endpoints
	mux.HandleFunc("/healthz", observability.HealthzHandler)
	mux.HandleFunc("/readyz", observability.ReadyzHandler)
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
