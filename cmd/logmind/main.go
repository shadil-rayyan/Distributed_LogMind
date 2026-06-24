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
	// Configure structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Initializing LogMind High-Throughput Engine...")

	cfg := config.LoadConfig()

	db, err := storage.InitDB(cfg.DBPath)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}

	observability.ReadyCheckFunc = db.PingContext

	repo := storage.NewSQLiteLogRepository(db)
	engine := detection.NewEngine(cfg.SlidingWindowSec)
	
	logChannel := make(chan domain.Log, cfg.LogChannelBuffer)
	logHandler := ingestion.NewLogHandler(logChannel)

	workerPool := ingestion.NewWorkerPool(cfg, repo, engine, logChannel)
	workerPool.Start()

	// Context for background tasks (like simulator)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if os.Getenv("ENABLE_SIMULATOR") == "true" {
		sim := simulator.NewMicroserviceSimulator(simulator.SimulatorConfig{
			TargetURL:     "http://localhost:" + cfg.Port + "/logs",
			TickInterval:  400 * time.Millisecond,
			SpikeInterval: 45 * time.Second,
			ActiveServices: []string{
				"payment-gateway",
				"auth-service",
				"inventory-api",
				"frontend-bff",
			},
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

	cancel() // Stop background simulator first

	// Phase 1: Close down network listener edge
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("Timeout while severing active web listeners", "error", err)
	}
	shutdownCancel()

	// Phase 2: Safely close internal transport channels to trigger flushing of lingering worker queues
	close(logChannel)

	// Phase 3: Block wait on storage engine threads to completely write historical entries out
	workerPool.Wait()

	// Phase 4: Safely drop file locking references to storage structures
	if err := db.Close(); err != nil {
		slog.Error("Encountered anomaly during data storage final flush", "error", err)
	}

	slog.Info("System state preserved. Exit completed successfully.")
}
