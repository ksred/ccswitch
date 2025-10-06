package git

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// SessionInfo represents information about a ccswitch session
type SessionInfo struct {
	Name   string
	Branch string
	Path   string
}
