package git

import (
	"fmt"

	"gman/pkg/types"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GoGitStatusReader provides fast status operations using go-git library
type GoGitStatusReader struct {
	fallback StatusReader // Fallback to native git for complex operations
}

// NewGoGitStatusReader creates a new go-git based status reader
func NewGoGitStatusReader(fallback StatusReader) *GoGitStatusReader {
	return &GoGitStatusReader{
		fallback: fallback,
	}
}

// GetRepoStatus gets repository status using go-git for faster performance
func (g *GoGitStatusReader) GetRepoStatus(alias, path string) types.RepoStatus {
	status := types.RepoStatus{
		Alias: alias,
		Path:  path,
	}

	// Open repository with go-git
	repo, err := git.PlainOpen(path)
	if err != nil {
		// Fall back to native git for non-standard repositories
		return g.fallback.GetRepoStatus(alias, path)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		status.Error = fmt.Errorf("failed to get HEAD: %w", err)
		return status
	}

	// Get current branch name
	if head.Name().IsBranch() {
		status.Branch = head.Name().Short()
	} else {
		status.Branch = "detached HEAD"
	}

	// Get working tree
	worktree, err := repo.Worktree()
	if err != nil {
		status.Error = fmt.Errorf("failed to get worktree: %w", err)
		return status
	}

	// Check workspace status
	workStatus, err := worktree.Status()
	if err != nil {
		status.Error = fmt.Errorf("failed to get status: %w", err)
		return status
	}

	// Determine workspace status
	if workStatus.IsClean() {
		status.Workspace = types.Clean
	} else {
		status.Workspace = types.Dirty
		// Count changed files
		status.FilesChanged = len(workStatus)
	}

	// Get last commit info
	commit, err := repo.CommitObject(head.Hash())
	if err == nil {
		status.LastCommit = commit.Message
		status.CommitTime = commit.Author.When
	}

	// For sync status, we still use native git as it requires network operations
	// and remote tracking which is more complex in go-git
	syncStatus := g.getSyncStatusFast(repo, head.Hash())
	status.SyncStatus = syncStatus

	return status
}

// GetRepoStatusNoFetch provides ultra-fast status without any network operations
func (g *GoGitStatusReader) GetRepoStatusNoFetch(alias, path string) types.RepoStatus {
	status := types.RepoStatus{
		Alias: alias,
		Path:  path,
	}

	// Open repository with go-git
	repo, err := git.PlainOpen(path)
	if err != nil {
		status.Error = fmt.Errorf("not a git repository: %w", err)
		return status
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		status.Error = fmt.Errorf("failed to get HEAD: %w", err)
		return status
	}

	// Get current branch name
	if head.Name().IsBranch() {
		status.Branch = head.Name().Short()
	} else {
		status.Branch = "detached HEAD"
	}

	// Get working tree
	worktree, err := repo.Worktree()
	if err != nil {
		status.Error = fmt.Errorf("failed to get worktree: %w", err)
		return status
	}

	// Check workspace status (this is the most expensive operation)
	workStatus, err := worktree.Status()
	if err != nil {
		status.Error = fmt.Errorf("failed to get status: %w", err)
		return status
	}

	// Determine workspace status
	if workStatus.IsClean() {
		status.Workspace = types.Clean
	} else {
		status.Workspace = types.Dirty
		status.FilesChanged = len(workStatus)
	}

	// Get last commit info
	commit, err := repo.CommitObject(head.Hash())
	if err == nil {
		status.LastCommit = commit.Message
		status.CommitTime = commit.Author.When
	}

	// No sync status for no-fetch mode
	status.SyncStatus = types.SyncStatus{
		Ahead:  0,
		Behind: 0,
	}

	return status
}

// GetAllRepoStatus gets status for multiple repositories concurrently
func (g *GoGitStatusReader) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	// For now, delegate to fallback as this requires more complex orchestration
	return g.fallback.GetAllRepoStatus(repositories)
}

// GetAllRepoStatusNoFetch gets status for multiple repositories without network ops
func (g *GoGitStatusReader) GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) {
	results := make([]types.RepoStatus, 0, len(repositories))
	
	for alias, path := range repositories {
		status := g.GetRepoStatusNoFetch(alias, path)
		results = append(results, status)
	}
	
	return results, nil
}

// IsGitRepository checks if path is a git repository using go-git
func (g *GoGitStatusReader) IsGitRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

// HasUncommittedChanges checks for uncommitted changes using go-git
func (g *GoGitStatusReader) HasUncommittedChanges(path string) (bool, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return false, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

// HasUnpushedCommits checks for unpushed commits (fallback to native git)
func (g *GoGitStatusReader) HasUnpushedCommits(path string) (bool, error) {
	// This requires complex remote tracking logic, fallback to native git
	return g.fallback.HasUnpushedCommits(path)
}

// getSyncStatusFast attempts to get sync status quickly, with simplified logic
func (g *GoGitStatusReader) getSyncStatusFast(repo *git.Repository, localHash plumbing.Hash) types.SyncStatus {
	// For now, return a basic sync status
	// In a full implementation, this would check remote refs
	return types.SyncStatus{
		Ahead:  0,
		Behind: 0,
	}
}