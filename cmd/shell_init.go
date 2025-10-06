package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newShellInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-init",
		Short: "Output shell integration script",
		Long: `Output the shell integration script that enables automatic directory switching.

To install the shell integration:

For bash:
  echo 'eval "$(ccswitch shell-init)"' >> ~/.bashrc
  source ~/.bashrc

For zsh:
  echo 'eval "$(ccswitch shell-init)"' >> ~/.zshrc
  source ~/.zshrc`,
		Run: shellInit,
	}
}

func shellInit(cmd *cobra.Command, args []string) {
	// Detect shell type
	shell := os.Getenv("SHELL")
	isZsh := os.Getenv("ZSH_VERSION") != "" || shell == "/bin/zsh" || shell == "/usr/bin/zsh"

	if isZsh {
		outputZshInit()
	} else {
		outputBashInit()
	}
}

func outputBashInit() {
	// Output the shell wrapper function
	fmt.Print(`# ccswitch shell wrapper function
ccswitch() {
    case "$1" in
        list)
            # For list command, need to preserve TTY for interactive selector
            local temp_file=$(mktemp)
            
            # Run command with TTY preserved, redirect output to temp file
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@" | tee "$temp_file"
            
            # Extract and execute the cd command if session was selected
            local cd_cmd=$(grep "^cd " "$temp_file" 2>/dev/null | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi

            # Clean up temp file
            rm -f "$temp_file"
            ;;
        cleanup|info|shell-init)
            # These commands don't need special handling
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@"
            ;;
        create|*)
            # For session creation (default command and explicit create)
            # Need to preserve stdin for interactive input, then capture cd command
            local temp_file=$(mktemp)

            # Run command with stdin preserved, redirect output to temp file
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@" | tee "$temp_file"

            # Extract and execute the cd command if session was created successfully
            local cd_cmd=$(grep "^cd " "$temp_file" 2>/dev/null | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi
            
            # Clean up temp file
            rm -f "$temp_file"
            ;;
    esac
}

# Bash completion for ccswitch
_ccswitch_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Only provide completions for cleanup command
    if [[ "$prev" == "cleanup" ]]; then
        # Get list of sessions by running list non-interactively
        local sessions=$(CCSWITCH_SHELL_WRAPPER=1 command ccswitch list 2>&1 | grep -E "^→" | sed 's/^→ //' | awk '{print $1}')
        COMPREPLY=($(compgen -W "$sessions" -- "$cur"))
    elif [[ "$COMP_CWORD" -eq 1 ]]; then
        # Complete command names
        COMPREPLY=($(compgen -W "list cleanup create info shell-init" -- "$cur"))
    fi
}

# Enable bash completion if available
if [[ -n "$BASH_VERSION" ]]; then
    complete -F _ccswitch_completions ccswitch
fi
`)
}

func outputZshInit() {
	fmt.Print(`# ccswitch shell wrapper function
ccswitch() {
    case "$1" in
        list)
            # For list command, need to preserve TTY for interactive selector
            local temp_file=$(mktemp)

            # Run command with TTY preserved, redirect output to temp file
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@" | tee "$temp_file"

            # Extract and execute the cd command if session was selected
            local cd_cmd=$(grep "^cd " "$temp_file" 2>/dev/null | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi

            # Clean up temp file
            rm -f "$temp_file"
            ;;
        cleanup|info|shell-init)
            # These commands don't need special handling
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@"
            ;;
        create|*)
            # For session creation (default command and explicit create)
            local temp_file=$(mktemp)

            # Run command with stdin preserved, redirect output to temp file
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@" | tee "$temp_file"

            # Extract and execute the cd command if session was created successfully
            local cd_cmd=$(grep "^cd " "$temp_file" 2>/dev/null | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi

            # Clean up temp file
            rm -f "$temp_file"
            ;;
    esac
}

# Zsh completion for ccswitch
_ccswitch() {
    local -a commands sessions
    commands=(
        'list:List all sessions'
        'cleanup:Clean up a session'
        'create:Create a new session'
        'info:Show session info'
        'shell-init:Initialize shell integration'
    )

    case "$words[2]" in
        cleanup)
            # Get list of sessions for cleanup completion
            sessions=(${(f)"$(CCSWITCH_SHELL_WRAPPER=1 command ccswitch list 2>&1 | grep -E '^→' | sed 's/^→ //' | awk '{print $1}')"})
            _describe -t sessions 'session' sessions
            ;;
        *)
            if [[ $CURRENT -eq 2 ]]; then
                _describe -t commands 'ccswitch command' commands
            fi
            ;;
    esac
}

# Enable Zsh completion for ccswitch
compdef _ccswitch ccswitch
`)
}
