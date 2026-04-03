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
	// CLI flags
	flag.Usage = func() {
		fmt.Printf("ai-commit - A humble AI-powered git commit generator\n\n")
		fmt.Printf("Usage:\n")
		flag.PrintDefaults()
	}

	versionFlag := flag.Bool("v", false, "version info")
	versionFullFlag := flag.Bool("version", false, "version info")
	configureFlag := flag.Bool("configure", false, "run config wizard")
	installHookFlag := flag.Bool("install-hook", false, "install git hook")
	uninstallHookFlag := flag.Bool("uninstall-hook", false, "uninstall git hook")
	hookModeFlag := flag.String("hook", "", "git hook mode")
	modeFlag := flag.String("mode", "", "commit mode (e.g. pro, troll)")
	contextFlag := flag.String("m", "", "custom context")
	dryRunFlag := flag.Bool("dry-run", false, "dry run mode")
	langFlag := flag.String("lang", "", "output language")
	interactiveFlag := flag.Bool("i", false, "interactive mode to select files")
	completionFlag := flag.String("completion", "", "shell completion script")
	flag.Parse()

	if *versionFlag || *versionFullFlag {
		fmt.Printf("ai-commit version %s\n", config.Version)
		return
	}

	// shell completion
	if *completionFlag != "" {
		ui.PrintCompletion(*completionFlag)
		return
	}

	tui := ui.NewUI()

	// manage hooks
	if *installHookFlag {
		if err := git.InstallHook(); err != nil {
			tui.PrintError(fmt.Sprintf("hook install failed: %v", err))
			os.Exit(1)
		}
		tui.PrintSuccess("git hook installed!")
		return
	}

	if *uninstallHookFlag {
		if err := git.UninstallHook(); err != nil {
			tui.PrintError(fmt.Sprintf("hook uninstall failed: %v", err))
			os.Exit(1)
		}
		tui.PrintSuccess("git hook removed!")
		return
	}

	// hook mode check
	if *hookModeFlag != "" {
		args := flag.Args()
		if len(args) > 0 {
			source := args[0]
			// skip if message already exists or merging/squashing
			if source == "message" || source == "template" || source == "merge" || source == "squash" || source == "commit" {
				return
			}
		}
	}

	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		tui.PrintError(fmt.Sprintf("config error: %v", err))
		os.Exit(1)
	}

	// manual config
	if *configureFlag {
		runConfigurationWizard(tui, cfg)
		return
	}

	// initial setup
	if cfg.APIKey == "" {
		tui.PrintInfo("no API key found")
		choice := tui.PromptUser("run setup wizard? [Y/n]", "Y")
		choice = strings.ToLower(choice)
		
		if choice == "y" || choice == "yes" {
			runConfigurationWizard(tui, cfg)
			cfg, _ = config.LoadConfig()
		} else {
			if cfg.Provider != "ollama" {
				tui.PrintInfo(fmt.Sprintf("no key for %s. exit.", cfg.Provider))
				os.Exit(0)
			}
		}
	}

	tui.ApplyConfig(cfg.UIColors)

	// prompt mode
	var prompt string
	if *modeFlag != "" {
		var ok bool
		prompt, ok = cfg.Modes[*modeFlag]
		if !ok {
			tui.PrintError("unknown mode")
			os.Exit(1)
		}
	} else if cfg.SystemPrompt != "" {
		prompt = cfg.SystemPrompt
	} else {
		prompt = cfg.Modes[cfg.DefaultMode]
	}

	if *contextFlag != "" {
		prompt = fmt.Sprintf("%s\n\nUser Context: %s", prompt, *contextFlag)
	}

	// set language
	language := cfg.Language
	if *langFlag != "" {
		language = *langFlag
	}
	if language != "" && language != "en" {
		prompt = fmt.Sprintf("%s\n\nLanguage: %s", prompt, language)
	}

	// interactive staging
	if *interactiveFlag {
		staged, _ := git.GetStagedFiles()
		unstaged, _ := git.GetUnstagedFiles()
		
		items := tui.ShowStagingUI(staged, unstaged)
		if items != nil {
			for _, item := range items {
				if item.Selected && !item.IsStaged {
					git.StageFile(item.Name)
				} else if !item.Selected && item.IsStaged {
					git.UnstageFile(item.Name)
				}
			}
		}
	}

	// get staged diff
	diff, err := git.GetStagedDiff(cfg)
	if err != nil {
		// if nothing staged and not in interactive mode, try to show interactive
		if !*interactiveFlag {
			staged, _ := git.GetStagedFiles()
			unstaged, _ := git.GetUnstagedFiles()
			if len(unstaged) > 0 {
				tui.PrintInfo("no staged changes. opening interactive selector...")
				items := tui.ShowStagingUI(staged, unstaged)
				if items != nil {
					for _, item := range items {
						if item.Selected && !item.IsStaged {
							git.StageFile(item.Name)
						} else if !item.Selected && item.IsStaged {
							git.UnstageFile(item.Name)
						}
					}
					// retry getting diff
					diff, err = git.GetStagedDiff(cfg)
				}
			}
		}
		
		if err != nil {
			tui.PrintError(err.Error())
			os.Exit(1)
		}
	}

	// scope detection
	files, err := git.GetStagedFiles()
	if err == nil && len(files) > 0 {
		if scope := git.DetectScope(files); scope != "" {
			prompt = fmt.Sprintf("%s\n\nSuggested Scope: %s", prompt, scope)
		}
		prompt = fmt.Sprintf("%s\n\nFiles Modified:\n%s", prompt, strings.Join(files, "\n"))
	}

	// init AI
	provider, err := api.NewProvider(cfg)
	if err != nil {
		tui.PrintError(fmt.Sprintf("provider error: %v", err))
		os.Exit(1)
	}

	// gen loop
	var commitMessage string
	for {
		stopChan := make(chan bool)
		go ui.LoadingSpinner(stopChan)

		msg, err := provider.GenerateCommitMessage(prompt, diff)
		stopChan <- true
		if err != nil {
			tui.PrintError(fmt.Sprintf("gen failed: %v", err))
			os.Exit(1)
		}

		commitMessage = strings.TrimSpace(msg)
		fmt.Printf("\n\n%sProposed Commit Message:%s\n%s\n", tui.Info, "\033[0m", commitMessage)

		if *dryRunFlag {
			tui.PrintInfo("dry run. skip commit.")
			return
		}

		// handle action
		choice := tui.AskForConfirmation(diff)
		switch choice {
		case "y", "yes":
			if *hookModeFlag != "" {
				os.WriteFile(*hookModeFlag, []byte(commitMessage), 0644)
				return
			}
			if err := git.Commit(commitMessage); err != nil {
				tui.PrintError(err.Error())
				os.Exit(1)
			}
			tui.PrintSuccess("committed!")
			return
		case "e", "edit":
			editedMsg, _ := ui.OpenInEditor(commitMessage)
			if editedMsg == "" {
				tui.PrintInfo("empty message. cancel.")
				return
			}
			if *hookModeFlag != "" {
				os.WriteFile(*hookModeFlag, []byte(editedMsg), 0644)
				return
			}
			git.Commit(editedMsg)
			tui.PrintSuccess("committed!")
			return
		case "r", "regenerate":
			continue
		default:
			tui.PrintInfo("cancelled.")
			return
		}
	}
}

func runConfigurationWizard(tui *ui.UI, cfg *config.Config) {
	fmt.Println()
	tui.PrintInfo("config wizard")
	
	cfg.Provider = tui.PromptUser("provider (openai/anthropic/gemini/ollama)", cfg.Provider)

	currentKey := "none"
	if len(cfg.APIKey) > 8 {
		currentKey = "..." + cfg.APIKey[len(cfg.APIKey)-4:]
	}
	
	if apiKey := tui.PromptUser(fmt.Sprintf("API key (%s)", currentKey), ""); apiKey != "" {
		cfg.APIKey = apiKey
	}

	cfg.ModelName = tui.PromptUser("model name", cfg.ModelName)
	cfg.APIURL = tui.PromptUser("API URL (optional)", cfg.APIURL)
	cfg.SystemPrompt = tui.PromptUser("system prompt (optional)", cfg.SystemPrompt)
	cfg.Language = tui.PromptUser("language (en/th/jp)", cfg.Language)

	if err := config.SaveConfig(cfg); err != nil {
		tui.PrintError("save failed")
		os.Exit(1)
	}

	tui.PrintSuccess("config saved!")
}
