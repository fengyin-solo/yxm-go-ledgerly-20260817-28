package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration from environment variables.
type Config struct {
	Addr       string
	DataDir    string
	AuthToken  string
	MaxBody    int64
	Timeout    time.Duration
	RateLimit  int
	RateWindow time.Duration
}

// Load returns a Config populated from LEDGERLY_* environment variables.
func Load() *Config {
	return &Config{
		Addr:       getEnv("LEDGERLY_ADDR", ":8080"),
		DataDir:    getEnv("LEDGERLY_DATA_DIR", ""),
		AuthToken:  getEnv("LEDGERLY_AUTH_TOKEN", ""),
		MaxBody:    int64Env("LEDGERLY_MAX_BODY", 4096),
		Timeout:    durationEnv("LEDGERLY_TIMEOUT", 10*time.Second),
		RateLimit:  intEnv("LEDGERLY_RATE_LIMIT", 100),
		RateWindow: durationEnv("LEDGERLY_RATE_WINDOW", 1*time.Minute),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func int64Env(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
