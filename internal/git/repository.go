package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepoName returns the repository name from the current directory
func GetRepoName(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	repoPath := strings.TrimSpace(string(output))
	return filepath.Base(repoPath), nil
}

// GetMainRepoPath returns the path to the main repository (not worktree)
func GetMainRepoPath(dir string) (string, error) {
	// First get the common git directory
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	gitDir := strings.TrimSpace(string(output))

	// If gitDir is just ".git", we're in the main repo already
	if gitDir == ".git" {
		cmd = exec.Command("git", "rev-parse", "--show-toplevel")
		cmd.Dir = dir
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}

	// The main repo path is the parent of the .git directory
	mainPath := filepath.Dir(gitDir)

	// If the path ends with .git, it's already correct
	// If not, we might be in the main repo already
	if !strings.HasSuffix(gitDir, ".git") {
		// We're likely in a bare repository or the main repo
		cmd = exec.Command("git", "rev-parse", "--show-toplevel")
		cmd.Dir = dir
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		mainPath = strings.TrimSpace(string(output))
	}

	return mainPath, nil
}

// IsGitRepository checks if the directory is a git repository
func IsGitRepository(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	if err == nil {
		return true
	}

	// Check if we're in a worktree or subdirectory
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err = cmd.Run()
	return err == nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
