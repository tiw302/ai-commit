package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tiw302/ai-commit/internal/config"
)

func TestGenerateCommitMessage(t *testing.T) {
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
		APIURL:    mockServer.URL, // Point to mock server URL
		APIKey:    "fake-key",
		ModelName: "test-model",
	}

	// 3. Call the function
	msg, err := GenerateCommitMessage(cfg, "prompt", "diff")
	if err != nil {
		t.Fatalf("GenerateCommitMessage returned error: %v", err)
	}

	// 4. Check if we got the expected message
	expected := "feat: add unit tests"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}
