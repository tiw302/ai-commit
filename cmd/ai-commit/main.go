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
	configureFlag := flag.Bool("configure", false, "Run the interactive configuration wizard")
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

	// 0.5 Check for configure flag
	if *configureFlag {
		runConfigurationWizard(tui, cfg)
		return
	}

	// 0.6 Check for first-time setup (missing API Key)
	if cfg.APIKey == "" {
		tui.PrintInfo("It looks like this is your first time running ai-commit (or your API key is missing).")
		// Custom prompt for setup
		choice := tui.PromptUser("Would you like to run the setup wizard now? [Y/n]", "Y")
		choice = strings.ToLower(choice)
		
		if choice == "y" || choice == "yes" {
			runConfigurationWizard(tui, cfg)
			// Reload config after wizard completes to pick up new values
			var err error
			cfg, err = config.LoadConfig()
			if err != nil {
				tui.PrintError(fmt.Sprintf("Configuration reload error: %v", err))
				os.Exit(1)
			}
		} else {
			tui.PrintInfo("You can run 'ai-commit --configure' later to set up your API key.")
		}
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
		}
	}
}

func runConfigurationWizard(tui *ui.UI, cfg *config.Config) {
	fmt.Println()
	tui.PrintInfo("Welcome to the ai-commit configuration wizard!")
	fmt.Println("Press Enter to keep the current value in [brackets].")
	fmt.Println()

	// 1. Provider
	cfg.Provider = tui.PromptUser("Select AI Provider (openai, anthropic, gemini, ollama)", cfg.Provider)

	// 2. API Key
	// Show a masked version of the current key if it exists
	currentKeyDisplay := "none"
	if len(cfg.APIKey) > 8 {
		currentKeyDisplay = "..." + cfg.APIKey[len(cfg.APIKey)-4:]
	} else if len(cfg.APIKey) > 0 {
		currentKeyDisplay = "***"
	}
	
	apiKey := tui.PromptUser(fmt.Sprintf("Enter API Key (current: %s)", currentKeyDisplay), "")
	if apiKey != "" {
		cfg.APIKey = apiKey
	}
	// Note: If user presses Enter for API key and it was already set, we keep it.
	// But PromptUser with empty default returns empty string if user presses Enter.
	// So we need to handle "keep existing" manually if we don't pass the actual key as default.
	// Passing actual key as default to PromptUser would show it in plain text, which is bad.
	// So the previous logic `if apiKey != ""` was actually correct for the API Key specifically.
	// I will revert to that logic for API Key only.

	// 3. Model Name
	cfg.ModelName = tui.PromptUser("Enter Model Name", cfg.ModelName)

	// 4. API URL (Optional)
	cfg.APIURL = tui.PromptUser("Enter API URL (optional)", cfg.APIURL)

	// 5. System Prompt (Optional)
	cfg.SystemPrompt = tui.PromptUser("Enter Global System Prompt (optional)", cfg.SystemPrompt)

	// Save
	if err := config.SaveConfig(cfg); err != nil {
		tui.PrintError(fmt.Sprintf("Failed to save configuration: %v", err))
		os.Exit(1)
	}

	tui.PrintSuccess("Configuration saved successfully!")
}

