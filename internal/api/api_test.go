package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tiw302/ai-commit/internal/config"
)

func TestOpenAIProvider_GenerateCommitMessage(t *testing.T) {
	// 1. Create a mock OpenAI API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response from OpenAI
		response := OpenAIResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{Message: Message{Role: "assistant", Content: "feat: add unit tests"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// 2. Prepare a test configuration pointing to our mock server
	cfg := &config.Config{
		Provider:  "openai",
		APIURL:    mockServer.URL, // Point to mock server URL
		APIKey:    "fake-key",
		ModelName: "test-model",
	}

	// 3. Initialize Provider
	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	// 4. Call the method
	msg, err := provider.GenerateCommitMessage("prompt", "diff")
	if err != nil {
		t.Fatalf("GenerateCommitMessage returned error: %v", err)
	}

	// 5. Check if we got the expected message
	expected := "feat: add unit tests"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestOllamaProvider_GenerateCommitMessage(t *testing.T) {
	// 1. Create a mock Ollama API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response from Ollama
		response := OllamaResponse{
			Model: "llama3",
			Message: Message{
				Role:    "assistant",
				Content: "refactor: implement interface",
			},
			Done: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// 2. Prepare configuration
	cfg := &config.Config{
		Provider:  "ollama",
		APIURL:    mockServer.URL,
		ModelName: "llama3",
	}

	// 3. Initialize Provider
	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	// 4. Call method
	msg, err := provider.GenerateCommitMessage("prompt", "diff")
	if err != nil {
		t.Fatalf("GenerateCommitMessage returned error: %v", err)
	}

	// 5. Verify
	expected := "refactor: implement interface"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestGeminiProvider_GenerateCommitMessage(t *testing.T) {
	// 1. Create a mock Gemini API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key is passed in URL
		if r.URL.Query().Get("key") != "fake-gemini-key" {
			t.Errorf("expected API key in query, got %q", r.URL.Query().Get("key"))
		}

		// Mock response from Gemini
		response := GeminiResponse{
			Candidates: []struct {
				Content GeminiContent `json:"content"`
			}{
				{
					Content: GeminiContent{
						Parts: []GeminiPart{{Text: "docs: update readme"}},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// 2. Prepare configuration
	cfg := &config.Config{
		Provider:  "gemini",
		APIURL:    mockServer.URL,
		APIKey:    "fake-gemini-key",
		ModelName: "gemini-1.5-flash",
	}

	// 3. Initialize Provider
	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	// 4. Call method
	msg, err := provider.GenerateCommitMessage("prompt", "diff")
	if err != nil {
		t.Fatalf("GenerateCommitMessage returned error: %v", err)
	}

	// 5. Verify
	expected := "docs: update readme"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}
