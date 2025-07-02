# Suggested Improvements for ccswitch

## 1. 🏗️ Code Structure & Modularity

### Current Issues:
- All code in a single file (367 lines)
- Mixed concerns (CLI, Git operations, TUI, business logic)
- Hard to test individual components

### Suggested Structure:
```
ccswitch/
├── cmd/
│   ├── root.go         # Root command setup
│   ├── create.go       # Create session command
│   ├── list.go         # List sessions command
│   └── cleanup.go      # Cleanup command
├── internal/
│   ├── git/
│   │   ├── worktree.go # Git worktree operations
│   │   └── branch.go   # Branch management
│   ├── session/
│   │   ├── manager.go  # Session management logic
│   │   └── types.go    # Session types
│   ├── ui/
│   │   ├── models.go   # TUI models
│   │   └── styles.go   # Lipgloss styles
│   └── utils/
│       └── slugify.go  # Utility functions
└── main.go
```

## 2. 🔧 Configuration & Customization

### Add a config file (~/.ccswitch/config.yaml):
```yaml
# Branch prefix customization
branch:
  prefix: "feature/"  # Allow users to change from "feature/"
  
# Worktree location
worktree:
  relative_path: "../"  # Allow customization of where worktrees go
  
# UI preferences
ui:
  show_emoji: true
  color_scheme: "default"
  
# Git settings
git:
  default_branch: "main"  # Support main/master/develop
  auto_fetch: true        # Fetch before creating branches
```

### Implementation:
```go
type Config struct {
    Branch struct {
        Prefix string `yaml:"prefix"`
    } `yaml:"branch"`
    Worktree struct {
        RelativePath string `yaml:"relative_path"`
    } `yaml:"worktree"`
    // ...
}
```

## 3. 🚀 Enhanced Features

### A. Session Templates
```bash
ccswitch --template bugfix
# Creates: bugfix/description instead of feature/description

ccswitch --template experiment
# Creates: experiment/description with special .gitignore
```

### B. Session Metadata
Store session info in `.git/ccswitch-sessions.json`:
```json
{
  "sessions": [{
    "name": "fix-auth-bug",
    "branch": "feature/fix-auth-bug",
    "created": "2024-01-01T10:00:00Z",
    "last_accessed": "2024-01-02T15:30:00Z",
    "description": "Fix authentication bug in login flow"
  }]
}
```

### C. Better Status Command
```bash
ccswitch status
# Shows:
# - Current session
# - Uncommitted changes
# - Time since creation
# - Related PRs (if integrated with GitHub CLI)
```

## 4. 🛡️ Error Handling & Recovery

### Implement a recovery mechanism:
```go
type SessionManager struct {
    // Add transaction-like behavior
}

func (sm *SessionManager) CreateSession(desc string) error {
    tx := sm.BeginTransaction()
    defer tx.Rollback() // Cleanup on error
    
    if err := tx.CreateBranch(); err != nil {
        return err
    }
    if err := tx.CreateWorktree(); err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## 5. 🧪 Improved Testing

### Add integration test helpers:
```go
// test/helpers.go
func NewTestRepo(t *testing.T) *TestRepo {
    // Creates isolated git repo for testing
}

func (tr *TestRepo) Cleanup() {
    // Ensures cleanup even if test fails
}
```

## 6. 📊 Usage Analytics (Optional)

### Track anonymous usage stats:
- Most used commands
- Average session lifetime
- Common error scenarios

## 7. 🔌 Plugin System

### Allow extensions:
```go
type Plugin interface {
    Name() string
    OnSessionCreate(session *Session) error
    OnSessionCleanup(session *Session) error
}

// Example: Jira integration
type JiraPlugin struct{}

func (j *JiraPlugin) OnSessionCreate(s *Session) error {
    // Create Jira ticket based on session name
    return nil
}
```

## 8. 🎯 Command Improvements

### A. Batch Operations
```bash
ccswitch cleanup --all --older-than 30d
# Cleanup all sessions older than 30 days

ccswitch list --format json
# Output in JSON for scripting
```

### B. Session Switching
```bash
ccswitch switch fix-auth-bug
# Switch to existing session without using 'cd'
```

### C. Session Cloning
```bash
ccswitch clone existing-session new-session
# Create new session based on existing one
```

## 9. 🔍 Better Git Integration

### Check for remote branches:
```go
func (gm *GitManager) CheckRemoteBranch(branch string) (bool, error) {
    // Warn if branch exists on remote
}
```

### Auto-sync capability:
```go
func (gm *GitManager) SyncSession(session *Session) error {
    // Pull latest changes
    // Show conflicts if any
}
```

## 10. 📝 Documentation Generation

### Auto-generate docs from sessions:
```bash
ccswitch docs
# Generates SESSIONS.md with all session history
```

## Implementation Priority

1. **High Priority** (Better UX):
   - Code modularization
   - Config file support
   - Better error messages
   - Session status command

2. **Medium Priority** (Enhanced features):
   - Session templates
   - Metadata storage
   - Batch operations
   - Remote branch checking

3. **Low Priority** (Nice to have):
   - Plugin system
   - Usage analytics
   - Documentation generation

## Example Refactored Code Structure

### internal/git/worktree.go
```go
package git

type WorktreeManager struct {
    repoPath string
}

func (wm *WorktreeManager) Create(path, branch string) error {
    // Isolated worktree creation logic
}

func (wm *WorktreeManager) List() ([]Worktree, error) {
    // List worktrees
}

func (wm *WorktreeManager) Remove(path string) error {
    // Remove worktree with validation
}
```

### internal/session/manager.go
```go
package session

type Manager struct {
    git    *git.WorktreeManager
    config *Config
}

func (m *Manager) Create(description string) (*Session, error) {
    // High-level session creation
    // Handles branch naming, validation, etc.
}
```

This modular approach would make the code:
- Easier to test
- More maintainable
- Extensible for new features
- Better separation of concerns