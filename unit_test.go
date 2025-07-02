package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests that don't require git operations

func TestSlugifyFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase conversion", "UPPERCASE", "uppercase"},
		{"space to dash", "hello world", "hello-world"},
		{"special chars removal", "test@#$%^&*()", "test"},
		{"multiple spaces", "too   many    spaces", "too-many-spaces"},
		{"trim dashes", "--trimmed--", "trimmed"},
		{"unicode handling", "café résumé", "caf-rsum"},
		{"numbers preserved", "test123", "test123"},
		{"dash preserved", "already-dashed", "already-dashed"},
		{"underscore preserved", "test_underscore", "test_underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseWorktreesFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []worktree
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \n  \t  ",
			expected: nil,
		},
		{
			name: "single worktree with branch",
			input: `worktree /home/user/project
HEAD abc123def
branch refs/heads/main
`,
			expected: []worktree{
				{path: "/home/user/project", branch: "main"},
			},
		},
		{
			name: "single worktree without branch",
			input: `worktree /home/user/project
HEAD abc123def
`,
			expected: []worktree{
				{path: "/home/user/project", branch: ""},
			},
		},
		{
			name: "multiple worktrees",
			input: `worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/feature1
HEAD def456
branch refs/heads/feature/new-feature

worktree /home/user/feature2
HEAD ghi789
branch refs/heads/bugfix/fix-issue
`,
			expected: []worktree{
				{path: "/home/user/project", branch: "main"},
				{path: "/home/user/feature1", branch: "feature/new-feature"},
				{path: "/home/user/feature2", branch: "bugfix/fix-issue"},
			},
		},
		{
			name: "worktree with bare repository",
			input: `worktree /home/user/project
HEAD abc123
branch refs/heads/main
bare

worktree /home/user/feature
HEAD def456
branch refs/heads/feature/test
`,
			expected: []worktree{
				{path: "/home/user/project", branch: "main"},
				{path: "/home/user/feature", branch: "feature/test"},
			},
		},
		{
			name: "malformed input",
			input: `this is not valid worktree output
just some random text
`,
			expected: nil,
		},
		{
			name: "worktree with detached HEAD",
			input: `worktree /home/user/project
HEAD abc123def
detached
`,
			expected: []worktree{
				{path: "/home/user/project", branch: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseWorktrees(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionItemInterface(t *testing.T) {
	item := sessionItem{
		session: worktree{
			path:        "/home/user/my-feature",
			branch:      "feature/awesome-feature",
			sessionName: "my-feature",
		},
	}

	t.Run("FilterValue", func(t *testing.T) {
		assert.Equal(t, "my-feature", item.FilterValue())
	})

	t.Run("Title", func(t *testing.T) {
		assert.Equal(t, "my-feature", item.Title())
	})

	t.Run("Description", func(t *testing.T) {
		expected := "feature/awesome-feature → /home/user/my-feature"
		assert.Equal(t, expected, item.Description())
	})
}

func TestGetCurrentDirFunction(t *testing.T) {
	dir := getCurrentDir()
	assert.NotEmpty(t, dir, "getCurrentDir should return a non-empty string")
	assert.DirExists(t, dir, "getCurrentDir should return an existing directory")
}

func TestRunCmdFunction(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command",
			cmd:     "true",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "failing command",
			cmd:     "false",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "command with args",
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "non-existent command",
			cmd:     "this-command-definitely-does-not-exist",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCmd(tt.cmd, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorktreeTypeFields(t *testing.T) {
	wt := worktree{
		path:        "/test/path",
		branch:      "test-branch",
		sessionName: "test-session",
	}

	assert.Equal(t, "/test/path", wt.path)
	assert.Equal(t, "test-branch", wt.branch)
	assert.Equal(t, "test-session", wt.sessionName)
}

// Benchmarks
func BenchmarkSlugifyShort(b *testing.B) {
	input := "Hello World"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slugify(input)
	}
}

func BenchmarkSlugifyLong(b *testing.B) {
	input := "This is a very long string with !@#$%^&*() special characters and     multiple     spaces"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slugify(input)
	}
}

func BenchmarkParseWorktreesSmall(b *testing.B) {
	input := `worktree /path/to/repo
HEAD abc123
branch refs/heads/main
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseWorktrees(input)
	}
}

func BenchmarkParseWorktreesLarge(b *testing.B) {
	input := `worktree /path/to/repo1
HEAD abc123
branch refs/heads/main

worktree /path/to/repo2
HEAD def456
branch refs/heads/feature/test

worktree /path/to/repo3
HEAD ghi789
branch refs/heads/bugfix/issue-123

worktree /path/to/repo4
HEAD jkl012
branch refs/heads/feature/new-feature

worktree /path/to/repo5
HEAD mno345
branch refs/heads/develop
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseWorktrees(input)
	}
}