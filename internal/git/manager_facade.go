package git

import (
	"gman/pkg/types"
)

// GitManager is a facade that combines all Git operation interfaces
type GitManager struct {
	status   StatusReader
	branch   BranchManager
	sync     SyncManager
	commit   CommitManager
	worktree WorktreeManager
	diff     DiffProvider
	executor CommandExecutor
}

// NewGitManager creates a new Git manager facade
func NewGitManager() *GitManager {
	// Create the native manager which implements all interfaces
	manager := NewManager()

	return &GitManager{
		status:   manager,
		branch:   manager,
		sync:     manager,
		commit:   manager,
		worktree: manager,
		diff:     manager,
		executor: manager,
	}
}

// NewGitManagerWithComponents creates a Git manager with specific implementations
func NewGitManagerWithComponents(
	status StatusReader,
	branch BranchManager,
	sync SyncManager,
	commit CommitManager,
	worktree WorktreeManager,
	diff DiffProvider,
	executor CommandExecutor,
) *GitManager {
	return &GitManager{
		status:   status,
		branch:   branch,
		sync:     sync,
		commit:   commit,
		worktree: worktree,
		diff:     diff,
		executor: executor,
	}
}

// StatusReader methods
func (g *GitManager) GetRepoStatus(alias, path string) types.RepoStatus {
	return g.status.GetRepoStatus(alias, path)
}

func (g *GitManager) GetRepoStatusNoFetch(alias, path string) types.RepoStatus {
	return g.status.GetRepoStatusNoFetch(alias, path)
}

func (g *GitManager) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	return g.status.GetAllRepoStatus(repositories)
}

func (g *GitManager) GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) {
	return g.status.GetAllRepoStatusNoFetch(repositories)
}

func (g *GitManager) IsGitRepository(path string) bool {
	return g.status.IsGitRepository(path)
}

func (g *GitManager) HasUncommittedChanges(path string) (bool, error) {
	return g.status.HasUncommittedChanges(path)
}

func (g *GitManager) HasUnpushedCommits(path string) (bool, error) {
	return g.status.HasUnpushedCommits(path)
}

// BranchManager methods
func (g *GitManager) GetCurrentBranch(path string) (string, error) {
	return g.branch.GetCurrentBranch(path)
}

func (g *GitManager) GetBranches(path string, includeRemote bool) ([]string, error) {
	return g.branch.GetBranches(path, includeRemote)
}

func (g *GitManager) CreateBranch(path, branchName string) error {
	return g.branch.CreateBranch(path, branchName)
}

func (g *GitManager) SwitchBranch(path, branchName string) error {
	return g.branch.SwitchBranch(path, branchName)
}

func (g *GitManager) CleanMergedBranches(path, mainBranch string) ([]string, error) {
	return g.branch.CleanMergedBranches(path, mainBranch)
}

// SyncManager methods
func (g *GitManager) SyncRepository(path, mode string) error {
	return g.sync.SyncRepository(path, mode)
}

func (g *GitManager) SyncAllRepositories(repositories map[string]string, mode string, maxConcurrency int) error {
	return g.sync.SyncAllRepositories(repositories, mode, maxConcurrency)
}

// CommitManager methods
func (g *GitManager) CommitChanges(path, message string, addAll bool) error {
	return g.commit.CommitChanges(path, message, addAll)
}

func (g *GitManager) PushChanges(path string, force, setUpstream bool) error {
	return g.commit.PushChanges(path, force, setUpstream)
}

func (g *GitManager) StashSave(path, message string) error {
	return g.commit.StashSave(path, message)
}

func (g *GitManager) StashPop(path string) error {
	return g.commit.StashPop(path)
}

func (g *GitManager) StashList(path string) ([]string, error) {
	return g.commit.StashList(path)
}

func (g *GitManager) StashClear(path string) error {
	return g.commit.StashClear(path)
}

// WorktreeManager methods
func (g *GitManager) AddWorktree(repoPath, worktreePath, branch string) error {
	return g.worktree.AddWorktree(repoPath, worktreePath, branch)
}

func (g *GitManager) ListWorktrees(repoPath string) ([]types.Worktree, error) {
	return g.worktree.ListWorktrees(repoPath)
}

func (g *GitManager) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	return g.worktree.RemoveWorktree(repoPath, worktreePath, force)
}

// DiffProvider methods
func (g *GitManager) DiffFileBetweenBranches(repoPath, branch1, branch2, filePath string) (string, error) {
	return g.diff.DiffFileBetweenBranches(repoPath, branch1, branch2, filePath)
}

func (g *GitManager) DiffFileBetweenRepos(repo1Path, repo2Path, filePath string) (string, error) {
	return g.diff.DiffFileBetweenRepos(repo1Path, repo2Path, filePath)
}

func (g *GitManager) GetFileContentFromBranch(repoPath, branch, filePath string) (string, error) {
	return g.diff.GetFileContentFromBranch(repoPath, branch, filePath)
}

// CommandExecutor methods
func (g *GitManager) RunCommand(path string, args ...string) (string, error) {
	return g.executor.RunCommand(path, args...)
}

// Convenience methods for creating specialized managers

// GetStatusReader returns a StatusReader interface
func (g *GitManager) GetStatusReader() StatusReader {
	return g.status
}

// GetBranchManager returns a BranchManager interface
func (g *GitManager) GetBranchManager() BranchManager {
	return g.branch
}

// GetSyncManager returns a SyncManager interface
func (g *GitManager) GetSyncManager() SyncManager {
	return g.sync
}

// GetCommitManager returns a CommitManager interface
func (g *GitManager) GetCommitManager() CommitManager {
	return g.commit
}

// GetWorktreeManager returns a WorktreeManager interface
func (g *GitManager) GetWorktreeManager() WorktreeManager {
	return g.worktree
}

// GetDiffProvider returns a DiffProvider interface
func (g *GitManager) GetDiffProvider() DiffProvider {
	return g.diff
}

// GetCommandExecutor returns a CommandExecutor interface
func (g *GitManager) GetCommandExecutor() CommandExecutor {
	return g.executor
}

// Ensure GitManager implements GitOperations
var _ GitOperations = (*GitManager)(nil)
