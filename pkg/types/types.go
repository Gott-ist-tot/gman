package types

import (
	"fmt"
	"time"
)

// Config represents the gman configuration
type Config struct {
	Repositories   map[string]string `yaml:"repositories"`
	CommandAliases map[string]string `yaml:"command_aliases,omitempty"`
	Settings       Settings          `yaml:"settings,omitempty"`
	RecentUsage    []RecentEntry     `yaml:"recent_usage,omitempty"`
	Groups         map[string]Group  `yaml:"groups,omitempty"`
}

// Settings contains user preferences
type Settings struct {
	DefaultSyncMode string `yaml:"default_sync_mode,omitempty"`
	ShowLastCommit  bool   `yaml:"show_last_commit"`
	ParallelJobs    int    `yaml:"parallel_jobs"`
}

// WorkspaceStatus represents the status of a git workspace
type WorkspaceStatus int

const (
	Clean WorkspaceStatus = iota
	Dirty
	Stashed
)

func (w WorkspaceStatus) String() string {
	switch w {
	case Clean:
		return "🟢 CLEAN"
	case Dirty:
		return "🔴 DIRTY"
	case Stashed:
		return "🟡 STASHED"
	default:
		return "❓ UNKNOWN"
	}
}

// SyncStatus represents the sync status with remote
type SyncStatus struct {
	Ahead     int
	Behind    int
	SyncError error // Error during sync operation (e.g., fetch failure)
}

func (s SyncStatus) String() string {
	// Check for sync errors first
	if s.SyncError != nil {
		return "❌ SYNC FAILED"
	}
	
	if s.Ahead > 0 && s.Behind > 0 {
		return fmt.Sprintf("🔄 %d↑ %d↓", s.Ahead, s.Behind)
	} else if s.Ahead > 0 {
		return fmt.Sprintf("↑ %d AHEAD", s.Ahead)
	} else if s.Behind > 0 {
		return fmt.Sprintf("↓ %d BEHIND", s.Behind)
	}
	return "✅ UP-TO-DATE"
}

// RepoStatus represents the status of a single repository
type RepoStatus struct {
	Alias         string
	Path          string
	Branch        string
	IsCurrent     bool
	Workspace     WorkspaceStatus
	SyncStatus    SyncStatus
	LastCommit    string
	FilesChanged  int           // Number of changed files
	CommitTime    time.Time     // Time of last commit
	Error         error
	
	// Enhanced status information
	RemoteURL       string        // URL of the remote origin
	RemoteBranch    string        // Name of the tracking remote branch
	StashCount      int           // Number of stashes
	TotalBranches   int           // Total number of branches (local + remote)
	LocalBranches   int           // Number of local branches
	RemoteBranches  int           // Number of remote branches
	LastFetchTime   time.Time     // Time of last fetch operation
}

// RecentEntry represents a recently used repository
type RecentEntry struct {
	Alias     string    `yaml:"alias"`
	AccessTime time.Time `yaml:"access_time"`
}

// Group represents a collection of repositories
type Group struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	Repositories []string `yaml:"repositories"`
	CreatedAt    time.Time `yaml:"created_at"`
}

// Worktree represents a Git worktree
type Worktree struct {
	Path       string `json:"path"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	IsBare     bool   `json:"is_bare"`
	IsDetached bool   `json:"is_detached"`
}

// SwitchTarget represents a target that can be switched to (repository or worktree)
type SwitchTarget struct {
	Alias       string `json:"alias"`        // Display name for the target
	Path        string `json:"path"`         // Actual filesystem path
	Type        string `json:"type"`         // "repository" or "worktree"
	RepoAlias   string `json:"repo_alias"`   // Parent repository alias (for worktrees)
	Branch      string `json:"branch,omitempty"` // Current branch (for worktrees)
	Description string `json:"description,omitempty"` // Additional info
}
