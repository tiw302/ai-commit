package git

import "testing"

func TestDetectScope(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"ui scope", []string{"internal/ui/button.go", "ui/styles.css"}, "ui"},
		{"api scope", []string{"internal/api/api.go", "api/handlers.go"}, "api"},
		{"docs scope", []string{"README.md", "docs/API.md"}, "docs"},
		{"cli scope", []string{"cmd/ai-commit/main.go"}, "cli"},
		{"config scope", []string{"internal/config/config.go"}, "config"},
		{"test scope", []string{"internal/config/config_test.go", "test/integration_test.go"}, "test"},
		{"mixed - ui dominant", []string{"internal/ui/a.go", "internal/ui/b.go", "internal/api/c.go"}, "ui"},
		{"unknown scope - top level pkg", []string{"pkg/util/helper.go"}, ""},
		{"custom scope", []string{"feature/login/logic.go"}, "feature"},
		{"build scope", []string{"go.mod", "Makefile"}, "build"},
		{"ci scope", []string{".github/workflows/ci.yml"}, "ci"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectScope(tt.files); got != tt.expected {
				t.Errorf("DetectScope(%v) = %v, want %v", tt.files, got, tt.expected)
			}
		})
	}
}
