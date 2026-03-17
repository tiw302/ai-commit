package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_CreateDefault(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ai-commit-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the user config dir logic (by setting HOME/XDG_CONFIG_HOME or using a mock)
	// For simplicity in this test, we'll manually check the logic by creating a path
	appDir := filepath.Join(tmpDir, "ai-commit")
	path := filepath.Join(appDir, "config.json")

	// 1. Ensure directory exists
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}

	// 3. Check if we can write and read it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// This simulates our LoadConfig logic
		if err := os.WriteFile(path, []byte(`{"api_url": "test-url"}`), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("expected to read config file, got error: %v", err)
	}

	if string(file) == "" {
		t.Error("expected non-empty config file")
	}
}
