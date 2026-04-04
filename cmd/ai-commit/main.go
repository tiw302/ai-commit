package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tiw302/ai-commit/internal/api"
	"github.com/tiw302/ai-commit/internal/config"
	"github.com/tiw302/ai-commit/internal/git"
	"github.com/tiw302/ai-commit/internal/ui"
)

func main() {
	// CLI flags
	flag.Usage = func() {
		fmt.Printf("ai-commit - A humble AI-powered git commit generator\n\n")
		fmt.Printf("Usage: ai-commit [flags]\n\n")
		fmt.Printf("Commands:\n")
		fmt.Printf("  --configure          Run the interactive configuration wizard\n")
		fmt.Printf("  --install-hook       Set up git hook to run ai-commit on git commit\n")
		fmt.Printf("  --uninstall-hook     Remove the git hook\n")
		fmt.Printf("  --changelog          Generate CHANGELOG.md from recent history\n")
		fmt.Printf("  --completion <shell> Generate shell completion script (bash, zsh, fish)\n\n")
		fmt.Printf("Options:\n")
		fmt.Printf("  -m <context>         Provide extra context for the AI\n")
		fmt.Printf("  --mode <mode>        Use a specific prompt mode (e.g., pro, troll)\n")
		fmt.Printf("  --lang <lang>        Specify output language (e.g., en, th, jp)\n")
		fmt.Printf("  -i                   Interactive mode to stage/unstage files\n")
		fmt.Printf("  --dry-run            Show proposed message without committing\n")
		fmt.Printf("  -v, --version        Show version information\n")
		fmt.Printf("\nExamples:\n")
		fmt.Printf("  $ git add . && ai-commit\n")
		fmt.Printf("  $ ai-commit -m \"fix login bug\"\n")
		fmt.Printf("  $ ai-commit --mode troll\n")
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
	changelogFlag := flag.Bool("changelog", false, "generate CHANGELOG.md from git history")
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
	if *installHookFlag || *uninstallHookFlag {
		handleHookManagement(tui, *installHookFlag)
		return
	}

	// changelog mode
	if *changelogFlag {
		handleChangelogGeneration(tui, *langFlag, *dryRunFlag)
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
		ensureAPIKey(tui, cfg)
	}

	tui.ApplyConfig(cfg.UIColors)

	// prompt mode
	prompt := resolvePrompt(tui, cfg, *modeFlag, *contextFlag, *langFlag)

	// interactive staging
	if *interactiveFlag {
		handleInteractiveStaging(tui)
	}

	// get staged diff
	diff, err := git.GetStagedDiff(cfg)
	if err != nil {
		diff, err = handleNoStagedChanges(tui, cfg, *interactiveFlag)
		if err != nil {
			tui.PrintError(err.Error())
			os.Exit(1)
		}
	}

	// scope detection
	prompt = enrichPromptWithScope(prompt)

	// init AI
	provider, err := api.NewProvider(cfg)
	if err != nil {
		tui.PrintError(fmt.Sprintf("provider error: %v", err))
		os.Exit(1)
	}

	// gen loop
	generateAndCommit(tui, provider, cfg, prompt, diff, *dryRunFlag, *hookModeFlag)
}

func handleHookManagement(tui *ui.UI, install bool) {
	if install {
		if err := git.InstallHook(); err != nil {
			tui.PrintError(fmt.Sprintf("hook install failed: %v", err))
			os.Exit(1)
		}
		tui.PrintSuccess("git hook installed!")
	} else {
		if err := git.UninstallHook(); err != nil {
			tui.PrintError(fmt.Sprintf("hook uninstall failed: %v", err))
			os.Exit(1)
		}
		tui.PrintSuccess("git hook removed!")
	}
}

func handleChangelogGeneration(tui *ui.UI, lang string, dryRun bool) {
	cfg, err := config.LoadConfig()
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	commits, err := git.GetRecentCommits(30)
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	provider, err := api.NewProvider(cfg)
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	prompt := "Generate a CHANGELOG.md update based on these recent git commits. Group them into relevant sections (Features, Bug Fixes, etc.). Only output the markdown content for the changelog."
	if lang != "" {
		prompt = fmt.Sprintf("%s\n\nLanguage: %s", prompt, lang)
	}

	stopChan := make(chan bool)
	go func() {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-stopChan:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r\033[36m%s\033[0m generating changelog...", spinner[i])
				i = (i + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	changelog, err := provider.GenerateCommitMessage(prompt, commits)
	stopChan <- true
	if err != nil {
		tui.PrintError(err.Error())
		os.Exit(1)
	}

	fmt.Printf("\n\n%sProposed Changelog:%s\n%s\n", tui.Info, "\033[0m", changelog)

	if dryRun {
		return
	}

	choice := tui.PromptUser("save to CHANGELOG.md? [y/N]", "n")
	if strings.ToLower(choice) == "y" {
		err := os.WriteFile("CHANGELOG.md", []byte(changelog), 0644)
		if err != nil {
			tui.PrintError(err.Error())
			os.Exit(1)
		}
		tui.PrintSuccess("CHANGELOG.md updated!")
	}
}

func ensureAPIKey(tui *ui.UI, cfg *config.Config) {
	tui.PrintInfo("no API key found")
	choice := tui.PromptUser("run setup wizard? [Y/n]", "Y")
	choice = strings.ToLower(choice)
	
	if choice == "y" || choice == "yes" {
		runConfigurationWizard(tui, cfg)
		// reload config after wizard
		if newCfg, err := config.LoadConfig(); err == nil {
			*cfg = *newCfg
		}
	} else {
		if cfg.Provider != "ollama" {
			tui.PrintInfo(fmt.Sprintf("no key for %s. exit.", cfg.Provider))
			os.Exit(0)
		}
	}
}

func resolvePrompt(tui *ui.UI, cfg *config.Config, mode, context, lang string) string {
	var prompt string
	if mode != "" {
		var ok bool
		prompt, ok = cfg.Modes[mode]
		if !ok {
			tui.PrintError("unknown mode")
			os.Exit(1)
		}
	} else if cfg.SystemPrompt != "" {
		prompt = cfg.SystemPrompt
	} else {
		prompt = cfg.Modes[cfg.DefaultMode]
	}

	if context != "" {
		prompt = fmt.Sprintf("%s\n\nUser Context: %s", prompt, context)
	}

	// set language
	language := cfg.Language
	if lang != "" {
		language = lang
	}
	if language != "" && language != "en" {
		prompt = fmt.Sprintf("%s\n\nLanguage: %s", prompt, language)
	}
	return prompt
}

func handleInteractiveStaging(tui *ui.UI) {
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

func handleNoStagedChanges(tui *ui.UI, cfg *config.Config, alreadyInteractive bool) (string, error) {
	if !alreadyInteractive {
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
				return git.GetStagedDiff(cfg)
			}
		}
	}
	return "", fmt.Errorf("no staged changes")
}

func enrichPromptWithScope(prompt string) string {
	files, err := git.GetStagedFiles()
	if err == nil && len(files) > 0 {
		if scope := git.DetectScope(files); scope != "" {
			prompt = fmt.Sprintf("%s\n\nSuggested Scope: %s", prompt, scope)
		}
		prompt = fmt.Sprintf("%s\n\nFiles Modified:\n%s", prompt, strings.Join(files, "\n"))
	}
	return prompt
}

func generateAndCommit(tui *ui.UI, provider api.AIProvider, cfg *config.Config, prompt, diff string, dryRun bool, hookMode string) {
	for {
		stopChan := make(chan bool)
		go ui.LoadingSpinner(stopChan)

		msg, err := provider.GenerateCommitMessage(prompt, diff)
		stopChan <- true
		if err != nil {
			tui.PrintError(fmt.Sprintf("gen failed: %v", err))
			os.Exit(1)
		}

		commitMessage := strings.TrimSpace(msg)
		fmt.Printf("\n\n%sProposed Commit Message:%s\n%s\n", tui.Info, "\033[0m", commitMessage)

		if dryRun {
			tui.PrintInfo("dry run. skip commit.")
			return
		}

		// handle action
		choice := tui.AskForConfirmation(diff)
		switch choice {
		case "y", "yes":
			if hookMode != "" {
				if err := os.WriteFile(hookMode, []byte(commitMessage), 0644); err != nil {
					tui.PrintError(fmt.Sprintf("failed to write commit message: %v", err))
					os.Exit(1)
				}
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
			if hookMode != "" {
				if err := os.WriteFile(hookMode, []byte(editedMsg), 0644); err != nil {
					tui.PrintError(fmt.Sprintf("failed to write commit message: %v", err))
					os.Exit(1)
				}
				return
			}
			if err := git.Commit(editedMsg); err != nil {
				tui.PrintError(err.Error())
				os.Exit(1)
			}
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
