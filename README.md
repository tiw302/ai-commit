# ai-commit

A humble AI-powered git commit message generator written in Go.
Made with care by a developer who is still learning and exploring.

---

Hello! I am still quite new to the Go ecosystem, and I built this small tool to help myself (and hopefully others) write more consistent commit messages using AI. It is a simple project, but I am trying my best to make it useful and well-structured.

## Features

- Zero-Config: Works out of the box by creating a default configuration for you.
- Custom Modes: Support for different prompt modes like professional or casual.
- Interactive Editing: Option to edit the AI-generated message in your system editor.
- Smart Filtering: Automatically ignores large or non-text files to optimize API usage.
- Developer Friendly: Simple CLI interface with clear feedback and error handling.

## Quick Start

### Installation

1. Ensure you have Go installed on your machine.
2. Clone this repository: `git clone https://github.com/tiw/ai-commit.git`
3. Build the binary: `go build -o ai-commit ./cmd/ai-commit`
4. (Optional) Move it to your path: `sudo mv ai-commit /usr/local/bin/`

### Setup your API Key

You will need an OpenAI-compatible API key. I am currently learning how to support more providers, so please be patient.

Method 1: Environment Variable (Recommended)
```bash
export AI_COMMIT_API_KEY="your-key-here"
```

Method 2: Configuration File
Run the tool once, and it will automatically create a configuration file at:
`~/.config/ai-commit/config.json` (Linux/macOS)
Just open it and paste your API key there.
