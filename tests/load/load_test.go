package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"logmind/internal/api"
	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/ingestion"
	"logmind/internal/storage"
)

func TestLoadIngestion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	cfg := &config.Config{
		Port:             "8080",
		DBPath:           ":memory:",
		MaxWorkerPool:    4,
		LogChannelBuffer: 10000,
		SlidingWindowSec: 60,
		ErrorThreshold:   3,
		BatchSize:        1000,
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

	var successCount int64
	var totalCount int64
	var rateLimitedCount int64

	concurrency := 10
	duration := 2 * time.Second
	deadline := time.Now().Add(duration)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 500 * time.Millisecond}
			logEntry := domain.Log{Service: fmt.Sprintf("service-%d", workerID), Level: "info", Message: "load testing payload"}
			bodyBytes, _ := json.Marshal(logEntry)

			for time.Now().Before(deadline) {
				atomic.AddInt64(&totalCount, 1)
				resp, err := client.Post(ts.URL+"/logs", "application/json", bytes.NewReader(bodyBytes))
				if err != nil {
					continue
				}
				if resp.StatusCode == http.StatusAccepted {
					atomic.AddInt64(&successCount, 1)
				} else if resp.StatusCode == http.StatusTooManyRequests {
					atomic.AddInt64(&rateLimitedCount, 1)
				}
				resp.Body.Close()
				time.Sleep(1 * time.Millisecond) // slight delay to avoid instant ratelimit saturation
			}
		}(i)
	}

	wg.Wait()

	t.Logf("--- Load Test Results ---")
	t.Logf("Total requests sent: %d", totalCount)
	t.Logf("Success requests (202 Accepted): %d", successCount)
	t.Logf("Rate Limited requests (429): %d", rateLimitedCount)
	t.Logf("Throughput: %.2f req/sec", float64(totalCount)/duration.Seconds())

	if successCount == 0 && totalCount > 0 {
		t.Error("Load test failed to register any successful logs")
	}
}
