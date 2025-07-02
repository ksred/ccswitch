package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Helper to capture stdout during command execution
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	return string(out)
}

// Helper to mock stdin
func mockStdin(input string, f func()) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte(input))
		w.Close()
	}()

	f()
	os.Stdin = oldStdin
}

func TestCreateSessionCommand(t *testing.T) {
	// Setup test repo
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)
	
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	tests := []struct {
		name        string
		input       string
		expectError bool
		contains    []string
		notContains []string
	}{
		{
			name:        "empty description",
			input:       "\n",
			expectError: true,
			contains:    []string{"What are you working on?", "Description cannot be empty"},
			notContains: []string{"Created session"},
		},
		{
			name:        "valid description",
			input:       "Fix authentication bug\n",
			expectError: false,
			contains:    []string{
				"What are you working on?",
				"Created session: feature/fix-authentication-bug",
				"Branch: feature/fix-authentication-bug",
				"cd ../fix-authentication-bug",
			},
			notContains: []string{"Failed"},
		},
		{
			name:        "description with special chars",
			input:       "Add feature #123 & improvements!\n",
			expectError: false,
			contains:    []string{
				"Created session: feature/add-feature-123-improvements",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing worktrees first
			exec.Command("git", "worktree", "prune").Run()
			
			output := captureOutput(func() {
				mockStdin(tt.input, func() {
					cmd := &cobra.Command{}
					createSession(cmd, []string{})
				})
			})

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
			
			for _, notExpected := range tt.notContains {
				assert.NotContains(t, output, notExpected)
			}

			// Cleanup created worktree if successful
			if !tt.expectError && strings.Contains(output, "Created session") {
				branchName := extractBranchName(output)
				exec.Command("git", "worktree", "remove", "../"+slugify(strings.TrimPrefix(tt.input, "\n"))).Run()
				exec.Command("git", "branch", "-D", branchName).Run()
			}
		})
	}
}

func TestListSessionsCommand(t *testing.T) {
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)
	
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	t.Run("no active sessions", func(t *testing.T) {
		output := captureOutput(func() {
			cmd := &cobra.Command{}
			listSessions(cmd, []string{})
		})
		
		assert.Contains(t, output, "No active sessions")
	})

	t.Run("with active sessions", func(t *testing.T) {
		// Create a test worktree
		cmd := exec.Command("git", "checkout", "-b", "feature/test-list")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "worktree", "add", "../test-list", "feature/test-list")
		cmd.Dir = testRepo
		err := cmd.Run()
		if err != nil {
			// Branch might already exist, try with a unique name
			exec.Command("git", "branch", "-D", "feature/test-list").Run()
			exec.Command("git", "checkout", "-b", "feature/test-list-unique").Run()
			exec.Command("git", "worktree", "add", "../test-list-unique", "feature/test-list-unique").Run()
			defer func() {
				exec.Command("git", "worktree", "remove", "--force", "../test-list-unique").Run()
				exec.Command("git", "branch", "-D", "feature/test-list-unique").Run()
			}()
		} else {
			defer func() {
				exec.Command("git", "worktree", "remove", "--force", "../test-list").Run()
				exec.Command("git", "branch", "-D", "feature/test-list").Run()
			}()
		}

		// List sessions would normally open TUI, but we can test getActiveSessions
		sessions := getActiveSessions()
		assert.NotEmpty(t, sessions)
		
		found := false
		for _, s := range sessions {
			if s.sessionName == "test-list" || s.sessionName == "test-list-unique" {
				found = true
				assert.Contains(t, s.branch, "test-list")
			}
		}
		assert.True(t, found, "Expected to find test-list session")
	})
}

func TestCleanupSessionCommand(t *testing.T) {
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)
	
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	t.Run("no sessions to cleanup", func(t *testing.T) {
		output := captureOutput(func() {
			cmd := &cobra.Command{}
			cleanupSession(cmd, []string{})
		})
		
		assert.Contains(t, output, "No active sessions to cleanup")
	})

	t.Run("cleanup specific session with branch deletion", func(t *testing.T) {
		// Create a test worktree
		cmd := exec.Command("git", "checkout", "-b", "feature/test-cleanup")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "worktree", "add", "../test-cleanup", "feature/test-cleanup")
		cmd.Dir = testRepo
		cmd.Run()
		
		// Go back to main branch
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = testRepo
		cmd.Run()

		output := captureOutput(func() {
			mockStdin("y\n", func() {
				cmd := &cobra.Command{}
				cleanupSession(cmd, []string{"test-cleanup"})
			})
		})
		
		assert.Contains(t, output, "Delete branch feature/test-cleanup?")
		assert.Contains(t, output, "Removed session and branch: test-cleanup")
		
		// Verify worktree and branch are gone
		worktrees, _ := exec.Command("git", "worktree", "list").Output()
		assert.NotContains(t, string(worktrees), "test-cleanup")
		
		branches, _ := exec.Command("git", "branch").Output()
		assert.NotContains(t, string(branches), "feature/test-cleanup")
	})

	t.Run("cleanup session without branch deletion", func(t *testing.T) {
		// Create a test worktree
		cmd := exec.Command("git", "checkout", "-b", "feature/test-cleanup-no-delete")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "worktree", "add", "../test-cleanup-no-delete", "feature/test-cleanup-no-delete")
		cmd.Dir = testRepo
		cmd.Run()
		
		// Go back to main branch
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = testRepo
		cmd.Run()

		output := captureOutput(func() {
			mockStdin("n\n", func() {
				cmd := &cobra.Command{}
				cleanupSession(cmd, []string{"test-cleanup-no-delete"})
			})
		})
		
		assert.Contains(t, output, "Delete branch feature/test-cleanup-no-delete?")
		assert.Contains(t, output, "Removed session: test-cleanup-no-delete")
		
		// Verify worktree is gone but branch remains
		worktrees, _ := exec.Command("git", "worktree", "list").Output()
		assert.NotContains(t, string(worktrees), "test-cleanup-no-delete")
		
		branches, _ := exec.Command("git", "branch").Output()
		assert.Contains(t, string(branches), "feature/test-cleanup-no-delete")
		
		// Cleanup
		exec.Command("git", "branch", "-D", "feature/test-cleanup-no-delete").Run()
	})

	t.Run("cleanup non-existent session", func(t *testing.T) {
		// Create a dummy worktree so we have something
		cmd := exec.Command("git", "checkout", "-b", "feature/dummy")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "worktree", "add", "../dummy", "feature/dummy")
		cmd.Dir = testRepo
		cmd.Run()
		
		defer func() {
			cmd = exec.Command("git", "worktree", "remove", "--force", "../dummy")
			cmd.Dir = testRepo
			cmd.Run()
			
			cmd = exec.Command("git", "branch", "-D", "feature/dummy")
			cmd.Dir = testRepo
			cmd.Run()
		}()

		output := captureOutput(func() {
			cmd := &cobra.Command{}
			cleanupSession(cmd, []string{"non-existent"})
		})
		
		assert.Contains(t, output, "Session not found: non-existent")
	})
}

func TestGetActiveSessionsFiltering(t *testing.T) {
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)
	
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	// The main worktree should not be included
	sessions := getActiveSessions()
	for _, session := range sessions {
		assert.NotEqual(t, filepath.Base(testRepo), session.sessionName)
	}

	// Create additional worktrees
	cmd := exec.Command("git", "checkout", "-b", "feature/test1")
	cmd.Dir = testRepo
	cmd.Run()
	
	cmd = exec.Command("git", "worktree", "add", "../test1", "feature/test1")
	cmd.Dir = testRepo
	cmd.Run()
	
	cmd = exec.Command("git", "checkout", "-b", "feature/test2")
	cmd.Dir = testRepo
	cmd.Run()
	
	cmd = exec.Command("git", "worktree", "add", "../test2", "feature/test2")
	cmd.Dir = testRepo
	cmd.Run()
	
	defer func() {
		cmd = exec.Command("git", "worktree", "remove", "--force", "../test1")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "worktree", "remove", "--force", "../test2")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "branch", "-D", "feature/test1")
		cmd.Dir = testRepo
		cmd.Run()
		
		cmd = exec.Command("git", "branch", "-D", "feature/test2")
		cmd.Dir = testRepo
		cmd.Run()
	}()

	sessions = getActiveSessions()
	assert.Len(t, sessions, 2)
	
	sessionNames := make(map[string]bool)
	for _, s := range sessions {
		sessionNames[s.sessionName] = true
	}
	
	assert.True(t, sessionNames["test1"])
	assert.True(t, sessionNames["test2"])
}

// Helper function to extract branch name from output
func extractBranchName(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Branch:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}

func TestCommandErrors(t *testing.T) {
	t.Run("runCmd with failing command", func(t *testing.T) {
		err := runCmd("git", "invalid-command")
		assert.Error(t, err)
	})

	t.Run("create session in non-git directory", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "non-git-*")
		defer os.RemoveAll(tmpDir)
		
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		output := captureOutput(func() {
			mockStdin("test description\n", func() {
				cmd := &cobra.Command{}
				createSession(cmd, []string{})
			})
		})
		
		assert.Contains(t, output, "Failed to create branch")
	})
}

func TestWorktreeRemovalError(t *testing.T) {
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)
	
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	// Create a worktree
	exec.Command("git", "checkout", "-b", "feature/test-removal").Run()
	exec.Command("git", "worktree", "add", "../test-removal", "feature/test-removal").Run()
	
	// Create a file in the worktree to simulate it being in use
	worktreePath := filepath.Join("..", "test-removal")
	testFile := filepath.Join(worktreePath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	
	// Try to cleanup - this might fail if the directory is locked
	output := captureOutput(func() {
		cmd := &cobra.Command{}
		cleanupSession(cmd, []string{"test-removal"})
	})
	
	// The output will depend on whether the removal succeeds or fails
	// Just ensure the command runs without panic
	assert.NotEmpty(t, output)
	
	// Force cleanup
	exec.Command("git", "worktree", "remove", "--force", "../test-removal").Run()
	exec.Command("git", "branch", "-D", "feature/test-removal").Run()
}

// Test edge case where git worktree list returns empty
func TestEmptyWorktreeList(t *testing.T) {
	// Mock scenario where parseWorktrees gets empty string
	worktrees := parseWorktrees("")
	assert.Empty(t, worktrees)
	
	worktrees = parseWorktrees("   \n  \n  ")
	assert.Empty(t, worktrees)
}