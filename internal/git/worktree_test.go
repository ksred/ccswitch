package git

import (
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