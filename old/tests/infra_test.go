package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// TestConfig holds common test configuration
type TestConfig struct {
	DBPath      string
	Registry    *tempolite.RegistryBuilder
	Destructive bool
}

// NewTestConfig creates a new test configuration with default settings
func NewTestConfig(t *testing.T, testName string) *TestConfig {
	dbPath := filepath.Join(".", "testdata", fmt.Sprintf("%s.db", testName))

	// Ensure the testdata directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	return &TestConfig{
		DBPath:      dbPath,
		Registry:    tempolite.NewRegistry(),
		Destructive: true,
	}
}

// NewTestTempolite creates a new Tempolite instance for testing
func NewTestTempolite(t *testing.T, cfg *TestConfig) *tempolite.Tempolite {
	tp, err := tempolite.New(
		context.Background(),
		cfg.Registry.Build(),
		tempolite.WithPath(cfg.DBPath),
		tempolite.WithDestructive(),
		// tempolite.WithDefaultLogLevel(slog.LevelDebug),
	)
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	return tp
}

// WaitWithTimeout waits for Tempolite operations with a timeout
func WaitWithTimeout(t *testing.T, tp *tempolite.Tempolite, timeout time.Duration) error {
	done := make(chan error)
	go func() {
		done <- tp.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for operations to complete")
	}
}
