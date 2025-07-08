package cmd

import (
	"fmt"

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
	// Output the shell wrapper function
	fmt.Print(`# ccswitch shell wrapper function
ccswitch() {
    case "$1" in
        list|cleanup|info|shell-init)
            # These commands don't need special handling
            CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@"
            ;;
        switch)
            # For switch command, capture output and execute cd command
            local output=$(CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@")
            echo "$output"
            
            # Extract and execute the cd command if switch was successful
            local cd_cmd=$(echo "$output" | grep "^cd " | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi
            ;;
        *)
            # For session creation (default command)
            local output=$(CCSWITCH_SHELL_WRAPPER=1 command ccswitch "$@")
            echo "$output"

            # Extract and execute the cd command if session was created successfully
            local cd_cmd=$(echo "$output" | grep "^cd " | tail -1)
            if [ -n "$cd_cmd" ]; then
                eval "$cd_cmd"
            fi
            ;;
    esac
}

# Bash completion for ccswitch
_ccswitch_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Only provide completions for switch and cleanup commands
    if [[ "$prev" == "switch" ]] || [[ "$prev" == "cleanup" ]]; then
        # Get list of sessions
        local sessions=$(command ccswitch list 2>/dev/null | grep -E "^[[:space:]]*[[:digit:]]+\." | awk '{print $2}' | grep -v "Active" | grep -v "No active")
        COMPREPLY=($(compgen -W "$sessions" -- "$cur"))
    elif [[ "$COMP_CWORD" -eq 1 ]]; then
        # Complete command names
        COMPREPLY=($(compgen -W "list switch cleanup info shell-init" -- "$cur"))
    fi
}

# Enable bash completion if available
if [[ -n "$BASH_VERSION" ]]; then
    complete -F _ccswitch_completions ccswitch
fi
`)
}