package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetRepoName(t *testing.T) {
	// This test requires a git repository, so we'll create a temporary one
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}
	
	// Test GetRepoName
	name, err := GetRepoName(tempDir)
	if err != nil {
		t.Fatalf("GetRepoName() failed: %v", err)
	}
	
	expectedName := filepath.Base(tempDir)
	if name != expectedName {
		t.Errorf("GetRepoName() = %q, expected %q", name, expectedName)
	}
}

func TestGetRepoNameNonGitDir(t *testing.T) {
	// Test with a non-git directory
	tempDir := t.TempDir()
	
	_, err := GetRepoName(tempDir)
	if err == nil {
		t.Error("GetRepoName() should fail for non-git directory")
	}
}

func TestGetMainRepoPath(t *testing.T) {
	// This test requires a git repository
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}
	
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
	
	// Test from main repository
	mainPath, err := GetMainRepoPath(tempDir)
	if err != nil {
		t.Fatalf("GetMainRepoPath() from main repo failed: %v", err)
	}
	
	// The path should be absolute and point to tempDir
	if !filepath.IsAbs(mainPath) {
		t.Errorf("GetMainRepoPath() should return absolute path, got %q", mainPath)
	}
	
	// Clean paths for comparison (macOS sometimes adds /private prefix)
	cleanMainPath := strings.TrimPrefix(mainPath, "/private")
	cleanTempDir := strings.TrimPrefix(tempDir, "/private")
	
	if cleanMainPath != cleanTempDir {
		t.Errorf("GetMainRepoPath() from main = %q, expected %q", mainPath, tempDir)
	}
	
	// Create a worktree
	worktreeDir := filepath.Join(tempDir, "worktree")
	cmd = exec.Command("git", "worktree", "add", worktreeDir)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to create worktree: %v", err)
	}
	
	// Test from worktree
	mainPathFromWorktree, err := GetMainRepoPath(worktreeDir)
	if err != nil {
		t.Fatalf("GetMainRepoPath() from worktree failed: %v", err)
	}
	
	// Clean up the path for comparison
	mainPathFromWorktree = strings.TrimSuffix(mainPathFromWorktree, string(filepath.Separator))
	mainPathFromWorktree = strings.TrimPrefix(mainPathFromWorktree, "/private")
	tempDirClean := strings.TrimSuffix(tempDir, string(filepath.Separator))
	tempDirClean = strings.TrimPrefix(tempDirClean, "/private")
	
	if mainPathFromWorktree != tempDirClean {
		t.Errorf("GetMainRepoPath() from worktree = %q, expected %q", mainPathFromWorktree, tempDirClean)
	}
}

func TestGetMainRepoPathHandlesDotGit(t *testing.T) {
	// This test specifically checks the fix for when git rev-parse --git-common-dir returns ".git"
	tempDir := t.TempDir()
	
	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}
	
	// Verify that git rev-parse --git-common-dir returns ".git" from main repo
	cmd = exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = tempDir
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("Failed to run git rev-parse: %v", err)
	}
	
	gitCommonDir := strings.TrimSpace(string(output))
	if gitCommonDir != ".git" {
		t.Skipf("Expected git rev-parse --git-common-dir to return '.git', got %q", gitCommonDir)
	}
	
	// Now test that GetMainRepoPath handles this correctly
	mainPath, err := GetMainRepoPath(tempDir)
	if err != nil {
		t.Fatalf("GetMainRepoPath() failed: %v", err)
	}
	
	// The path should be absolute, not "."
	if mainPath == "." {
		t.Errorf("GetMainRepoPath() returned '.', should return absolute path")
	}
	
	if !filepath.IsAbs(mainPath) {
		t.Errorf("GetMainRepoPath() should return absolute path, got %q", mainPath)
	}
	
	// Clean paths for comparison
	cleanMainPath := strings.TrimPrefix(mainPath, "/private")
	cleanTempDir := strings.TrimPrefix(tempDir, "/private")
	
	if cleanMainPath != cleanTempDir {
		t.Errorf("GetMainRepoPath() = %q, expected %q", mainPath, tempDir)
	}
}

func TestGetMainRepoPathNonGitDir(t *testing.T) {
	// Test with a non-git directory
	tempDir := t.TempDir()
	
	_, err := GetMainRepoPath(tempDir)
	if err == nil {
		t.Error("GetMainRepoPath() should fail for non-git directory")
	}
}