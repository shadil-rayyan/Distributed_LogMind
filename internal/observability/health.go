package observability

import (
	"context"
	"net/http"
	"time"
)

var ReadyCheckFunc func(ctx context.Context) error

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if ReadyCheckFunc != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ReadyCheckFunc(ctx); err != nil {
			http.Error(w, "Service Unavailable: database ping failed: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

