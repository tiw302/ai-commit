package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Version is the current version of the ai-commit tool.
const Version = "0.1.0"

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

// LoadConfig reads the configuration from the user's config directory.
// It creates a default configuration if none exists.
func LoadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "ai-commit")
	path := filepath.Join(appDir, "config.json")

	// Ensure app directory exists.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		defaultCfg := &Config{
			APIURL:    "https://api.openai.com/v1/chat/completions",
			ModelName: "gpt-4o",
			Modes: map[string]string{
				"pro":   "You are a professional software engineer. Generate a concise commit message based on the diff below.",
				"troll": "You are a sarcastic dev. Roast the code and generate a funny commit message.",
			},
			DefaultMode: "pro",
		}

		data, _ := json.MarshalIndent(defaultCfg, "", "  ")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	if envKey := os.Getenv("AI_COMMIT_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	return &cfg, nil
}
