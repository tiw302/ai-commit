package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tiw/ai-commit/api"
	"github.com/tiw/ai-commit/config"
	"github.com/tiw/ai-commit/git"
	"github.com/tiw/ai-commit/ui"
)

func main() {
	// CLI Flags
	modeFlag := flag.String("mode", "", "The mode for the commit message (e.g., pro, troll)")
	contextFlag := flag.String("m", "", "Short user context/instruction for the commit")
	flag.Parse()

	tui := ui.NewUI()

	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		tui.PrintError(fmt.Sprintf("Configuration error: %v", err))
		os.Exit(1)
	}

	// 2. Determine Prompt Mode
	mode := cfg.DefaultMode
	if *modeFlag != "" {
		mode = *modeFlag
	}

	prompt, ok := cfg.Modes[mode]
	if !ok {
		tui.PrintError(fmt.Sprintf("Unknown mode '%s'. Check your config.json", mode))
		os.Exit(1)
	}

	// Append user context if provided
	if *contextFlag != "" {
		prompt = fmt.Sprintf("%s\n\nUser Context: %s", prompt, *contextFlag)
	}

	// 3. Get Staged Diff
	diff, err := git.GetStagedDiff()
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	// 4. Generate Commit Message via AI
	var commitMessage string
	for {
		stopChan := make(chan bool)
		go ui.LoadingSpinner(stopChan)

		msg, err := api.GenerateCommitMessage(cfg, prompt, diff)
		stopChan <- true
		if err != nil {
			tui.PrintError(fmt.Sprintf("AI Generation failed: %v", err))
			os.Exit(1)
		}

		commitMessage = strings.TrimSpace(msg)
		fmt.Printf("\n\n%sProposed Commit Message:%s\n%s\n", tui.Info, "\033[0m", commitMessage)

		// 5. Interactive Prompt
		choice := ui.AskForConfirmation()
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
			if err := git.Commit(editedMsg); err != nil {
				tui.PrintError(fmt.Sprintf("Failed to commit: %v", err))
				os.Exit(1)
			}
			tui.PrintSuccess("Changes committed successfully with edited message!")
			return
		case "r", "regenerate":
			tui.PrintInfo("Regenerating...")
			continue
		case "n", "no":
			tui.PrintInfo("Commit cancelled.")
			return
		default:
			tui.PrintInfo("Unknown option. Operation cancelled.")
			return
		}
	}
}
