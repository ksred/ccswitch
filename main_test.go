package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommand is used to mock exec.Command calls
type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCommand) Output() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// Helper to create a temporary git repository for testing
func createTestRepo(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "ccsplit-test-*")
	assert.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	assert.NoError(t, err)

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	err = cmd.Run()
	assert.NoError(t, err)

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(testFile, []byte("# Test Repo"), 0644)
	assert.NoError(t, err)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	err = cmd.Run()
	assert.NoError(t, err)

	return tmpDir
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "special characters",
			input:    "Fix bug #123 & test @feature!",
			expected: "fix-bug-123-test-feature",
		},
		{
			name:     "multiple spaces",
			input:    "Too   many    spaces",
			expected: "too-many-spaces",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  trim me  ",
			expected: "trim-me",
		},
		{
			name:     "already slugified",
			input:    "already-slugified",
			expected: "already-slugified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseWorktrees(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []worktree
	}{
		{
			name: "single worktree",
			input: `worktree /path/to/repo
HEAD abcdef123456
branch refs/heads/main

`,
			expected: []worktree{
				{
					path:   "/path/to/repo",
					branch: "main",
				},
			},
		},
		{
			name: "multiple worktrees",
			input: `worktree /path/to/repo
HEAD abcdef123456
branch refs/heads/main

worktree /path/to/feature-branch
HEAD fedcba654321
branch refs/heads/feature/new-feature

`,
			expected: []worktree{
				{
					path:   "/path/to/repo",
					branch: "main",
				},
				{
					path:   "/path/to/feature-branch",
					branch: "feature/new-feature",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWorktrees(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionItem(t *testing.T) {
	item := sessionItem{
		session: worktree{
			path:        "/path/to/feature",
			branch:      "feature/test-feature",
			sessionName: "test-feature",
		},
	}

	assert.Equal(t, "test-feature", item.FilterValue())
	assert.Equal(t, "test-feature", item.Title())
	assert.Equal(t, "feature/test-feature → /path/to/feature", item.Description())
}

func TestSessionDelegate(t *testing.T) {
	delegate := sessionDelegate{}
	
	assert.Equal(t, 1, delegate.Height())
	assert.Equal(t, 0, delegate.Spacing())
	assert.Nil(t, delegate.Update(nil, nil))
}

func TestListModel(t *testing.T) {
	items := []list.Item{
		sessionItem{session: worktree{sessionName: "test1", branch: "feature/test1"}},
		sessionItem{session: worktree{sessionName: "test2", branch: "feature/test2"}},
	}
	
	l := list.New(items, sessionDelegate{}, 50, 10)
	model := listModel{list: l}

	// Test Init
	assert.Nil(t, model.Init())

	// Test quit commands
	quitTests := []string{"q", "ctrl+c"}
	for _, key := range quitTests {
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		assert.NotNil(t, cmd)
		_, ok := updatedModel.(listModel)
		assert.True(t, ok)
	}

	// Test View
	view := model.View()
	assert.True(t, strings.HasPrefix(view, "\n"))
}

func TestCleanupModel(t *testing.T) {
	items := []list.Item{
		sessionItem{session: worktree{sessionName: "test1", branch: "feature/test1"}},
		sessionItem{session: worktree{sessionName: "test2", branch: "feature/test2"}},
	}
	
	l := list.New(items, sessionDelegate{}, 50, 10)
	model := cleanupModel{list: l}

	// Test Init
	assert.Nil(t, model.Init())

	// Test quit commands
	quitTests := []string{"q", "ctrl+c"}
	for _, key := range quitTests {
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		assert.NotNil(t, cmd)
		cleanupModel, ok := updatedModel.(cleanupModel)
		assert.True(t, ok)
		assert.Equal(t, "", cleanupModel.selected)
	}

	// Test View
	view := model.View()
	assert.True(t, strings.HasPrefix(view, "\n"))
}

func TestCreateSessionWithEmptyDescription(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simulate empty input
	oldStdin := os.Stdin
	pr, pw, _ := os.Pipe()
	os.Stdin = pr
	go func() {
		pw.Write([]byte("\n"))
		pw.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	createSession(nil, nil)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	output := string(out)

	assert.Contains(t, output, "What are you working on?")
	assert.Contains(t, output, "Description cannot be empty")
}

func TestRunCmd(t *testing.T) {
	// Test successful command
	err := runCmd("echo", "test")
	assert.NoError(t, err)

	// Test failing command
	err = runCmd("false")
	assert.Error(t, err)

	// Test non-existent command
	err = runCmd("this-command-does-not-exist")
	assert.Error(t, err)
}

func TestGetCurrentDir(t *testing.T) {
	dir := getCurrentDir()
	assert.NotEmpty(t, dir)
	
	// Should return a valid directory path
	_, err := os.Stat(dir)
	assert.NoError(t, err)
}

func TestSessionDelegateRender(t *testing.T) {
	delegate := sessionDelegate{}
	
	// Create a test model with items
	items := []list.Item{
		sessionItem{session: worktree{sessionName: "test1", branch: "feature/test1"}},
		sessionItem{session: worktree{sessionName: "test2", branch: "feature/test2"}},
	}
	l := list.New(items, delegate, 50, 10)
	
	// Test rendering selected item
	var buf strings.Builder
	delegate.Render(&buf, l, 0, items[0])
	output := buf.String()
	assert.Contains(t, output, "test1")
	assert.Contains(t, output, "feature/test1")
	
	// Test rendering with invalid item type by checking cast failure
	// The render function checks if the item can be cast to sessionItem
	// and returns early if not, so we just need to ensure it handles this gracefully
}

func TestWorktreeParsingEdgeCases(t *testing.T) {
	// Test worktree without branch
	input := `worktree /path/to/repo
HEAD abcdef123456

`
	result := parseWorktrees(input)
	assert.Len(t, result, 1)
	assert.Equal(t, "/path/to/repo", result[0].path)
	assert.Empty(t, result[0].branch)
	
	// Test malformed input
	input = `malformed input
not a valid worktree format`
	result = parseWorktrees(input)
	assert.Empty(t, result)
}

// Integration test for the full workflow
func TestIntegrationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test repository
	testRepo := createTestRepo(t)
	defer os.RemoveAll(testRepo)

	// Change to the test repo directory
	originalDir, _ := os.Getwd()
	os.Chdir(testRepo)
	defer os.Chdir(originalDir)

	// Test creating a session (would need to mock stdin for full test)
	// This is a placeholder for the integration test structure
	t.Run("full workflow", func(t *testing.T) {
		// 1. Create a worktree manually to test listing
		cmd := exec.Command("git", "checkout", "-b", "feature/test-integration")
		cmd.Dir = testRepo
		err := cmd.Run()
		assert.NoError(t, err)
		
		cmd = exec.Command("git", "worktree", "add", "../test-integration", "feature/test-integration")
		cmd.Dir = testRepo
		err = cmd.Run()
		assert.NoError(t, err)
		
		// 2. Test getActiveSessions
		sessions := getActiveSessions()
		assert.NotEmpty(t, sessions)
		
		// 3. Cleanup
		cmd = exec.Command("git", "worktree", "remove", "../test-integration")
		cmd.Dir = testRepo
		err = cmd.Run()
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkSlugify(b *testing.B) {
	testStrings := []string{
		"Simple test",
		"Complex @#$% test with special chars!!!",
		"Very long description that needs to be slugified into a branch name",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			slugify(s)
		}
	}
}

func BenchmarkParseWorktrees(b *testing.B) {
	input := `worktree /path/to/repo
HEAD abcdef123456
branch refs/heads/main

worktree /path/to/feature1
HEAD fedcba654321
branch refs/heads/feature/feature1

worktree /path/to/feature2
HEAD 123456abcdef
branch refs/heads/feature/feature2
`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseWorktrees(input)
	}
}