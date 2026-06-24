package unit_test

import (
	"testing"
	"time"

	"logmind/internal/config"
)

func TestConfigValidateAcceptsValidValues(t *testing.T) {
	cfg := &config.Config{
		Port:             "8080",
		DBPath:           "./logmind.db",
		MaxWorkerPool:    4,
		LogChannelBuffer: 100,
		SlidingWindowSec: 60,
		ErrorThreshold:   3,
		BatchSize:        50,
		BatchTimeout:     100 * time.Millisecond,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to be valid, got error: %v", err)
	}
}

func TestConfigValidateRejectsInvalidValues(t *testing.T) {
	cfg := &config.Config{
		Port:             "8080",
		DBPath:           "./logmind.db",
		MaxWorkerPool:    0,
		LogChannelBuffer: 0,
		SlidingWindowSec: 0,
		ErrorThreshold:   0,
		BatchSize:        0,
		BatchTimeout:     0,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation to fail for zero values")
	}
}
