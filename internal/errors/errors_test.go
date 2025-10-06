package errors

import (
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
		want    string
	}{
		{
			name:    "wrap with message",
			err:     errors.New("original error"),
			message: "additional context",
			want:    "additional context: original error",
		},
		{
			name:    "wrap nil error",
			err:     nil,
			message: "should return nil",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.err, tt.message)
			if tt.err == nil {
				if result != nil {
					t.Errorf("Wrap(nil, %q) = %v, expected nil", tt.message, result)
				}
			} else {
				if result.Error() != tt.want {
					t.Errorf("Wrap(%v, %q) = %q, expected %q", tt.err, tt.message, result.Error(), tt.want)
				}
			}
		})
	}
}

func TestErrorCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{"IsUncommittedChanges true", ErrUncommittedChanges, IsUncommittedChanges, true},
		{"IsUncommittedChanges false", ErrBranchExists, IsUncommittedChanges, false},
		{"IsUncommittedChanges wrapped", Wrap(ErrUncommittedChanges, "context"), IsUncommittedChanges, true},

		{"IsBranchExists true", ErrBranchExists, IsBranchExists, true},
		{"IsBranchExists false", ErrWorktreeExists, IsBranchExists, false},

		{"IsWorktreeExists true", ErrWorktreeExists, IsWorktreeExists, true},
		{"IsWorktreeExists false", ErrSessionNotFound, IsWorktreeExists, false},

		{"IsAlreadyOnBranch true", ErrAlreadyOnBranch, IsAlreadyOnBranch, true},
		{"IsAlreadyOnBranch false", ErrNoSessions, IsAlreadyOnBranch, false},

		{"IsSessionNotFound true", ErrSessionNotFound, IsSessionNotFound, true},
		{"IsSessionNotFound false", ErrBranchNotFound, IsSessionNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			if result != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestErrorHint(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "uncommitted changes hint",
			err:  ErrUncommittedChanges,
			want: "Use 'git stash' to temporarily save changes",
		},
		{
			name: "branch exists hint",
			err:  ErrBranchExists,
			want: "Use 'git branch -D <branch>' to delete it first",
		},
		{
			name: "worktree exists hint",
			err:  ErrWorktreeExists,
			want: "Use a different description or remove the existing directory",
		},
		{
			name: "already on branch hint",
			err:  ErrAlreadyOnBranch,
			want: "Switch to main/master branch first, or use a different description",
		},
		{
			name: "session not found hint",
			err:  ErrSessionNotFound,
			want: "Use 'ccswitch list' to see available sessions",
		},
		{
			name: "unknown error no hint",
			err:  errors.New("unknown error"),
			want: "",
		},
		{
			name: "wrapped error preserves hint",
			err:  Wrap(ErrUncommittedChanges, "context"),
			want: "Use 'git stash' to temporarily save changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorHint(tt.err)
			if result != tt.want {
				t.Errorf("ErrorHint(%v) = %q, expected %q", tt.err, result, tt.want)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Ensure all error constants are defined and unique
	errors := []error{
		ErrUncommittedChanges,
		ErrBranchExists,
		ErrBranchNotFound,
		ErrWorktreeExists,
		ErrWorktreeNotFound,
		ErrSessionNotFound,
		ErrAlreadyOnBranch,
		ErrNoSessions,
	}

	seen := make(map[string]bool)
	for _, err := range errors {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("Duplicate error message: %q", msg)
		}
		seen[msg] = true

		if msg == "" {
			t.Error("Error constant has empty message")
		}
	}
}
