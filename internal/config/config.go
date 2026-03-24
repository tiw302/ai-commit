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
	Provider      string            `json:"provider"`
	APIURL        string            `json:"api_url"`
	APIKey        string            `json:"api_key"`
	ModelName     string            `json:"model_name"`
	SystemPrompt  string            `json:"system_prompt,omitempty"`
	MaxDiffLength int               `json:"max_diff_length"`
	ExcludeFiles  []string          `json:"exclude_files"`
	UIColors      UIColors          `json:"ui_colors"`
	Modes         map[string]string `json:"modes"`
	DefaultMode   string            `json:"default_mode"`
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
			Provider:      "openai",
			APIURL:        "https://api.openai.com/v1/chat/completions",
			ModelName:     "gpt-4o",
			MaxDiffLength: 50000,
			ExcludeFiles: []string{
				"package-lock.json",
				"yarn.lock",
				"pnpm-lock.yaml",
				"go.sum",
				"*.svg",
				"*.png",
				"*.jpg",
				"*.pdf",
			},
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

	// Check for project-specific config by walking up the directory tree
	if projectConfigPath, err := findProjectConfig(); err == nil && projectConfigPath != "" {
		projectFile, err := os.ReadFile(projectConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read project config file: %w", err)
		}
		if err := json.Unmarshal(projectFile, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse project config JSON: %w", err)
		}
	}

	if envKey := os.Getenv("AI_COMMIT_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	return &cfg, nil
}

// findProjectConfig searches for .ai-commit.json starting from the current directory
// and walking up to the root. Returns the absolute path if found.
func findProjectConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, ".ai-commit.json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}
	return "", nil
}

// SaveConfig writes the configuration to the user's config directory.
func SaveConfig(cfg *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not determine user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "ai-commit")
	path := filepath.Join(appDir, "config.json")

	// Ensure app directory exists.
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
