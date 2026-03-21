# ai-commit

[![CI](https://github.com/tiw302/ai-commit/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/tiw302/ai-commit/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A humble but powerful **Universal AI-powered git commit generator** written in Go.
One tool to rule all your favorite AI providers: OpenAI, Anthropic, Google Gemini, and Local LLMs (Ollama).

---

`ai-commit` helps you write consistent, high-quality commit messages by leveraging Large Language Models (LLMs) to analyze your staged changes. It supports multiple providers and offers a flexible configuration system.

## Quick Start (3 Steps)

1. **Install:** `make install` (or `go build -o ai-commit ./cmd/ai-commit`)
2. **Setup:** `ai-commit --configure` (Interactive setup wizard)
3. **Run:** `git add . && ai-commit`

---

## Features

- **Universal Provider Support:** Works with OpenAI, Anthropic (Claude), Google Gemini, and Ollama (Local LLMs).
- **Interactive Configuration:** Easy setup wizard (`--configure`) to switch providers and keys instantly.
- **OpenAI Compatible:** Supports any API that follows OpenAI's format (Groq, DeepSeek, OpenRouter, Mistral, etc.).
- **Zero-Config:** Automatically generates a default configuration file.
- **Custom Modes:** Support for different prompt styles (e.g., `pro`, `conventional`).
- **Interactive TUI:** Review, edit in your preferred editor, or regenerate the message instantly.
- **Smart Filtering:** Automatically ignores binaries and large lockfiles to optimize token usage.

## Configuration

### Easy Setup (Recommended)
You can easily switch providers or update your API key using the built-in wizard:
```bash
ai-commit --configure
```

### Manual Configuration
The tool creates a config file at `~/.config/ai-commit/config.json`.

#### Supported Providers

**1. OpenAI (or Groq, DeepSeek, OpenRouter)**
```json
{
  "provider": "openai",
  "api_url": "https://api.openai.com/v1/chat/completions",
  "api_key": "YOUR_API_KEY",
  "model_name": "gpt-4o"
}
```

**2. Anthropic (Claude)**
```json
{
  "provider": "anthropic",
  "api_url": "https://api.anthropic.com/v1/messages",
  "api_key": "YOUR_ANTHROPIC_KEY",
  "model_name": "claude-3-5-sonnet-20240620"
}
```

**3. Google Gemini**
```json
{
  "provider": "gemini",
  "api_key": "YOUR_GEMINI_KEY",
  "model_name": "gemini-1.5-flash"
}
```

**4. Ollama (Local LLM)**
```json
{
  "provider": "ollama",
  "api_url": "http://localhost:11434/api/chat",
  "model_name": "llama3"
}
```

#### Advanced Settings

- **`system_prompt`**: Override the default system instructions globally.
- **`max_diff_length`**: Set the maximum number of characters for the diff (default: 50,000).
- **`exclude_files`**: List of glob patterns to ignore (e.g., `["*.lock", "*.svg"]`).

```json
{
  "system_prompt": "You are a senior developer. Write commit messages using Conventional Commits.",
  "max_diff_length": 50000
}
```

## Usage

1. Stage your changes: `git add .`
2. Run the tool: `ai-commit`
3. Review the AI's suggestion:
   - Press **y** to accept and commit.
   - Press **e** to edit the message first.
   - Press **r** to try generating a new one.

### CLI Flags

- `-m "context"`: Give the AI a hint (e.g., `ai-commit -m "fix UI bug"`).
- `--configure`: Run the interactive configuration wizard.
- `--mode`: Change the style (e.g., `ai-commit --mode troll`).
- `-v, --version`: Show the version.

## Roadmap

- [x] **Multi-backend support** (OpenAI, Ollama, Anthropic, Gemini)
- [x] **Custom System Prompts** via config file
- [x] **Interactive Configuration Wizard** (`--configure`)
- [ ] **Git Hook Integration** (Run automatically on `git commit`)

---

## Contributing
Contributions are welcome! Please feel free to open an issue or submit a pull request for any bugs or feature requests.

## License

This project is licensed under the [MIT License](LICENSE) - see the [LICENSE](LICENSE) file for details.
