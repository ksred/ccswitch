package git

import (
	"fmt"
	"os"
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
	// Debug: log the command being executed
	fmt.Fprintf(os.Stderr, "DEBUG: git worktree add %s %s (from dir: %s)\n", path, branch, wm.repoPath)
	
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
	
	// First, find and add the main repository
	for _, wt := range worktrees {
		// Check if this is the main worktree (not in .ccswitch directory)
		if !strings.Contains(wt.Path, ".ccswitch") && wt.Branch != "" {
			// This is likely the main repository
			sessions = append(sessions, SessionInfo{
				Name:   "main",
				Branch: wt.Branch,
				Path:   wt.Path,
			})
			break // There should only be one main repository
		}
	}
	
	// Then add all ccswitch worktrees for this specific repo
	// The pattern should match worktrees that belong to this repository
	for _, wt := range worktrees {
		// Check if it's a ccswitch worktree and extract the repo name from path
		if strings.Contains(wt.Path, ".ccswitch/worktrees/") && wt.Branch != "" {
			// Extract repo name from path to ensure we only show worktrees for current repo
			parts := strings.Split(wt.Path, string(filepath.Separator))
			for i, part := range parts {
				if part == ".ccswitch" && i+2 < len(parts) && parts[i+1] == "worktrees" {
					if i+2 < len(parts) && parts[i+2] == repoName {
						sessionName := filepath.Base(wt.Path)
						sessions = append(sessions, SessionInfo{
							Name:   sessionName,
							Branch: wt.Branch,
							Path:   wt.Path,
						})
					}
					break
				}
			}
		}
	}
	
	return sessions
}