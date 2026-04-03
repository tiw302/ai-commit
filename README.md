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
- **Modern Interactive TUI:** Polished terminal experience powered by `bubbletea`. Review, edit, or regenerate with arrow keys and shortcuts.
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

### Project-specific Configuration

You can place a `.ai-commit.json` file in your project root to override global settings. This is useful for team-wide configuration (e.g., specific models or exclude patterns).

**Example `.ai-commit.json`:**
```json
{
  "model_name": "gpt-4o",
  "max_diff_length": 50000,
  "exclude_files": [
    "package-lock.json",
    "yarn.lock",
    "go.sum"
  ],
  "default_mode": "pro"
}
```
*Note: Sensitive keys (like `api_key`) should generally be kept in your global user config or environment variables, not committed to the repository.*

## Usage

1. Stage your changes: `git add .`
2. Run the tool: `ai-commit`
3. Review the AI's suggestion using the **Interactive TUI**:
   - Use **Arrow Keys** or **j/k** to navigate.
   - Press **Enter** or **y** to accept and commit.
   - Press **e** to edit the message in your preferred editor.
   - Press **r** to try generating a new message.
   - Press **q** or **Esc** to cancel.

### CLI Flags

- `-m "context"`: Give the AI a hint (e.g., `ai-commit -m "fix UI bug"`).
- `--dry-run`: Print the commit message without committing.
- `--lang "th"`: The language for the commit message (e.g., en, th, jp).
- `--configure`: Run the interactive configuration wizard.
- `--mode`: Change the style (e.g., `ai-commit --mode troll`).
- `--install-hook`: Setup a git hook to run `ai-commit` automatically on `git commit`.
- `--uninstall-hook`: Remove the git hook.
- `-v, --version`: Show the version.

## Git Hook Integration

You can set up `ai-commit` to run automatically whenever you execute `git commit` (without the `-m` flag).

1. **Install the hook:**
   ```bash
   ai-commit --install-hook
   ```
2. **Usage:**
   Stage your changes (`git add .`) and then simply run:
   ```bash
   git commit
   ```
   `ai-commit` will be triggered to generate a message for you.

*Note: If you provide a message manually (e.g., `git commit -m "my message"`), the hook will be skipped automatically.*

## Shell Autocompletion

To enable shell autocompletion for `ai-commit`, follow these steps for your respective shell:

### Bash
1. Generate the completion script:
   ```bash
   ai-commit --completion bash > ~/.ai-commit-completion.sh
   ```
2. Source the script in your `~/.bashrc` or `~/.bash_profile`:
   ```bash
   echo 'source ~/.ai-commit-completion.sh' >> ~/.bashrc
   ```
3. Reload your shell: `source ~/.bashrc`

### Zsh
1. Generate the completion script:
   ```zsh
   ai-commit --completion zsh > ~/.ai-commit-completion.zsh
   ```
2. Add the script to your `~/.zshrc`:
   ```zsh
   echo 'source ~/.ai-commit-completion.zsh' >> ~/.zshrc
   ```
3. Reload your shell: `source ~/.zshrc`

### Fish
1. Generate the completion script:
   ```fish
   ai-commit --completion fish > ~/.config/fish/completions/ai-commit.fish
   ```
2. Reload your shell or open a new terminal.

After setting this up, you should be able to use tab completion for `ai-commit` commands and flags.

---

## Roadmap

- [x] **Multi-backend support** (OpenAI, Ollama, Anthropic, Gemini)
- [x] **Custom System Prompts** via config file
- [x] **Interactive Configuration Wizard** (`--configure`)
- [x] **Git Hook Integration** (Run automatically on `git commit`)
- [x] **Project-specific Configuration** (`.ai-commit.json` in repository root)
- [x] **Conventional Commits** (Better support and automatic scope detection)
- [x] **Multi-language Support** (Generate commit messages in your preferred language)
- [x] **Dry Run Mode** (`--dry-run` flag)
- [x] **Shell Autocompletion** (Bash, Zsh, Fish)
- [x] **Enhanced TUI** (Polished UI experience using `bubbletea`)
- [x] **Visual Diff Preview** (Rich syntax-highlighted diffs directly in the TUI)
- [ ] **Interactive Hunk Selection** (Pick specific code changes to commit within the tool)
- [ ] **AI-powered Changelog Generator** (Create `CHANGELOG.md` updates from your history)
- [ ] **Commit Analysis & Refinement** (Let the AI review and improve your manual commit messages)
- [ ] **Cost & Token Tracking** (Monitor usage and expenses for paid API providers)

---

## Contributing
Contributions are welcome! Please feel free to open an issue or submit a pull request for any bugs or feature requests.

## License

This project is licensed under the [MIT License](LICENSE) - see the [LICENSE](LICENSE) file for details.
