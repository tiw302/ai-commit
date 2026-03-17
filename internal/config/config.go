package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UIColors defines the hex or ANSI color codes for terminal output styling.
type UIColors struct {
	Success string `json:"success"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
	Info    string `json:"info"`
}

// Config represents the main configuration structure for the ai-commit tool.
type Config struct {
	APIURL      string            `json:"api_url"`
	APIKey      string            `json:"api_key"`
	ModelName   string            `json:"model_name"`
	UIColors    UIColors          `json:"ui_colors"`
	Modes       map[string]string `json:"modes"`
	DefaultMode string            `json:"default_mode"`
}

// LoadConfig reads the configuration from the OS-specific config directory (e.g., ~/.config/ai-commit/config.json).
// It prioritizes the AI_COMMIT_API_KEY environment variable if it is set.
func LoadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user config directory: %w", err)
	}

	path := filepath.Join(configDir, "ai-commit", "config.json")
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Environment variable takes precedence over the config file for security.
	if envKey := os.Getenv("AI_COMMIT_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	return &cfg, nil
}
