package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// current version
const Version = "0.1.0"

// TUI colors
type UIColors struct {
	Success string `json:"success"`
	Error   string `json:"error"`
	Warning string `json:"warning"`
	Info    string `json:"info"`
}

// app settings
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
	Language      string            `json:"language,omitempty"`
}

// load config from disk
func LoadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(configDir, "ai-commit")
	path := filepath.Join(appDir, "config.json")

	// init default config if missing
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return nil, err
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
				"pro":   "Conventional Commits style.",
				"troll": "Sarcastic developer style.",
			},
			DefaultMode: "pro",
			Language:    "en",
		}

		data, err := json.MarshalIndent(defaultCfg, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return nil, err
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	// merge project-specific config
	if projectPath, err := findProjectConfig(); err == nil && projectPath != "" {
		projectFile, err := os.ReadFile(projectPath)
		if err == nil {
			json.Unmarshal(projectFile, &cfg)
		}
	}

	if envKey := os.Getenv("AI_COMMIT_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	return &cfg, nil
}

// search for .ai-commit.json in parent dirs
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
			break
		}
		dir = parent
	}
	return "", nil
}

// save config to disk
func SaveConfig(cfg *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(configDir, "ai-commit")
	path := filepath.Join(appDir, "config.json")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
