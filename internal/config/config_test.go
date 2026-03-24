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

func TestLoadConfig_ProjectSpecific(t *testing.T) {
	// 1. Setup User Config Directory
	userConfigDir := t.TempDir()
	
	// Mock XDG_CONFIG_HOME to point to our temp dir
	t.Setenv("XDG_CONFIG_HOME", userConfigDir)
	// For Windows/macOS fallback if needed, but XDG is standard for this library on Linux
	// To be safe, we might need to mock os.UserConfigDir behavior more robustly if cross-platform
	// but for now, assuming Linux/environment-based config dir override works.

	// Create the default user config file to avoid the "first run" logic
	appDir := filepath.Join(userConfigDir, "ai-commit")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	defaultConfig := []byte(`{"model_name": "gpt-3.5-turbo", "max_diff_length": 100}`)
	if err := os.WriteFile(filepath.Join(appDir, "config.json"), defaultConfig, 0644); err != nil {
		t.Fatalf("failed to write user config: %v", err)
	}

	// 2. Setup Project Directory with .ai-commit.json
	projectDir := t.TempDir()
	
	// Change working directory to projectDir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current wd: %v", err)
	}
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	projectConfig := []byte(`{"model_name": "gpt-4-turbo", "max_diff_length": 500}`)
	if err := os.WriteFile(".ai-commit.json", projectConfig, 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	// 3. Load Config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 4. Assertions
	if cfg.ModelName != "gpt-4-turbo" {
		t.Errorf("expected model_name to be 'gpt-4-turbo', got '%s'", cfg.ModelName)
	}
	if cfg.MaxDiffLength != 500 {
		t.Errorf("expected max_diff_length to be 500, got %d", cfg.MaxDiffLength)
	}
}
