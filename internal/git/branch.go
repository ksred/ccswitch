package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// BranchManager handles git branch operations
type BranchManager struct {
	repoPath string
}

// NewBranchManager creates a new BranchManager
func NewBranchManager(repoPath string) *BranchManager {
	return &BranchManager{repoPath: repoPath}
}

// Create creates a new branch
func (bm *BranchManager) Create(name string) error {
	cmd := exec.Command("git", "branch", name)
	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch: %w, output: %s", err, string(output))
	}
	return nil
}

// Delete deletes a branch
func (bm *BranchManager) Delete(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, name)
	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w, output: %s", err, string(output))
	}
	return nil
}

// Exists checks if a branch exists
func (bm *BranchManager) Exists(name string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name) // #nosec G204
	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// GetCurrent returns the current branch name
func (bm *BranchManager) GetCurrent() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (bm *BranchManager) HasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = bm.repoPath
	output, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(output)) != ""
}
