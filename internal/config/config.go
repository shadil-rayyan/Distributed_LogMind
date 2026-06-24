package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	DBPath           string
	MaxWorkerPool    int
	LogChannelBuffer int
	SlidingWindowSec int
	ErrorThreshold   int
	BatchSize        int
	BatchTimeout     time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DBPath:           getEnv("LOGMIND_DB_PATH", "./logmind.db"),
		MaxWorkerPool:    getEnvAsInt("MAX_WORKERS", 4),
		LogChannelBuffer: getEnvAsInt("LOG_CHANNEL_BUFFER", 10000),
		SlidingWindowSec: getEnvAsInt("SLIDING_WINDOW_SEC", 60),
		ErrorThreshold:   getEnvAsInt("ERROR_THRESHOLD", 3),
		BatchSize:        getEnvAsInt("BATCH_SIZE", 1000),
		BatchTimeout:     100 * time.Millisecond,
	}
}

func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT must not be empty")
	}
	if c.DBPath == "" {
		return fmt.Errorf("LOGMIND_DB_PATH must not be empty")
	}
	if c.MaxWorkerPool <= 0 {
		return fmt.Errorf("MAX_WORKERS must be greater than zero")
	}
	if c.LogChannelBuffer <= 0 {
		return fmt.Errorf("LOG_CHANNEL_BUFFER must be greater than zero")
	}
	if c.SlidingWindowSec <= 0 {
		return fmt.Errorf("SLIDING_WINDOW_SEC must be greater than zero")
	}
	if c.ErrorThreshold <= 0 {
		return fmt.Errorf("ERROR_THRESHOLD must be greater than zero")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE must be greater than zero")
	}
	if c.BatchTimeout <= 0 {
		return fmt.Errorf("BATCH_TIMEOUT must be greater than zero")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
