package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksred/ccswitch/internal/config"
	"github.com/ksred/ccswitch/internal/errors"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/utils"
)

// Manager handles session operations
type Manager struct {
	worktreeManager *git.WorktreeManager
	branchManager   *git.BranchManager
	config          *config.Config
	repoPath        string
	repoName        string
}

// NewManager creates a new session manager
func NewManager(repoPath string) *Manager {
	// Get the main repository path to ensure we list all worktrees
	mainRepoPath, err := git.GetMainRepoPath(repoPath)
	if err != nil {
		// Fallback to the provided path if we can't get the main repo
		mainRepoPath = repoPath
	}
	
	repoName := filepath.Base(mainRepoPath)
	cfg, _ := config.Load()
	
	return &Manager{
		worktreeManager: git.NewWorktreeManager(mainRepoPath),
		branchManager:   git.NewBranchManager(repoPath), // Keep current path for branch operations
		config:          cfg,
		repoPath:        repoPath,
		repoName:        repoName,
	}
}

// CreateSession creates a new work session
func (m *Manager) CreateSession(description string) error {
	branchName := m.config.Branch.Prefix + utils.Slugify(description)
	sessionName := utils.Slugify(description)

	// Check if we're already on the branch we want to create
	currentBranch, err := m.branchManager.GetCurrent()
	if err == nil && currentBranch == branchName {
		return fmt.Errorf("%w: %s", errors.ErrAlreadyOnBranch, branchName)
	}

	// Check if branch already exists
	if m.branchManager.Exists(branchName) {
		return fmt.Errorf("%w: %s", errors.ErrBranchExists, branchName)
	}

	// Get worktree path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "failed to get home directory")
	}
	
	// Get repo name from the main repo path
	mainRepoPath, err := git.GetMainRepoPath(m.repoPath)
	if err != nil {
		mainRepoPath = m.repoPath
	}
	repoName := filepath.Base(mainRepoPath)
	
	worktreeBasePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName)
	worktreePath := filepath.Join(worktreeBasePath, sessionName)

	// Check if worktree directory already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("%w: %s", errors.ErrWorktreeExists, worktreePath)
	}

	// Ensure the worktree base directory exists
	if err := os.MkdirAll(worktreeBasePath, 0755); err != nil {
		return errors.Wrap(err, "failed to create worktree directory")
	}

	// Create branch
	if err := m.branchManager.Create(branchName); err != nil {
		return err
	}

	// Create worktree
	if err := m.worktreeManager.Create(worktreePath, branchName); err != nil {
		// Try to clean up the branch we just created
		m.branchManager.Delete(branchName, false)
		return err
	}

	return nil
}

// CheckoutSession creates a worktree for an existing branch
func (m *Manager) CheckoutSession(branchName string) error {
	sessionName := utils.Slugify(branchName)

	// Check if branch exists
	if !m.branchManager.Exists(branchName) {
		return fmt.Errorf("%w: %s", errors.ErrBranchNotFound, branchName)
	}

	// Check if we're already on the branch we want to checkout
	currentBranch, err := m.branchManager.GetCurrent()
	if err == nil && currentBranch == branchName {
		return fmt.Errorf("%w: %s", errors.ErrAlreadyOnBranch, branchName)
	}

	// Get worktree path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "failed to get home directory")
	}

	// Get repo name from the main repo path
	mainRepoPath, err := git.GetMainRepoPath(m.repoPath)
	if err != nil {
		mainRepoPath = m.repoPath
	}
	repoName := filepath.Base(mainRepoPath)

	worktreeBasePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName)
	worktreePath := filepath.Join(worktreeBasePath, sessionName)

	// Check if worktree directory already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("%w: %s", errors.ErrWorktreeExists, worktreePath)
	}

	// Ensure the worktree base directory exists
	if err := os.MkdirAll(worktreeBasePath, 0755); err != nil {
		return errors.Wrap(err, "failed to create worktree directory")
	}

	// Create worktree for existing branch
	if err := m.worktreeManager.Create(worktreePath, branchName); err != nil {
		return err
	}

	return nil
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() ([]git.SessionInfo, error) {
	worktrees, err := m.worktreeManager.List()
	if err != nil {
		return nil, err
	}
	return git.GetSessionsFromWorktrees(worktrees, m.repoName), nil
}

// RemoveSession removes a session and optionally its branch
func (m *Manager) RemoveSession(sessionPath string, deleteBranch bool, branchName string) error {
	// Remove worktree
	if err := m.worktreeManager.Remove(sessionPath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch if requested
	if deleteBranch && branchName != "" {
		if err := m.branchManager.Delete(branchName, false); err != nil {
			// Check if we need to force delete
			if strings.Contains(err.Error(), "not fully merged") {
				return m.branchManager.Delete(branchName, true)
			}
			return err
		}
	}

	return nil
}

// GetSessionPath returns the path for a session
func (m *Manager) GetSessionPath(sessionName string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ccswitch", "worktrees", m.repoName, sessionName)
}