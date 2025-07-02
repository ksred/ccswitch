package errors

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrUncommittedChanges = errors.New("uncommitted changes")
	ErrBranchExists       = errors.New("branch already exists")
	ErrBranchNotFound     = errors.New("branch not found")
	ErrWorktreeExists     = errors.New("worktree already exists")
	ErrWorktreeNotFound   = errors.New("worktree not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrAlreadyOnBranch    = errors.New("already on branch")
	ErrNoSessions         = errors.New("no active sessions")
)

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsUncommittedChanges checks if the error is due to uncommitted changes
func IsUncommittedChanges(err error) bool {
	return errors.Is(err, ErrUncommittedChanges)
}

// IsBranchExists checks if the error is due to branch already existing
func IsBranchExists(err error) bool {
	return errors.Is(err, ErrBranchExists)
}

// IsWorktreeExists checks if the error is due to worktree already existing
func IsWorktreeExists(err error) bool {
	return errors.Is(err, ErrWorktreeExists)
}

// IsAlreadyOnBranch checks if the error is due to already being on the branch
func IsAlreadyOnBranch(err error) bool {
	return errors.Is(err, ErrAlreadyOnBranch)
}

// IsSessionNotFound checks if the error is due to session not found
func IsSessionNotFound(err error) bool {
	return errors.Is(err, ErrSessionNotFound)
}

// ErrorHint provides helpful hints for common errors
func ErrorHint(err error) string {
	switch {
	case IsUncommittedChanges(err):
		return "Use 'git stash' to temporarily save changes"
	case IsBranchExists(err):
		return "Use 'git branch -D <branch>' to delete it first"
	case IsWorktreeExists(err):
		return "Use a different description or remove the existing directory"
	case IsAlreadyOnBranch(err):
		return "Switch to main/master branch first, or use a different description"
	case IsSessionNotFound(err):
		return "Use 'ccswitch list' to see available sessions"
	default:
		return ""
	}
}