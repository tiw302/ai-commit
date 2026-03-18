package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tiw302/ai-commit/internal/config"
)

// AIProvider is the interface that all AI backends must implement.
type AIProvider interface {
	GenerateCommitMessage(prompt, diff string) (string, error)
}

// NewProvider returns the appropriate AIProvider based on configuration.
func NewProvider(cfg *config.Config) (AIProvider, error) {
	switch cfg.Provider {
	case "openai":
		return &OpenAIProvider{cfg: cfg}, nil
	case "ollama":
		// TODO: Implement OllamaProvider
		return nil, fmt.Errorf("ollama provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
}

// Message represents a single chat message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the payload structure for Chat Completion APIs.
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// OpenAIResponse represents the structure of the API's JSON response.
type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// OpenAIProvider implements the AIProvider interface for OpenAI.
type OpenAIProvider struct {
	cfg *config.Config
}

// GenerateCommitMessage sends the git diff and the selected prompt to the AI provider.
// It returns the generated commit message or an error if the request fails.
func (p *OpenAIProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	// Construct the request payload.
	reqBody := OpenAIRequest{
		Model: p.cfg.ModelName,
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: "Here is the git diff of my changes:\n\n" + diff},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request data: %w", err)
	}

	// Prepare the HTTP request.
	req, err := http.NewRequest("POST", p.cfg.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Set the API Key (if provided in config or env).
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	// Execute the request with timeout.
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to the AI API: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned an error (Status %d): %s", resp.StatusCode, string(body))
	}

	var aiResp OpenAIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("failed to decode AI response: %w", err)
	}

	// Extract the generated message from the first choice.
	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("AI returned an empty response")
	}

	return aiResp.Choices[0].Message.Content, nil
}
