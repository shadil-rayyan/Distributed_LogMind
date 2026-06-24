package integration

import (
	"context"
	"testing"
	"time"

	"logmind/internal/config"
	"logmind/internal/detection"
	"logmind/internal/domain"
	"logmind/internal/storage"
)

func TestIncidentStateRestoresFromSQLite(t *testing.T) {
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

	base := time.Now().Unix()
	logs := []domain.Log{
		{Service: "payment", Level: "error", Message: "db timeout", Timestamp: base - 5},
		{Service: "payment", Level: "error", Message: "db timeout", Timestamp: base - 4},
		{Service: "payment", Level: "error", Message: "db timeout", Timestamp: base - 3},
		{Service: "search", Level: "info", Message: "healthy", Timestamp: base - 2},
	}

	if err := repo.InsertBatch(context.Background(), logs); err != nil {
		t.Fatalf("failed to seed logs: %v", err)
	}

	restoredEngine := detection.NewEngine(cfg.SlidingWindowSec)
	recent, err := repo.RecentErrors(context.Background(), base-int64(cfg.SlidingWindowSec))
	if err != nil {
		t.Fatalf("failed to load recent errors: %v", err)
	}
	restoredEngine.Replay(recent)

	metricMap := restoredEngine.GetMetricMap()
	metrics, ok := metricMap["payment"]
	if !ok {
		t.Fatal("expected payment service to be restored")
	}

	if count := metrics.GetActiveCount(base); count != 3 {
		t.Fatalf("expected 3 restored active errors, got %d", count)
	}

	if _, ok := metricMap["search"]; ok {
		t.Fatal("non-error logs should not create incident state")
	}
}
