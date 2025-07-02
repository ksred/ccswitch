package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// WorktreeManager handles git worktree operations
type WorktreeManager struct {
	repoPath string
}

// NewWorktreeManager creates a new WorktreeManager
func NewWorktreeManager(repoPath string) *WorktreeManager {
	return &WorktreeManager{repoPath: repoPath}
}

// Create creates a new worktree
func (wm *WorktreeManager) Create(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	cmd.Dir = wm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w, output: %s", err, string(output))
	}
	return nil
}

// List returns all worktrees
func (wm *WorktreeManager) List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = wm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	return ParseWorktrees(string(output)), nil
}

// Remove removes a worktree
func (wm *WorktreeManager) Remove(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path, "--force")
	cmd.Dir = wm.repoPath
	_, err := cmd.CombinedOutput()
	return err
}

// ParseWorktrees parses git worktree list --porcelain output
func ParseWorktrees(output string) []Worktree {
	var worktrees []Worktree
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	var currentWorktree Worktree
	branchRegex := regexp.MustCompile(`^branch refs/heads/(.+)$`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentWorktree.Path != "" {
				worktrees = append(worktrees, currentWorktree)
				currentWorktree = Worktree{}
			}
			continue
		}
		
		if strings.HasPrefix(line, "worktree ") {
			currentWorktree.Path = strings.TrimPrefix(line, "worktree ")
		} else if matches := branchRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentWorktree.Branch = matches[1]
		} else if strings.HasPrefix(line, "HEAD ") {
			currentWorktree.Commit = strings.TrimPrefix(line, "HEAD ")
		}
	}
	
	if currentWorktree.Path != "" {
		worktrees = append(worktrees, currentWorktree)
	}
	
	return worktrees
}

// GetSessionsFromWorktrees extracts session information from worktrees
func GetSessionsFromWorktrees(worktrees []Worktree, repoName string) []SessionInfo {
	var sessions []SessionInfo
	pattern := filepath.Join(".ccswitch", "worktrees", repoName)
	
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pattern) && wt.Branch != "" {
			sessionName := filepath.Base(wt.Path)
			sessions = append(sessions, SessionInfo{
				Name:   sessionName,
				Branch: wt.Branch,
				Path:   wt.Path,
			})
		}
	}
	
	return sessions
}