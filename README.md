# ai-commit ( •⌄• )✧

[![CI](https://github.com/tiw302/ai-commit/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/tiw302/ai-commit/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A humble but powerful **Universal AI-powered git commit generator** written in Go.
One tool to rule all your favorite AI providers: OpenAI, Anthropic, Google Gemini, and Local LLMs (Ollama).

---

Hello! I am still quite new to the Go ecosystem, and I built this tool to help myself (and hopefully others) write more consistent commit messages using the AI of your choice. ( ◡‿◡ *)

## Quick Start (3 Steps)

1. **Install:** `make install` (or `go build -o ai-commit ./cmd/ai-commit`)
2. **Setup:** `export AI_COMMIT_API_KEY="your-key-here"`
3. **Run:** `git add . && ai-commit`

---

## Features

- **Universal Provider Support:** Works with OpenAI, Anthropic (Claude), Google Gemini, and Ollama (Local LLMs).
- **OpenAI Compatible:** Supports any API that follows OpenAI's format (Groq, DeepSeek, OpenRouter, Mistral, etc.).
- **Zero-Config:** Automatically generates a default configuration file for you.
- **Custom Modes:** Support for different prompt styles (e.g., `pro` for serious work, `troll` for fun).
- **Interactive TUI:** Review, edit in your preferred editor, or regenerate the message instantly.
- **Smart Filtering:** Automatically ignores binaries and large files to save tokens.

## Configuration

The tool creates a config file at `~/.config/ai-commit/config.json`. You can switch providers easily:

### 1. OpenAI (or Groq, DeepSeek, OpenRouter)
```json
{
  "provider": "openai",
  "api_url": "https://api.openai.com/v1/chat/completions",
  "api_key": "YOUR_API_KEY",
  "model_name": "gpt-4o"
}
```
*Tip: To use **Groq**, just change `api_url` to `https://api.groq.com/openai/v1/chat/completions`.*

### 2. Anthropic (Claude)
```json
{
  "provider": "anthropic",
  "api_url": "https://api.anthropic.com/v1/messages",
  "api_key": "YOUR_ANTHROPIC_KEY",
  "model_name": "claude-3-5-sonnet-20240620"
}
```

### 3. Google Gemini
```json
{
  "provider": "gemini",
  "api_key": "YOUR_GEMINI_KEY",
  "model_name": "gemini-1.5-flash"
}
```

### 4. Ollama (Local LLM)
```json
{
  "provider": "ollama",
  "api_url": "http://localhost:11434/api/chat",
  "model_name": "llama3"
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
- `--mode`: Change the style (e.g., `ai-commit --mode troll`).
- `-v, --version`: Show the version.

## Roadmap

- [x] **Multi-backend support** (OpenAI, Ollama, Anthropic, Gemini)
- [ ] **Custom System Prompts** via config file
- [ ] **Git Hook Integration** (Run automatically on `git commit`)

---

## Contributing (｡◕‿◕｡)
I am just a beginner, so please be kind! If you find a bug or have an idea, feel free to open an issue or send a PR. I am always happy to learn from you! (✿◠‿◠)

## License

This project is licensed under the [MIT License](LICENSE) - see the [LICENSE](LICENSE) file for details. 
