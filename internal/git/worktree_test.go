package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWorktrees(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Worktree
	}{
		{
			name: "single worktree with branch",
			input: `worktree /home/user/project
HEAD abc123def
branch refs/heads/main
`,
			expected: []Worktree{
				{Path: "/home/user/project", Branch: "main", Commit: "abc123def"},
			},
		},
		{
			name: "multiple worktrees including main",
			input: `worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/.ccswitch/worktrees/project/feature1
HEAD def456
branch refs/heads/feature/new-feature

worktree /home/user/.ccswitch/worktrees/project/feature2
HEAD ghi789
branch refs/heads/bugfix/fix-issue
`,
			expected: []Worktree{
				{Path: "/home/user/project", Branch: "main", Commit: "abc123"},
				{Path: "/home/user/.ccswitch/worktrees/project/feature1", Branch: "feature/new-feature", Commit: "def456"},
				{Path: "/home/user/.ccswitch/worktrees/project/feature2", Branch: "bugfix/fix-issue", Commit: "ghi789"},
			},
		},
		{
			name: "worktree without branch (detached HEAD)",
			input: `worktree /home/user/project
HEAD abc123def
detached
`,
			expected: []Worktree{
				{Path: "/home/user/project", Branch: "", Commit: "abc123def"},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []Worktree{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWorktrees(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseWorktrees() returned %d worktrees, expected %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i].Path != tt.expected[i].Path {
					t.Errorf("Worktree[%d].Path = %s, expected %s", i, result[i].Path, tt.expected[i].Path)
				}
				if result[i].Branch != tt.expected[i].Branch {
					t.Errorf("Worktree[%d].Branch = %s, expected %s", i, result[i].Branch, tt.expected[i].Branch)
				}
				if result[i].Commit != tt.expected[i].Commit {
					t.Errorf("Worktree[%d].Commit = %s, expected %s", i, result[i].Commit, tt.expected[i].Commit)
				}
			}
		})
	}
}

func TestWorktreeManagerCreate(t *testing.T) {
	// This test requires a git repository
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}
	
	// Configure git user for the test repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	cmd.Run()
	
	// Create an initial commit (required for worktrees)
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to add file: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to commit: %v", err)
	}
	
	// Create a branch for the worktree
	branchName := "feature/test-branch"
	cmd = exec.Command("git", "branch", branchName)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}
	
	// Create WorktreeManager
	wm := NewWorktreeManager(tempDir)
	
	// Test creating a worktree
	worktreePath := filepath.Join(tempDir, "test-worktree")
	err := wm.Create(worktreePath, branchName)
	if err != nil {
		t.Fatalf("WorktreeManager.Create() failed: %v", err)
	}
	
	// Verify the worktree was created
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Errorf("Worktree directory was not created at %s", worktreePath)
	}
	
	// Verify it's a valid git worktree
	cmd = exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("Created worktree is not a valid git directory: %v", err)
	}
	
	toplevel := strings.TrimSpace(string(output))
	// Clean paths for comparison (macOS sometimes adds /private prefix)
	cleanToplevel := strings.TrimPrefix(toplevel, "/private")
	cleanWorktreePath := strings.TrimPrefix(worktreePath, "/private")
	if cleanToplevel != cleanWorktreePath {
		t.Errorf("Worktree toplevel = %q, expected %q", toplevel, worktreePath)
	}
	
	// Verify the branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	output, err = cmd.Output()
	if err != nil {
		t.Errorf("Failed to get current branch: %v", err)
	}
	
	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != branchName {
		t.Errorf("Worktree branch = %q, expected %q", currentBranch, branchName)
	}
}

func TestWorktreeManagerCreateWithMainRepoPath(t *testing.T) {
	// This test verifies that worktrees are created correctly when WorktreeManager
	// is initialized with the main repo path (not ".")
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}
	
	// Configure git user for the test repo
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	cmd.Run()
	
	// Create an initial commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to add file: %v", err)
	}
	
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to commit: %v", err)
	}
	
	// Get the main repo path (should handle the ".git" case correctly)
	mainRepoPath, err := GetMainRepoPath(tempDir)
	if err != nil {
		t.Fatalf("GetMainRepoPath() failed: %v", err)
	}
	
	// Create a branch
	branchName := "feature/path-test"
	cmd = exec.Command("git", "branch", branchName)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}
	
	// Create WorktreeManager with the main repo path
	wm := NewWorktreeManager(mainRepoPath)
	
	// Simulate the path that would be used in real usage
	repoName := filepath.Base(mainRepoPath)
	
	// For testing, use a path within tempDir instead
	testWorktreePath := filepath.Join(tempDir, ".ccswitch", "worktrees", repoName, "test-session")
	
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(testWorktreePath), 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	
	// Create the worktree
	err = wm.Create(testWorktreePath, branchName)
	if err != nil {
		t.Fatalf("WorktreeManager.Create() failed: %v", err)
	}
	
	// Verify the worktree was created in the correct location
	if _, err := os.Stat(testWorktreePath); os.IsNotExist(err) {
		t.Errorf("Worktree was not created at expected path %s", testWorktreePath)
	}
	
	// List worktrees and verify our new one is there
	worktrees, err := wm.List()
	if err != nil {
		t.Fatalf("WorktreeManager.List() failed: %v", err)
	}
	
	found := false
	for _, wt := range worktrees {
		// Clean paths for comparison
		cleanWtPath := strings.TrimPrefix(wt.Path, "/private")
		cleanTestPath := strings.TrimPrefix(testWorktreePath, "/private")
		
		if cleanWtPath == cleanTestPath {
			found = true
			if wt.Branch != branchName {
				t.Errorf("Worktree branch = %q, expected %q", wt.Branch, branchName)
			}
			break
		}
	}
	
	if !found {
		t.Errorf("Created worktree not found in worktree list")
	}
}

func TestGetSessionsFromWorktrees(t *testing.T) {
	tests := []struct {
		name      string
		worktrees []Worktree
		repoName  string
		expected  []SessionInfo
	}{
		{
			name: "includes main repository and ccswitch sessions",
			worktrees: []Worktree{
				{Path: "/home/user/myrepo", Branch: "main", Commit: "abc123"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/feature1", Branch: "feature/test", Commit: "def456"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/bugfix1", Branch: "bugfix/issue", Commit: "ghi789"},
			},
			repoName: "myrepo",
			expected: []SessionInfo{
				{Name: "main", Branch: "main", Path: "/home/user/myrepo"},
				{Name: "feature1", Branch: "feature/test", Path: "/home/user/.ccswitch/worktrees/myrepo/feature1"},
				{Name: "bugfix1", Branch: "bugfix/issue", Path: "/home/user/.ccswitch/worktrees/myrepo/bugfix1"},
			},
		},
		{
			name: "handles multiple non-ccswitch worktrees (picks first as main)",
			worktrees: []Worktree{
				{Path: "/home/user/repo1", Branch: "master", Commit: "abc123"},
				{Path: "/home/user/repo2", Branch: "develop", Commit: "def456"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/feature", Branch: "feature/new", Commit: "ghi789"},
			},
			repoName: "myrepo",
			expected: []SessionInfo{
				{Name: "main", Branch: "master", Path: "/home/user/repo1"},
				{Name: "feature", Branch: "feature/new", Path: "/home/user/.ccswitch/worktrees/myrepo/feature"},
			},
		},
		{
			name: "filters out worktrees without branches",
			worktrees: []Worktree{
				{Path: "/home/user/myrepo", Branch: "main", Commit: "abc123"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/detached", Branch: "", Commit: "def456"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/feature", Branch: "feature/test", Commit: "ghi789"},
			},
			repoName: "myrepo",
			expected: []SessionInfo{
				{Name: "main", Branch: "main", Path: "/home/user/myrepo"},
				{Name: "feature", Branch: "feature/test", Path: "/home/user/.ccswitch/worktrees/myrepo/feature"},
			},
		},
		{
			name: "only ccswitch worktrees (no main repo)",
			worktrees: []Worktree{
				{Path: "/home/user/.ccswitch/worktrees/myrepo/feature1", Branch: "feature/one", Commit: "abc123"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/feature2", Branch: "feature/two", Commit: "def456"},
			},
			repoName: "myrepo",
			expected: []SessionInfo{
				{Name: "feature1", Branch: "feature/one", Path: "/home/user/.ccswitch/worktrees/myrepo/feature1"},
				{Name: "feature2", Branch: "feature/two", Path: "/home/user/.ccswitch/worktrees/myrepo/feature2"},
			},
		},
		{
			name:      "empty worktrees",
			worktrees: []Worktree{},
			repoName:  "myrepo",
			expected:  []SessionInfo{},
		},
		{
			name: "filters worktrees for different repo",
			worktrees: []Worktree{
				{Path: "/home/user/myrepo", Branch: "main", Commit: "abc123"},
				{Path: "/home/user/.ccswitch/worktrees/otherrepo/feature", Branch: "feature/test", Commit: "def456"},
				{Path: "/home/user/.ccswitch/worktrees/myrepo/bugfix", Branch: "bugfix/issue", Commit: "ghi789"},
			},
			repoName: "myrepo",
			expected: []SessionInfo{
				{Name: "main", Branch: "main", Path: "/home/user/myrepo"},
				{Name: "bugfix", Branch: "bugfix/issue", Path: "/home/user/.ccswitch/worktrees/myrepo/bugfix"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSessionsFromWorktrees(tt.worktrees, tt.repoName)
			if len(result) != len(tt.expected) {
				t.Errorf("GetSessionsFromWorktrees() returned %d sessions, expected %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i].Name != tt.expected[i].Name {
					t.Errorf("Session[%d].Name = %s, expected %s", i, result[i].Name, tt.expected[i].Name)
				}
				if result[i].Branch != tt.expected[i].Branch {
					t.Errorf("Session[%d].Branch = %s, expected %s", i, result[i].Branch, tt.expected[i].Branch)
				}
				if result[i].Path != tt.expected[i].Path {
					t.Errorf("Session[%d].Path = %s, expected %s", i, result[i].Path, tt.expected[i].Path)
				}
			}
		})
	}
}