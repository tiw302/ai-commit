# ai-commit 🤖

A tiny, humble AI-powered git commit message generator written in Go.  
*Made with ❤️ by a developer who's still learning and exploring!*

---

Hello! I'm just a beginner in the Go world, and I built this small tool to help myself (and maybe you!) write better, more consistent commit messages using AI. It's not perfect, but I'm trying my best to make it useful! 🌻

## 🚀 Quick Start

### Installation
1. Make sure you have Go installed on your machine.
2. Clone this repo: `git clone https://github.com/tiw/ai-commit.git`
3. Build the binary: `go build -o ai-commit ./cmd/ai-commit`
4. (Optional) Move it to your path: `sudo mv ai-commit /usr/local/bin/`

### Setup your API Key
You'll need an OpenAI-compatible API key (I'm still learning how to support more providers!). 

**Method 1: Environment Variable (Recommended)**
```bash
export AI_COMMIT_API_KEY="your-key-here"
```

**Method 2: Config File**
Run the tool once, and it will create a config file at:
`~/.config/ai-commit/config.json` (Linux/macOS)
Just open it and paste your key!

