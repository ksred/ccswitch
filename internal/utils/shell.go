package utils

import (
	"os"
)

// IsShellIntegrationActive checks if we're running inside the shell wrapper
func IsShellIntegrationActive() bool {
	// The shell wrapper sets this environment variable (we'll add it)
	return os.Getenv("CCSWITCH_SHELL_WRAPPER") == "1"
}

// GetShellIntegrationInstructions returns instructions for setting up shell integration
func GetShellIntegrationInstructions() string {
	shell := os.Getenv("SHELL")
	if shell == "/bin/zsh" || shell == "/usr/bin/zsh" {
		return `To enable automatic directory switching, add this to your ~/.zshrc:
  eval "$(ccswitch shell-init)"
  
Then reload your shell:
  source ~/.zshrc`
	}
	
	return `To enable automatic directory switching, add this to your ~/.bashrc:
  eval "$(ccswitch shell-init)"
  
Then reload your shell:
  source ~/.bashrc`
}