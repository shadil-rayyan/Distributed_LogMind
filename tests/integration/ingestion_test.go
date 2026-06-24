package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"logmind/internal/api"
	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/ingestion"
	"logmind/internal/storage"
)

func TestLogIngestionAndIncidentDetection(t *testing.T) {
	cfg := &config.Config{
		Port:             "8080",
		DBPath:           ":memory:",
		MaxWorkerPool:    1,
		LogChannelBuffer: 10,
		SlidingWindowSec: 60,
		ErrorThreshold:   3,
		BatchSize:        5,
		BatchTimeout:     50 * time.Millisecond,
	}

	db, err := storage.InitDB(cfg.DBPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	repo := storage.NewSQLiteLogRepository(db)
	engine := detection.NewEngine(cfg.SlidingWindowSec)
	logChannel := make(chan domain.Log, cfg.LogChannelBuffer)
	logHandler := ingestion.NewLogHandler(logChannel)

	workerPool := ingestion.NewWorkerPool(cfg, repo, engine, logChannel)
	workerPool.Start()
	defer func() {
		close(logChannel)
		workerPool.Wait()
	}()

	mux := api.NewRouter(cfg, logHandler, engine)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 1. POST 5 error logs from same service
	for i := 0; i < 5; i++ {
		logEntry := domain.Log{Service: "payment", Level: "error"}
		body, _ := json.Marshal(logEntry)
		resp, err := http.Post(ts.URL+"/logs", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("post log failed: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected 202 Accepted, got %d", resp.StatusCode)
		}
	}

	// wait for worker to flush
	time.Sleep(100 * time.Millisecond)

	// 2. GET /incidents
	resp, err := http.Get(ts.URL + "/incidents")
	if err != nil {
		t.Fatalf("get incidents failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Incidents []domain.Incident `json:"incidents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if len(result.Incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(result.Incidents))
	}
	if result.Incidents[0].Service != "payment" {
		t.Errorf("expected incident for payment, got %s", result.Incidents[0].Service)
	}

	// 3. GET / (Dashboard HTML frontend)
	respHtml, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get root page failed: %v", err)
	}
	defer respHtml.Body.Close()

	if respHtml.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", respHtml.StatusCode)
	}
}
