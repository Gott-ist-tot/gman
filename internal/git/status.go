package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gman/pkg/types"
)

// StatusManager implements StatusReader interface
type StatusManager struct {
	currentDir string
}

// NewStatusManager creates a new status manager
func NewStatusManager() *StatusManager {
	currentDir, _ := os.Getwd()
	return &StatusManager{
		currentDir: currentDir,
	}
}

// GetRepoStatus gets the status of a single repository with fetch
func (s *StatusManager) GetRepoStatus(alias, path string) types.RepoStatus {
	return s.getRepoStatusInternal(alias, path, true)
}

// GetRepoStatusNoFetch gets the status of a single repository without fetch (for fast loading)
func (s *StatusManager) GetRepoStatusNoFetch(alias, path string) types.RepoStatus {
	return s.getRepoStatusInternal(alias, path, false)
}

// getRepoStatusInternal gets the status of a single repository
func (s *StatusManager) getRepoStatusInternal(alias, path string, withFetch bool) types.RepoStatus {
	status := types.RepoStatus{
		Alias: alias,
		Path:  path,
	}

	// Check if path exists and is a git repository
	if !s.isGitRepository(path) {
		status.Error = fmt.Errorf("not a git repository")
		return status
	}

	// Check if this is the current working directory
	status.IsCurrent = s.isCurrentRepository(path)

	// Get current branch
	branch, err := s.getCurrentBranch(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = branch

	// Get workspace status
	workspaceStatus, err := s.getWorkspaceStatus(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.Workspace = workspaceStatus

	// Get sync status
	syncStatus, err := s.getSyncStatusInternal(path, withFetch)
	if err != nil {
		status.Error = err
		return status
	}
	status.SyncStatus = syncStatus

	// Get last commit
	lastCommit, err := s.getLastCommit(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.LastCommit = lastCommit

	// Get files changed count
	filesChanged, err := s.getFilesChangedCount(path)
	if err != nil {
		// Don't fail for this, just set to 0
		filesChanged = 0
	}
	status.FilesChanged = filesChanged

	// Get commit time
	commitTime, err := s.getLastCommitTime(path)
	if err != nil {
		// Don't fail for this, use zero time
		commitTime = time.Time{}
	}
	status.CommitTime = commitTime

	return status
}

// GetAllRepoStatus gets status for multiple repositories concurrently
func (s *StatusManager) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	return s.getAllRepoStatusInternal(repositories, true)
}

// GetAllRepoStatusNoFetch gets status for multiple repositories concurrently without fetch
func (s *StatusManager) GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) {
	return s.getAllRepoStatusInternal(repositories, false)
}

// getAllRepoStatusInternal gets status for multiple repositories with configurable fetch behavior
func (s *StatusManager) getAllRepoStatusInternal(repositories map[string]string, withFetch bool) ([]types.RepoStatus, error) {
	repoCount := len(repositories)
	if repoCount == 0 {
		return []types.RepoStatus{}, nil
	}

	var wg sync.WaitGroup
	// Use buffered channel with capacity equal to repo count to avoid blocking
	statusChan := make(chan types.RepoStatus, repoCount)

	// Dynamic semaphore size based on repository count
	// For small numbers, use fewer goroutines to reduce overhead
	// For large numbers, cap at reasonable limit to prevent resource exhaustion
	maxConcurrency := 5
	if repoCount < 3 {
		maxConcurrency = repoCount
	} else if repoCount > 20 {
		maxConcurrency = 10 // Increase for large repo counts but cap to prevent memory issues
	}
	
	semaphore := make(chan struct{}, maxConcurrency)

	for alias, path := range repositories {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			var status types.RepoStatus
			if withFetch {
				status = s.GetRepoStatus(alias, path)
			} else {
				status = s.GetRepoStatusNoFetch(alias, path)
			}
			statusChan <- status
		}(alias, path)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// Pre-allocate slice with known capacity to reduce memory allocations
	statuses := make([]types.RepoStatus, 0, repoCount)
	for status := range statusChan {
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// IsGitRepository checks if a path is a git repository
func (s *StatusManager) IsGitRepository(path string) bool {
	return s.isGitRepository(path)
}

// HasUncommittedChanges checks if repository has uncommitted changes
func (s *StatusManager) HasUncommittedChanges(path string) (bool, error) {
	// Use git status --porcelain to check for changes
	output, err := s.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// If output is empty, no changes
	return strings.TrimSpace(output) != "", nil
}

// HasUnpushedCommits checks if repository has unpushed commits
func (s *StatusManager) HasUnpushedCommits(path string) (bool, error) {
	// Get current branch
	branch, err := s.getCurrentBranch(path)
	if err != nil {
		return false, err
	}

	// Check if upstream exists
	_, err = s.runGitCommand(path, "rev-parse", "--verify", fmt.Sprintf("origin/%s", branch))
	if err != nil {
		// No upstream branch, so no unpushed commits can be determined
		return false, nil
	}

	// Check for commits ahead of origin
	output, err := s.runGitCommand(path, "rev-list", "--count", fmt.Sprintf("origin/%s..HEAD", branch))
	if err != nil {
		return false, fmt.Errorf("failed to check unpushed commits: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return false, fmt.Errorf("failed to parse commit count: %w", err)
	}

	return count > 0, nil
}

// Private helper methods

func (s *StatusManager) isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *StatusManager) isCurrentRepository(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absCurrent, err := filepath.Abs(s.currentDir)
	if err != nil {
		return false
	}

	return absPath == absCurrent
}

func (s *StatusManager) getCurrentBranch(path string) (string, error) {
	output, err := s.runGitCommand(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (s *StatusManager) getWorkspaceStatus(path string) (types.WorkspaceStatus, error) {
	// Check for uncommitted changes
	hasChanges, err := s.HasUncommittedChanges(path)
	if err != nil {
		return types.Clean, err
	}

	if hasChanges {
		return types.Dirty, nil
	}

	// Check for stashes
	output, err := s.runGitCommand(path, "stash", "list")
	if err != nil {
		// If stash command fails, assume clean
		return types.Clean, nil
	}

	if strings.TrimSpace(output) != "" {
		return types.Stashed, nil
	}

	return types.Clean, nil
}

func (s *StatusManager) getSyncStatus(path string) (types.SyncStatus, error) {
	return s.getSyncStatusInternal(path, true)
}

func (s *StatusManager) getSyncStatusInternal(path string, withFetch bool) (types.SyncStatus, error) {
	status := types.SyncStatus{}

	// Get current branch
	branch, err := s.getCurrentBranch(path)
	if err != nil {
		return status, err
	}

	// Try to fetch latest remote info if requested, but continue even if it fails
	if withFetch {
		_, fetchErr := s.runGitCommand(path, "fetch", "origin", "--quiet")
		if fetchErr != nil {
			// Store fetch error but continue with cached remote state
			status.SyncError = fmt.Errorf("failed to fetch from remote: %w", fetchErr)
		}
	}

	// Check if upstream exists
	upstream := fmt.Sprintf("origin/%s", branch)
	_, err = s.runGitCommand(path, "rev-parse", "--verify", upstream)
	if err != nil {
		// No upstream, so no sync status - but preserve fetch error if it occurred
		return status, nil
	}

	// Get ahead/behind counts using cached remote state
	output, err := s.runGitCommand(path, "rev-list", "--left-right", "--count", fmt.Sprintf("%s...HEAD", upstream))
	if err != nil {
		// If we already have a fetch error, preserve it; otherwise report sync calculation error
		if status.SyncError == nil {
			status.SyncError = fmt.Errorf("failed to calculate sync status: %w", err)
		}
		return status, nil
	}

	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) >= 2 {
		if behind, err := strconv.Atoi(parts[0]); err == nil {
			status.Behind = behind
		}
		if ahead, err := strconv.Atoi(parts[1]); err == nil {
			status.Ahead = ahead
		}
	}

	return status, nil
}

func (s *StatusManager) getLastCommit(path string) (string, error) {
	output, err := s.runGitCommand(path, "log", "-1", "--pretty=format:%s")
	if err != nil {
		return "", fmt.Errorf("failed to get last commit: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (s *StatusManager) getFilesChangedCount(path string) (int, error) {
	output, err := s.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return 0, fmt.Errorf("failed to get changed files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}

	return len(lines), nil
}

func (s *StatusManager) getLastCommitTime(path string) (time.Time, error) {
	output, err := s.runGitCommand(path, "log", "-1", "--pretty=format:%ct")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get commit time: %w", err)
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit time: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

func (s *StatusManager) runGitCommand(path string, args ...string) (string, error) {
	// This would use the same logic as Manager.runGitCommand
	// For now, we'll delegate to a Manager instance
	manager := &Manager{currentDir: s.currentDir}
	output, err := manager.RunCommand(path, args...)
	return output, err
}
