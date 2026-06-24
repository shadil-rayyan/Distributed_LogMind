package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"logmind/internal/domain"
)

type SimulatorConfig struct {
	TargetURL      string
	TickInterval   time.Duration
	SpikeInterval  time.Duration
	ActiveServices []string
}

type MicroserviceSimulator struct {
	cfg SimulatorConfig
}

func NewMicroserviceSimulator(cfg SimulatorConfig) *MicroserviceSimulator {
	return &MicroserviceSimulator{cfg: cfg}
}

func (ms *MicroserviceSimulator) Start(ctx context.Context) {
	slog.Info("Starting Microservice Log Traffic Simulator...", "target_url", ms.cfg.TargetURL, "tick_interval", ms.cfg.TickInterval)

	ticker := time.NewTicker(ms.cfg.TickInterval)
	defer ticker.Stop()

	spikeTracker := time.NewTicker(ms.cfg.SpikeInterval)
	defer spikeTracker.Stop()

	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Microservice Log Traffic Simulator...")
			return
		case <-ticker.C:
			for _, service := range ms.cfg.ActiveServices {
				go func(svc string) {
					level := "info"
					msg := "API request processed successfully"

					// 2% chance of a random organic error
					if rand.Float32() < 0.02 {
						level = "error"
						msg = "Internal lookup failure / connection reset by peer"
					}

					ms.sendPayload(client, domain.Log{
						Service: svc,
						Level:   level,
						Message: msg,
					})
				}(service)
			}
		case <-spikeTracker.C:
			if len(ms.cfg.ActiveServices) == 0 {
				continue
			}
			targetService := ms.cfg.ActiveServices[rand.Intn(len(ms.cfg.ActiveServices))]
			slog.Warn("🚨 [SIMULATOR] Inducing critical error spike for service", "service", targetService)

			// Generate an error burst (exceeding typical incident threshold)
			for i := 0; i < 6; i++ {
				go func(svc string, count int) {
					ms.sendPayload(client, domain.Log{
						Service: svc,
						Level:   "error",
						Message: fmt.Sprintf("CRITICAL: Out of memory scenario alert segment #%d", count),
					})
				}(targetService, i)
			}
		}
	}
}

func (ms *MicroserviceSimulator) sendPayload(client *http.Client, payload domain.Log) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	resp, err := client.Post(ms.cfg.TargetURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Debug("Failed to deliver simulated log payload", "error", err)
		return
	}
	defer resp.Body.Close()
}
