# ai-commit ( •⌄• )✧

[![CI](https://github.com/tiw/ai-commit/actions/workflows/ci.yml/badge.svg)](https://github.com/tiw/ai-commit/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A humble AI-powered git commit message generator written in Go.
Made with care by a developer who is still learning and exploring.

---

Hello! I am still quite new to the Go ecosystem, and I built this small tool to help myself (and hopefully others) write more consistent commit messages using AI. It is a simple project, but I am trying my best to make it useful and well-structured. ( ◡‿◡ *)

## Quick Start (3 Steps)

1. **Install:** `make install` (or `go build -o ai-commit ./cmd/ai-commit`)
2. **Setup:** `export AI_COMMIT_API_KEY="your-key-here"`
3. **Run:** `git add . && ai-commit`

---

## Example Workflow

```bash
# 1. You made some changes to your code
# 2. Add them as usual
git add .

# 3. Just type:
ai-commit

# 4. Result:
# ? Accept this commit? [y]es / [n]o / [e]dit / [r]egenerate: y
# ✔ Changes committed successfully!
```

## Features

- Zero-Config: Works out of the box by creating a default configuration for you.
- Custom Modes: Support for different prompt modes like professional or casual.
- Interactive Editing: Option to edit the AI-generated message in your system editor.
- Smart Filtering: Automatically ignores large or non-text files to optimize API usage.
- Developer Friendly: Simple CLI interface with clear feedback, help documentation, and version tracking.

## Configuration

The tool creates a configuration file at `~/.config/ai-commit/config.json`. You can customize it to your heart's content:

```json
{
  "api_url": "https://api.openai.com/v1/chat/completions",
  "api_key": "",
  "model_name": "gpt-4o",
  "ui_colors": {
    "success": "\u001b[32m",
    "error": "\u001b[31m",
    "warning": "\u001b[33m",
    "info": "\u001b[34m"
  },
  "modes": {
    "pro": "You are a professional software engineer...",
    "troll": "You are a sarcastic dev..."
  },
  "default_mode": "pro"
}
```

## Usage

Using the tool is meant to be as simple as possible:

1. Stage your changes as usual: `git add .`
2. Run the tool: `ai-commit`
3. Review the AI's suggestion:
   - Press **y** to accept and commit.
   - Press **e** to edit the message first.
   - Press **r** to try generating a new one.

### CLI Flags

- `-m "context"`: Give the AI a hint (e.g., `ai-commit -m "fix UI bug"`).
- `--mode`: Change the style (e.g., `ai-commit --mode troll`).
- `-h, --help`: Show help and examples.
- `-v, --version`: Show the version.

## Development

I have included a few tools to help with development and ensure code quality:

```bash
make build    # Build the binary locally
make test     # Run the automated unit tests
make clean    # Remove build artifacts and clean cache
```

---

## Contributing (｡◕‿◕｡)
I am just a beginner, so please be kind! If you find a bug or have an idea to make this tool even better, feel free to open an issue or send a PR. I am always happy to learn from you! (✿◠‿◠)

## License
MIT
