package git

import (
	"gman/pkg/types"
)

// HybridManager implements a mixed strategy using both go-git and native git
// - Uses go-git for fast, read-only operations (status, branch info)
// - Uses native git for write operations and complex workflows
type HybridManager struct {
	nativeGit    *Manager           // Native git manager for write operations
	goGitStatus  *GoGitStatusReader // go-git for fast status operations
	enableGoGit  bool               // Feature flag to enable/disable go-git optimization
}

// NewHybridManager creates a new hybrid git manager
func NewHybridManager() *HybridManager {
	nativeManager := NewManager()
	goGitStatus := NewGoGitStatusReader(nativeManager)
	
	return &HybridManager{
		nativeGit:   nativeManager,
		goGitStatus: goGitStatus,
		enableGoGit: true, // Enable by default, can be configured
	}
}

// SetGoGitEnabled enables or disables go-git optimization
func (h *HybridManager) SetGoGitEnabled(enabled bool) {
	h.enableGoGit = enabled
}

// IsGoGitEnabled returns whether go-git optimization is enabled
func (h *HybridManager) IsGoGitEnabled() bool {
	return h.enableGoGit
}

// StatusReader interface implementation with hybrid strategy

// GetRepoStatus uses go-git for fast status when enabled, falls back to native git
func (h *HybridManager) GetRepoStatus(alias, path string) types.RepoStatus {
	if h.enableGoGit {
		return h.goGitStatus.GetRepoStatus(alias, path)
	}
	return h.nativeGit.GetRepoStatus(alias, path)
}

// GetRepoStatusNoFetch uses go-git for ultra-fast status without network operations
func (h *HybridManager) GetRepoStatusNoFetch(alias, path string) types.RepoStatus {
	if h.enableGoGit {
		return h.goGitStatus.GetRepoStatusNoFetch(alias, path)
	}
	return h.nativeGit.GetRepoStatusNoFetch(alias, path)
}

// GetAllRepoStatus delegates to the appropriate implementation
func (h *HybridManager) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	if h.enableGoGit {
		return h.goGitStatus.GetAllRepoStatus(repositories)
	}
	return h.nativeGit.GetAllRepoStatus(repositories)
}

// GetAllRepoStatusNoFetch uses go-git for fast bulk status operations
func (h *HybridManager) GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) {
	if h.enableGoGit {
		return h.goGitStatus.GetAllRepoStatusNoFetch(repositories)
	}
	return h.nativeGit.GetAllRepoStatusNoFetch(repositories)
}

// IsGitRepository uses go-git for fast repository detection
func (h *HybridManager) IsGitRepository(path string) bool {
	if h.enableGoGit {
		return h.goGitStatus.IsGitRepository(path)
	}
	return h.nativeGit.IsGitRepository(path)
}

// HasUncommittedChanges uses go-git for fast change detection
func (h *HybridManager) HasUncommittedChanges(path string) (bool, error) {
	if h.enableGoGit {
		return h.goGitStatus.HasUncommittedChanges(path)
	}
	return h.nativeGit.HasUncommittedChanges(path)
}

// HasUnpushedCommits always uses native git (requires network operations)
func (h *HybridManager) HasUnpushedCommits(path string) (bool, error) {
	return h.nativeGit.HasUnpushedCommits(path)
}

// All write operations and complex workflows delegate to native git

// BranchManager interface - mostly read operations can use go-git
func (h *HybridManager) GetCurrentBranch(path string) (string, error) {
	return h.nativeGit.GetCurrentBranch(path)
}

func (h *HybridManager) GetBranches(path string, includeRemote bool) ([]string, error) {
	return h.nativeGit.GetBranches(path, includeRemote)
}

func (h *HybridManager) CreateBranch(path, branchName string) error {
	return h.nativeGit.CreateBranch(path, branchName)
}

func (h *HybridManager) SwitchBranch(path, branchName string) error {
	return h.nativeGit.SwitchBranch(path, branchName)
}

func (h *HybridManager) CleanMergedBranches(path, mainBranch string) ([]string, error) {
	return h.nativeGit.CleanMergedBranches(path, mainBranch)
}

// SyncManager interface - all delegate to native git (requires network)
func (h *HybridManager) SyncRepository(path, mode string) error {
	return h.nativeGit.SyncRepository(path, mode)
}

func (h *HybridManager) SyncAllRepositories(repositories map[string]string, mode string, maxConcurrency int) error {
	return h.nativeGit.SyncAllRepositories(repositories, mode, maxConcurrency)
}

// CommitManager interface - all delegate to native git (write operations)
func (h *HybridManager) CommitChanges(path, message string, addAll bool) error {
	return h.nativeGit.CommitChanges(path, message, addAll)
}

func (h *HybridManager) PushChanges(path string, force, setUpstream bool) error {
	return h.nativeGit.PushChanges(path, force, setUpstream)
}

func (h *HybridManager) StashSave(path, message string) error {
	return h.nativeGit.StashSave(path, message)
}

func (h *HybridManager) StashPop(path string) error {
	return h.nativeGit.StashPop(path)
}

func (h *HybridManager) StashList(path string) ([]string, error) {
	return h.nativeGit.StashList(path)
}

func (h *HybridManager) StashClear(path string) error {
	return h.nativeGit.StashClear(path)
}

// WorktreeManager interface - all delegate to native git
func (h *HybridManager) AddWorktree(repoPath, worktreePath, branch string) error {
	return h.nativeGit.AddWorktree(repoPath, worktreePath, branch)
}

func (h *HybridManager) ListWorktrees(repoPath string) ([]types.Worktree, error) {
	return h.nativeGit.ListWorktrees(repoPath)
}

func (h *HybridManager) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	return h.nativeGit.RemoveWorktree(repoPath, worktreePath, force)
}

// DiffProvider interface - all delegate to native git
func (h *HybridManager) DiffFileBetweenBranches(repoPath, branch1, branch2, filePath string) (string, error) {
	return h.nativeGit.DiffFileBetweenBranches(repoPath, branch1, branch2, filePath)
}

func (h *HybridManager) DiffFileBetweenRepos(repo1Path, repo2Path, filePath string) (string, error) {
	return h.nativeGit.DiffFileBetweenRepos(repo1Path, repo2Path, filePath)
}

func (h *HybridManager) GetFileContentFromBranch(repoPath, branch, filePath string) (string, error) {
	return h.nativeGit.GetFileContentFromBranch(repoPath, branch, filePath)
}

// CommandExecutor interface - delegates to native git
func (h *HybridManager) RunCommand(path string, args ...string) (string, error) {
	return h.nativeGit.RunCommand(path, args...)
}

// Ensure HybridManager implements all interfaces
var (
	_ StatusReader    = (*HybridManager)(nil)
	_ BranchManager   = (*HybridManager)(nil)
	_ SyncManager     = (*HybridManager)(nil)
	_ CommitManager   = (*HybridManager)(nil)
	_ WorktreeManager = (*HybridManager)(nil)
	_ DiffProvider    = (*HybridManager)(nil)
	_ CommandExecutor = (*HybridManager)(nil)
	_ GitOperations   = (*HybridManager)(nil)
)