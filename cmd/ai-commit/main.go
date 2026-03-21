package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tiw302/ai-commit/internal/api"
	"github.com/tiw302/ai-commit/internal/config"
	"github.com/tiw302/ai-commit/internal/git"
	"github.com/tiw302/ai-commit/internal/ui"
)

func main() {
	// CLI Flags
	flag.Usage = func() {
		fmt.Printf("ai-commit - A humble AI-powered git commit generator\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  ai-commit [flags]\n\n")
		fmt.Printf("Examples:\n")
		fmt.Printf("  ai-commit\n")
		fmt.Printf("  ai-commit -m \"fix login bug\"\n")
		fmt.Printf("  ai-commit --mode troll\n\n")
		fmt.Printf("Flags:\n")
		flag.PrintDefaults()
	}

	versionFlag := flag.Bool("v", false, "Print version and exit")
	versionFullFlag := flag.Bool("version", false, "Print version and exit")
	modeFlag := flag.String("mode", "", "The mode for the commit message (e.g., pro, troll)")
	contextFlag := flag.String("m", "", "Short user context/instruction for the commit")
	flag.Parse()

	// 0. Check for version flag
	if *versionFlag || *versionFullFlag {
		fmt.Printf("ai-commit version %s\n", config.Version)
		return
	}

	tui := ui.NewUI()

	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		tui.PrintError(fmt.Sprintf("Configuration error: %v", err))
		os.Exit(1)
	}

	// Apply colors from config
	tui.ApplyConfig(cfg.UIColors)

	// 2. Determine Prompt Mode
	var prompt string

	if *modeFlag != "" {
		// CLI flag takes precedence
		var ok bool
		prompt, ok = cfg.Modes[*modeFlag]
		if !ok {
			tui.PrintError(fmt.Sprintf("Unknown mode '%s'. Check your config.json", *modeFlag))
			os.Exit(1)
		}
	} else if cfg.SystemPrompt != "" {
		// Use custom system prompt if defined in config
		prompt = cfg.SystemPrompt
	} else {
		// Fallback to default mode
		var ok bool
		prompt, ok = cfg.Modes[cfg.DefaultMode]
		if !ok {
			tui.PrintError(fmt.Sprintf("Unknown mode '%s'. Check your config.json", cfg.DefaultMode))
			os.Exit(1)
		}
	}

	// Append user context if provided
	if *contextFlag != "" {
		prompt = fmt.Sprintf("%s\n\nUser Context: %s", prompt, *contextFlag)
	}

	// 3. Get Staged Diff
	diff, err := git.GetStagedDiff(cfg)
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	// 4. Initialize AI Provider
	provider, err := api.NewProvider(cfg)
	if err != nil {
		tui.PrintError(fmt.Sprintf("AI Provider error: %v", err))
		os.Exit(1)
	}

	// 5. Generate Commit Message via AI
	var commitMessage string
	for {
		stopChan := make(chan bool)
		go ui.LoadingSpinner(stopChan)

		msg, err := provider.GenerateCommitMessage(prompt, diff)
		stopChan <- true
		if err != nil {
			tui.PrintError(fmt.Sprintf("AI Generation failed: %v", err))
			os.Exit(1)
		}

		commitMessage = strings.TrimSpace(msg)
		fmt.Printf("\n\n%sProposed Commit Message:%s\n%s\n", tui.Info, "\033[0m", commitMessage)

		// 6. Interactive Prompt
		choice := tui.AskForConfirmation()
		switch choice {
		case "y", "yes":
			if err := git.Commit(commitMessage); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to commit: %v", err))
				os.Exit(1)
			}
			tui.PrintSuccess("Changes committed successfully!")
			return
		case "e", "edit":
			editedMsg, err := ui.OpenInEditor(commitMessage)
			if err != nil {
				tui.PrintError(fmt.Sprintf("Editor error: %v", err))
				os.Exit(1)
			}
			if editedMsg == "" {
				tui.PrintInfo("Commit message is empty. Cancelled.")
				return
			}
			if err := git.Commit(editedMsg); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to commit: %v", err))
				os.Exit(1)
			}
			tui.PrintSuccess("Changes committed!")
			return
		case "r", "regenerate":
			tui.PrintInfo("Regenerating...")
			continue
		case "n", "no":
			tui.PrintInfo("Commit cancelled.")
			return
		default:
			tui.PrintInfo("Operation cancelled.")
			return
		}	}
}
