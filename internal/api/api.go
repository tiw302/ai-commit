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

// AI backend interface
type AIProvider interface {
	GenerateCommitMessage(prompt, diff string) (string, error)
}

// factory for AI providers
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
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}

// chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI structures
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// OpenAI implementation
type OpenAIProvider struct {
	cfg *config.Config
}

func (p *OpenAIProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := OpenAIRequest{
		Model: p.cfg.ModelName,
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: "diff:\n\n" + diff},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.cfg.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	body, err := doRequest(req, 30*time.Second)
	if err != nil {
		return "", err
	}

	var aiResp OpenAIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if aiResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", aiResp.Error.Message)
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("empty AI response")
	}

	return aiResp.Choices[0].Message.Content, nil
}

// Ollama implementation
type OllamaProvider struct {
	cfg *config.Config
}

type OllamaResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func (p *OllamaProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}{
		Model: p.cfg.ModelName,
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: "diff:\n\n" + diff},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/chat"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := doRequest(req, 60*time.Second)
	if err != nil {
		return "", err
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

// Anthropic implementation
type AnthropicProvider struct {
	cfg *config.Config
}

func (p *AnthropicProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		System    string    `json:"system"`
		Messages  []Message `json:"messages"`
	}{
		Model:     p.cfg.ModelName,
		MaxTokens: 1024,
		System:    prompt,
		Messages: []Message{
			{Role: "user", Content: "diff:\n\n" + diff},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/messages"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	body, err := doRequest(req, 30*time.Second)
	if err != nil {
		return "", err
	}

	var anthroResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &anthroResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if anthroResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", anthroResp.Error.Message)
	}

	if len(anthroResp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return anthroResp.Content[0].Text, nil
}

// Gemini structures
type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiRequest struct {
	Contents          []GeminiContent `json:"contents"`
	SystemInstruction *GeminiContent  `json:"system_instruction,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
}

// Gemini implementation
type GeminiProvider struct {
	cfg *config.Config
}

func (p *GeminiProvider) GenerateCommitMessage(prompt, diff string) (string, error) {
	reqBody := GeminiRequest{
		SystemInstruction: &GeminiContent{
			Parts: []GeminiPart{{Text: prompt}},
		},
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{{Text: "diff:\n\n" + diff}},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := p.cfg.APIURL
	if apiURL == "" {
		model := p.cfg.ModelName
		if model == "" {
			model = "gemini-1.5-flash"
		}
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	}

	fullURL := fmt.Sprintf("%s?key=%s", apiURL, p.cfg.APIKey)
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := doRequest(req, 30*time.Second)
	if err != nil {
		return "", err
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func doRequest(req *http.Request, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status code %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

