package ui

import (
	"fmt"
)

// BashCompletion is the bash completion script for ai-commit.
const BashCompletion = `
_ai_commit_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-v -version -configure -install-hook -uninstall-hook -hook -mode -m -dry-run -lang"

    case "${prev}" in
        -mode)
            # You could dynamically fetch modes here if needed
            COMPREPLY=( $(compgen -W "pro troll" -- ${cur}) )
            return 0
            ;;
        -lang)
            COMPREPLY=( $(compgen -W "en th jp" -- ${cur}) )
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}
complete -F _ai_commit_completion ai-commit
`

// ZshCompletion is the zsh completion script for ai-commit.
const ZshCompletion = `
#compdef ai-commit

_ai_commit() {
    local -a opts
    opts=(
        '-v[Print version and exit]'
        '-version[Print version and exit]'
        '-configure[Run the interactive configuration wizard]'
        '-install-hook[Install ai-commit as a git hook]'
        '-uninstall-hook[Uninstall the ai-commit git hook]'
        '-hook[Run in git hook mode (path to commit message file)]:file:_files'
        '-mode[The mode for the commit message (e.g., pro, troll)]:mode:(pro troll)'
        '-m[Short user context/instruction for the commit]:message:'
        '-dry-run[Print the commit message without committing]'
        '-lang[The language for the commit message (e.g., en, th, jp)]:language:(en th jp)'
    )

    _arguments -s : $opts
}

_ai_commit "$@"
`

// FishCompletion is the fish completion script for ai-commit.
const FishCompletion = `
complete -c ai-commit -s v -l version -d 'Print version and exit'
complete -c ai-commit -l configure -d 'Run the interactive configuration wizard'
complete -c ai-commit -l install-hook -d 'Install ai-commit as a git hook'
complete -c ai-commit -l uninstall-hook -d 'Uninstall the ai-commit git hook'
complete -c ai-commit -l hook -d 'Run in git hook mode (path to commit message file)' -r
complete -c ai-commit -l mode -d 'The mode for the commit message (e.g., pro, troll)' -r -f -a 'pro troll'
complete -c ai-commit -s m -d 'Short user context/instruction for the commit' -r
complete -c ai-commit -l dry-run -d 'Print the commit message without committing'
complete -c ai-commit -l lang -d 'The language for the commit message (e.g., en, th, jp)' -r -f -a 'en th jp'
`

// PrintCompletion prints the completion script for the given shell.
func PrintCompletion(shell string) {
	switch shell {
	case "bash":
		fmt.Print(BashCompletion)
	case "zsh":
		fmt.Print(ZshCompletion)
	case "fish":
		fmt.Print(FishCompletion)
	default:
		fmt.Printf("Unknown shell: %s. Supported shells: bash, zsh, fish\n", shell)
	}
}
