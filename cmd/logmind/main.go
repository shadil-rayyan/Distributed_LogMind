package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"logmind/internal/api"
	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/ingestion"
	"logmind/internal/observability"
	"logmind/internal/simulator"
	"logmind/internal/storage"
)

func init() {
	// Programmatic memory limit as a safety net if GOMEMLIMIT env var is not set.
	// We are targeting 128MB container, so we leave some headroom.
	debug.SetMemoryLimit(96 << 20) // 96 MiB soft limit
	debug.SetGCPercent(50)         // More aggressive GC
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Initializing LogMind High-Throughput Engine...")

	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := storage.InitDB(cfg.DBPath)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}

	observability.ReadyCheckFunc = db.PingContext

	repo := storage.NewSQLiteLogRepository(db)
	engine := detection.NewEngine(cfg.SlidingWindowSec)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := restoreIncidentState(startupCtx, repo, engine, cfg.SlidingWindowSec); err != nil {
		startupCancel()
		slog.Error("failed to restore incident state from SQLite", "error", err)
		os.Exit(1)
	}
	startupCancel()

	logChannel := make(chan domain.Log, cfg.LogChannelBuffer)
	logHandler := ingestion.NewLogHandler(logChannel)

	workerPool := ingestion.NewWorkerPool(cfg, repo, engine, logChannel)
	workerPool.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if os.Getenv("ENABLE_SIMULATOR") == "true" {
		sim := simulator.NewMicroserviceSimulator(simulator.SimulatorConfig{
			TargetURL:      "http://localhost:" + cfg.Port + "/logs",
			TickInterval:   400 * time.Millisecond,
			SpikeInterval:  45 * time.Second,
			ActiveServices: []string{"payment-gateway", "auth-service", "inventory-api", "frontend-bff"},
		})
		go sim.Start(ctx)
	}

	mux := api.NewRouter(cfg, logHandler, engine)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("HTTP Engine ingestion layer actively listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Catastrophic server binding panic", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownSig
	slog.Info("Termination vector identified. Initiating graceful engine shutdown sequence...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("Timeout while severing active web listeners", "error", err)
	}
	shutdownCancel()

	close(logChannel)
	workerPool.Wait()

	if err := db.Close(); err != nil {
		slog.Error("Encountered anomaly during data storage final flush", "error", err)
	}

	slog.Info("System state preserved. Exit completed successfully.")
}

func restoreIncidentState(ctx context.Context, repo storage.LogRepository, engine *detection.Engine, windowSec int) error {
	sinceUnix := time.Now().Add(-time.Duration(windowSec) * time.Second).Unix()
	logs, err := repo.RecentErrors(ctx, sinceUnix)
	if err != nil {
		return err
	}

	engine.Replay(logs)
	slog.Info("restored incident state from SQLite", "error_logs", len(logs), "window_seconds", windowSec)
	return nil
}
