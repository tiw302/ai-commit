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
		return &OllamaProvider{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
}

// --- OpenAI Provider ---

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

func (p *OpenAIProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := OpenAIRequest{
		Model: p.cfg.ModelName,
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: "Here is the git diff of my changes:\n\n" + diff},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.cfg.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (Status %d): %s", resp.StatusCode, string(body))
	}

	var aiResp OpenAIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from AI")
	}

	return aiResp.Choices[0].Message.Content, nil
}

// --- Ollama Provider ---

// OllamaRequest represents the request payload for Ollama's Chat API.
type OllamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// OllamaResponse represents the response structure from Ollama's Chat API.
type OllamaResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// OllamaProvider implements the AIProvider interface for local Ollama.
type OllamaProvider struct {
	cfg *config.Config
}

func (p *OllamaProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := OllamaRequest{
		Model: p.cfg.ModelName,
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: "Here is the git diff of my changes:\n\n" + diff},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Default Ollama URL if not provided: http://localhost:11434/api/chat
	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/chat"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second} // Ollama might be slow on local machines
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama error (Status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Message.Content, nil
}
