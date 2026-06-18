package main

import (
	"bytes" // Added missing import
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand" // Added missing import
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------- CONFIG & CONSTANTS ----------------

const (
	MaxWorkerPoolSize = 4
	LogChannelBuffer  = 5000
	SlidingWindowSec  = 60
	ErrorThreshold    = 3
)

// ---------------- DATA MODELS ----------------

type Log struct {
	Service   string `json:"service"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type Incident struct {
	Service   string `json:"service"`
	Type      string `json:"type"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Count     int    `json:"error_count"`
	UpdatedAt int64  `json:"updated_at"`
}

// ---------------- MICROSERVICE SIMULATOR ----------------

type SimulatorConfig struct {
	TargetURL      string
	TickInterval   time.Duration
	ActiveServices []string
}

type MicroserviceSimulator struct {
	cfg SimulatorConfig
}

func NewMicroserviceSimulator(cfg SimulatorConfig) *MicroserviceSimulator {
	return &MicroserviceSimulator{cfg: cfg}
}

func (ms *MicroserviceSimulator) Start(ctx context.Context) {
	log.Printf("📡 Microservice Simulator started hitting target: %s", ms.cfg.TargetURL)

	ticker := time.NewTicker(ms.cfg.TickInterval)
	defer ticker.Stop()

	spikeTracker := time.NewTicker(45 * time.Second)
	defer spikeTracker.Stop()

	client := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Microservice Simulator...")
			return
		case <-ticker.C:
			for _, service := range ms.cfg.ActiveServices {
				go func(svc string) {
					level := "info"
					msg := "API request completed successfully"

					if rand.Float32() < 0.02 {
						level = "error"
						msg = "Internal lookup failure / connection reset by peer"
					}

					ms.sendPayload(client, Log{
						Service: svc,
						Level:   level,
						Message: msg,
					})
				}(service)
			}
		case <-spikeTracker.C:
			targetService := ms.cfg.ActiveServices[rand.Intn(len(ms.cfg.ActiveServices))]
			log.Printf("🚨 [SIMULATOR] Inducing an error spike on '%s'...", targetService)

			for i := 0; i < 6; i++ {
				go func(svc string, count int) {
					ms.sendPayload(client, Log{
						Service: svc,
						Level:   "error",
						Message: fmt.Sprintf("CRITICAL: Out of memory scenario alert #%d", count),
					})
				}(targetService, i)
			}
		}
	}
}

func (ms *MicroserviceSimulator) sendPayload(client *http.Client, payload Log) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	resp, err := client.Post(ms.cfg.TargetURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return // Silently ignore drops during initialization or teardowns
	}
	defer resp.Body.Close()
}

// ---------------- SYSTEM STATE ----------------

type App struct {
	db            *sql.DB
	logChannel    chan Log
	mu            sync.RWMutex
	incidentState map[string][]int64
	wg            sync.WaitGroup
}

// ---------------- DB INIT (WAL MODE) ----------------

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./logmind.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	createTable := `
    CREATE TABLE IF NOT EXISTS logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service TEXT,
        level TEXT,
        message TEXT,
        timestamp INTEGER
    );
    CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
    `
	if _, err = db.Exec(createTable); err != nil {
		log.Fatalf("Failed to schema DB: %v", err)
	}
	return db
}

// ---------------- WORKERS ----------------

func (app *App) startWorkerPool(ctx context.Context) {
	for i := 0; i < MaxWorkerPoolSize; i++ {
		app.wg.Add(1)
		go func(workerID int) {
			defer app.wg.Done()
			for {
				select {
				case <-ctx.Done():
					for logData := range app.logChannel {
						app.processLog(logData)
					}
					return
				case logData, ok := <-app.logChannel:
					if !ok {
						return
					}
					app.processLog(logData)
				}
			}
		}(i)
	}
}

func (app *App) processLog(l Log) {
	_, err := app.db.Exec(
		"INSERT INTO logs(service, level, message, timestamp) VALUES (?, ?, ?, ?)",
		l.Service, l.Level, l.Message, l.Timestamp,
	)
	if err != nil {
		log.Printf("[ERROR] Database write failed: %v", err)
	}

	if l.Level == "error" {
		app.mu.Lock()
		app.incidentState[l.Service] = append(app.incidentState[l.Service], l.Timestamp)
		app.mu.Unlock()
	}
}

// ---------------- HTTP HANDLERS ----------------

func (app *App) logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var l Log
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if l.Service == "" || l.Level == "" {
		http.Error(w, "Missing mandatory fields", http.StatusUnprocessableEntity)
		return
	}

	l.Timestamp = time.Now().Unix()

	select {
	case app.logChannel <- l:
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"queued"}`))
	default:
		http.Error(w, "Server busy, queue overflow", http.StatusServiceUnavailable)
	}
}

func (app *App) incidentHandler(w http.ResponseWriter, r *http.Request) {
	var incidents []Incident
	now := time.Now().Unix()

	app.mu.RLock()
	for service, timestamps := range app.incidentState {
		count := len(timestamps)
		if count >= ErrorThreshold {
			incidents = append(incidents, Incident{
				Service:   service,
				Type:      "error_spike_realtime",
				Severity:  "high",
				Message:   fmt.Sprintf("Real-time error spike: %d errors in last %ds", count, SlidingWindowSec),
				Count:     count,
				UpdatedAt: now,
			})
		}
	}
	app.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"window_seconds": SlidingWindowSec,
		"incidents":      incidents,
	})
}

// ---------------- BACKGROUND JANITOR ----------------

func (app *App) startEvictionLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().Unix()
			app.mu.Lock()
			for service, timestamps := range app.incidentState {
				var validTimestamps []int64
				for _, t := range timestamps {
					if now-t <= SlidingWindowSec {
						validTimestamps = append(validTimestamps, t)
					}
				}

				if len(validTimestamps) == 0 {
					delete(app.incidentState, service)
				} else {
					app.incidentState[service] = validTimestamps
				}
			}
			app.mu.Unlock()
		}
	}
}

// ---------------- MAIN / LIFECYCLE ----------------

func main() {
	log.Println("Starting LogMind Engine...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &App{
		db:            initDB(),
		logChannel:    make(chan Log, LogChannelBuffer),
		incidentState: make(map[string][]int64),
	}
	defer app.db.Close()

	app.startWorkerPool(ctx)
	go app.startEvictionLoop(ctx)

	// Launch the simulator directly alongside the engine server
	go func() {
		time.Sleep(1 * time.Second) // Let server bind port first
		simulator := NewMicroserviceSimulator(SimulatorConfig{
			TargetURL:    "http://localhost:8080/logs",
			TickInterval: 400 * time.Millisecond,
			ActiveServices: []string{
				"payment-gateway",
				"auth-service",
				"inventory-api",
				"frontend-bff",
			},
		})
		simulator.Start(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/logs", app.logHandler)
	mux.HandleFunc("/incidents", app.incidentHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("HTTP Server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	<-shutdownSig
	log.Println("Shutdown signal received. Stopping gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	cancel()
	close(app.logChannel)
	app.wg.Wait()

	log.Println("System exited safely.")
}