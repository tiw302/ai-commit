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
	case "anthropic":
		return &AnthropicProvider{cfg: cfg}, nil
	case "gemini":
		return &GeminiProvider{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Provider)
	}
}

// Message represents a single chat message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the payload for Chat Completion APIs.
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// OpenAIResponse represents the structure of the API JSON response.
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

// OllamaRequest represents the request for Ollama's Chat API.
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

// AnthropicRequest represents the request for Anthropic Messages API.
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []Message `json:"messages"`
}

// AnthropicResponse represents the response structure from Anthropic's Messages API.
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
}

// AnthropicProvider implements the AIProvider interface for Anthropic Claude.
type AnthropicProvider struct {
	cfg *config.Config
}

func (p *AnthropicProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := AnthropicRequest{
		Model:     p.cfg.ModelName,
		MaxTokens: 1024,
		System:    prompt,
		Messages: []Message{
			{Role: "user", Content: "Here is the git diff of my changes:\n\n" + diff},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/messages"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error (Status %d): %s", resp.StatusCode, string(body))
	}

	var anthroResp AnthropicResponse
	if err := json.Unmarshal(body, &anthroResp); err != nil {
		return "", err
	}

	if len(anthroResp.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic")
	}

	return anthroResp.Content[0].Text, nil
}

// GeminiRequest represents the Google Generative Language API structure.
type GeminiRequest struct {
	Contents          []GeminiContent `json:"contents"`
	SystemInstruction *GeminiContent  `json:"system_instruction,omitempty"`
}

// GeminiContent represents a single content block in Gemini API.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a text or file part in Gemini API.
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure from Gemini API.
type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GeminiProvider implements the AIProvider interface for Google Gemini.
// It uses the standard Generative Language API.
type GeminiProvider struct {
	cfg *config.Config
}

// GenerateCommitMessage sends a prompt and diff to Gemini and returns the generated message.
func (p *GeminiProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	// Construct the request body following Google's specification.
	reqBody := GeminiRequest{
		SystemInstruction: &GeminiContent{
			Parts: []GeminiPart{{Text: prompt}},
		},
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: "Here is the git diff of my changes:\n\n" + diff}},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Default Gemini API URL template.
	// We use v1beta to support system instructions.
	apiURL := p.cfg.APIURL
	if apiURL == "" {
		// Use gemini-1.5-flash as a fast and cost-effective default.
		model := p.cfg.ModelName
		if model == "" {
			model = "gemini-1.5-flash"
		}
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	}

	// Append API key as a query parameter for Google API.
	fullURL := fmt.Sprintf("%s?key=%s", apiURL, p.cfg.APIKey)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp GeminiResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Error.Message != "" {
			return "", fmt.Errorf("gemini error (%s): %s", errResp.Error.Status, errResp.Error.Message)
		}
		return "", fmt.Errorf("gemini error (Status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if we have candidates and content.
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
