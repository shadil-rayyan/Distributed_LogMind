package config

import (
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
		LogChannelBuffer: getEnvAsInt("LOG_CHANNEL_BUFFER", 10000), // Reduced for memory optimization
		SlidingWindowSec: getEnvAsInt("SLIDING_WINDOW_SEC", 60),
		ErrorThreshold:   getEnvAsInt("ERROR_THRESHOLD", 3),
		BatchSize:        getEnvAsInt("BATCH_SIZE", 1000),
		BatchTimeout:     100 * time.Millisecond,
	}
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
