package types

import "fmt"

// Config represents the gman configuration
type Config struct {
	Repositories   map[string]string `yaml:"repositories"`
	CommandAliases map[string]string `yaml:"command_aliases,omitempty"`
	Settings       Settings          `yaml:"settings,omitempty"`
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
	Ahead  int
	Behind int
}

func (s SyncStatus) String() string {
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
	Alias      string
	Path       string
	Branch     string
	IsCurrent  bool
	Workspace  WorkspaceStatus
	SyncStatus SyncStatus
	LastCommit string
	Error      error
}