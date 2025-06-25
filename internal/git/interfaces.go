package git

import (
	"gman/pkg/types"
)

// StatusReader provides repository status information
type StatusReader interface {
	// Repository status operations
	GetRepoStatus(alias, path string) types.RepoStatus
	GetRepoStatusNoFetch(alias, path string) types.RepoStatus // Fast status without network operations
	GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error)
	GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) // Fast status for multiple repos
	IsGitRepository(path string) bool

	// Change detection
	HasUncommittedChanges(path string) (bool, error)
	HasUnpushedCommits(path string) (bool, error)
}

// BranchManager handles branch operations
type BranchManager interface {
	// Branch information
	GetCurrentBranch(path string) (string, error)
	GetBranches(path string, includeRemote bool) ([]string, error)

	// Branch operations
	CreateBranch(path, branchName string) error
	SwitchBranch(path, branchName string) error
	CleanMergedBranches(path, mainBranch string) ([]string, error)
}

// SyncManager handles repository synchronization
type SyncManager interface {
	// Sync operations
	SyncRepository(path, mode string) error
	SyncAllRepositories(repositories map[string]string, mode string, maxConcurrency int) error
}

// CommitManager handles commit and push/pull operations
type CommitManager interface {
	// Commit operations
	CommitChanges(path, message string, addAll bool) error
	PushChanges(path string, force, setUpstream bool) error

	// Stash operations
	StashSave(path, message string) error
	StashPop(path string) error
	StashList(path string) ([]string, error)
	StashClear(path string) error
}

// WorktreeManager handles Git worktree operations
type WorktreeManager interface {
	// Worktree operations
	AddWorktree(repoPath, worktreePath, branch string) error
	ListWorktrees(repoPath string) ([]types.Worktree, error)
	RemoveWorktree(repoPath, worktreePath string, force bool) error
}

// DiffProvider handles file comparison operations
type DiffProvider interface {
	// Diff operations
	DiffFileBetweenBranches(repoPath, branch1, branch2, filePath string) (string, error)
	DiffFileBetweenRepos(repo1Path, repo2Path, filePath string) (string, error)
	GetFileContentFromBranch(repoPath, branch, filePath string) (string, error)
}

// CommandExecutor provides low-level Git command execution
type CommandExecutor interface {
	// Command execution
	RunCommand(path string, args ...string) (string, error)
}

// GitOperations combines all Git operation interfaces
type GitOperations interface {
	StatusReader
	BranchManager
	SyncManager
	CommitManager
	WorktreeManager
	DiffProvider
	CommandExecutor
}

// Ensure Manager implements all interfaces
var (
	_ StatusReader    = (*Manager)(nil)
	_ BranchManager   = (*Manager)(nil)
	_ SyncManager     = (*Manager)(nil)
	_ CommitManager   = (*Manager)(nil)
	_ WorktreeManager = (*Manager)(nil)
	_ DiffProvider    = (*Manager)(nil)
	_ CommandExecutor = (*Manager)(nil)
	_ GitOperations   = (*Manager)(nil)
)
