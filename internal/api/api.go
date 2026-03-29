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

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", p.cfg.APIURL, bytes.NewBuffer(jsonData))
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
		return "", fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var aiResp OpenAIResponse
	json.Unmarshal(body, &aiResp)

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

	jsonData, _ := json.Marshal(reqBody)
	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/chat"
	}

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ollamaResp OllamaResponse
	json.Unmarshal(body, &ollamaResp)

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

	jsonData, _ := json.Marshal(reqBody)
	apiURL := p.cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/messages"
	}

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
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
	var anthroResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	json.Unmarshal(body, &anthroResp)

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

	jsonData, _ := json.Marshal(reqBody)
	apiURL := p.cfg.APIURL
	if apiURL == "" {
		model := p.cfg.ModelName
		if model == "" {
			model = "gemini-1.5-flash"
		}
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	}

	fullURL := fmt.Sprintf("%s?key=%s", apiURL, p.cfg.APIKey)
	req, _ := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var geminiResp GeminiResponse
	json.Unmarshal(body, &geminiResp)

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
