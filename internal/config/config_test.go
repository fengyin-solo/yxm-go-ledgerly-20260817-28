package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv()
	c := Load()
	if c.Addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %s", c.Addr)
	}
	if c.MaxBody != 4096 {
		t.Fatalf("expected default max body 4096, got %d", c.MaxBody)
	}
	if c.Timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %v", c.Timeout)
	}
	if c.RateLimit != 100 {
		t.Fatalf("expected default rate limit 100, got %d", c.RateLimit)
	}
	if c.RateWindow != 1*time.Minute {
		t.Fatalf("expected default rate window 1m, got %v", c.RateWindow)
	}
}

func TestLoadOverrides(t *testing.T) {
	clearEnv()
	os.Setenv("LEDGERLY_ADDR", ":9090")
	os.Setenv("LEDGERLY_DATA_DIR", "/tmp/ledgerly")
	os.Setenv("LEDGERLY_MAX_BODY", "8192")
	os.Setenv("LEDGERLY_TIMEOUT", "5s")
	os.Setenv("LEDGERLY_RATE_LIMIT", "50")
	os.Setenv("LEDGERLY_RATE_WINDOW", "30s")
	os.Setenv("LEDGERLY_AUTH_TOKEN", "myuser")
	defer clearEnv()

	c := Load()
	if c.Addr != ":9090" {
		t.Fatalf("expected :9090, got %s", c.Addr)
	}
	if c.DataDir != "/tmp/ledgerly" {
		t.Fatalf("expected /tmp/ledgerly, got %s", c.DataDir)
	}
	if c.MaxBody != 8192 {
		t.Fatalf("expected 8192, got %d", c.MaxBody)
	}
	if c.Timeout != 5*time.Second {
		t.Fatalf("expected 5s, got %v", c.Timeout)
	}
	if c.RateLimit != 50 {
		t.Fatalf("expected 50, got %d", c.RateLimit)
	}
	if c.RateWindow != 30*time.Second {
		t.Fatalf("expected 30s, got %v", c.RateWindow)
	}
	if c.AuthToken != "myuser" {
		t.Fatalf("expected myuser, got %s", c.AuthToken)
	}
}

func TestLoadBadInt64(t *testing.T) {
	clearEnv()
	os.Setenv("LEDGERLY_MAX_BODY", "notanumber")
	defer clearEnv()
	c := Load()
	if c.MaxBody != 4096 {
		t.Fatalf("expected fallback to 4096 for bad int64")
	}
}

func TestLoadBadDuration(t *testing.T) {
	clearEnv()
	os.Setenv("LEDGERLY_TIMEOUT", "bad")
	defer clearEnv()
	c := Load()
	if c.Timeout != 10*time.Second {
		t.Fatalf("expected fallback to 10s for bad duration")
	}
}

func clearEnv() {
	vars := []string{
		"LEDGERLY_ADDR",
		"LEDGERLY_DATA_DIR",
		"LEDGERLY_MAX_BODY",
		"LEDGERLY_TIMEOUT",
		"LEDGERLY_RATE_LIMIT",
		"LEDGERLY_RATE_WINDOW",
		"LEDGERLY_AUTH_TOKEN",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
