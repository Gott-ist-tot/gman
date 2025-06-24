package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gman/pkg/types"
)

// Manager handles git operations
type Manager struct {
	currentDir string
}

// NewManager creates a new git manager
func NewManager() *Manager {
	currentDir, _ := os.Getwd()
	return &Manager{
		currentDir: currentDir,
	}
}

// GetRepoStatus gets the status of a single repository
func (g *Manager) GetRepoStatus(alias, path string) types.RepoStatus {
	status := types.RepoStatus{
		Alias: alias,
		Path:  path,
	}

	// Check if path exists and is a git repository
	if !g.isGitRepository(path) {
		status.Error = fmt.Errorf("not a git repository")
		return status
	}

	// Check if this is the current working directory
	status.IsCurrent = g.isCurrentRepository(path)

	// Get current branch
	branch, err := g.getCurrentBranch(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = branch

	// Get workspace status
	workspaceStatus, err := g.getWorkspaceStatus(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.Workspace = workspaceStatus

	// Get sync status
	syncStatus, err := g.getSyncStatus(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.SyncStatus = syncStatus

	// Get last commit
	lastCommit, err := g.getLastCommit(path)
	if err != nil {
		status.Error = err
		return status
	}
	status.LastCommit = lastCommit

	return status
}

// GetAllRepoStatus gets status for multiple repositories concurrently
func (g *Manager) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	var wg sync.WaitGroup
	statusChan := make(chan types.RepoStatus, len(repositories))

	for alias, path := range repositories {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			status := g.GetRepoStatus(alias, path)
			statusChan <- status
		}(alias, path)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// Collect results
	var statuses []types.RepoStatus
	for status := range statusChan {
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// SyncRepository synchronizes a repository with remote
func (g *Manager) SyncRepository(path, mode string) error {
	cmd := g.buildSyncCommand(mode)
	return g.runGitCommand(path, cmd...)
}

// SyncAllRepositories synchronizes multiple repositories concurrently
func (g *Manager) SyncAllRepositories(repositories map[string]string, mode string, maxConcurrency int) error {
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for alias, path := range repositories {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			err := g.SyncRepository(path, mode)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("%s: %w", alias, err))
				mu.Unlock()
			}
		}(alias, path)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("sync failed for some repositories: %v", errors)
	}

	return nil
}

// RunCommand runs a git command in the specified repository
func (g *Manager) RunCommand(path string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// isGitRepository checks if the given path is a git repository
func (g *Manager) isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular() // Support for git worktrees
}

// isCurrentRepository checks if the given path is the current working directory
func (g *Manager) isCurrentRepository(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	
	currentAbs, err := filepath.Abs(g.currentDir)
	if err != nil {
		return false
	}
	
	return absPath == currentAbs
}

// getCurrentBranch gets the current branch name
func (g *Manager) getCurrentBranch(path string) (string, error) {
	output, err := g.RunCommand(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return output, nil
}

// getWorkspaceStatus gets the workspace status
func (g *Manager) getWorkspaceStatus(path string) (types.WorkspaceStatus, error) {
	// Check for stashes
	stashOutput, err := g.RunCommand(path, "stash", "list")
	if err == nil && stashOutput != "" {
		return types.Stashed, nil
	}

	// Check for uncommitted changes
	statusOutput, err := g.RunCommand(path, "status", "--porcelain")
	if err != nil {
		return types.Clean, fmt.Errorf("failed to get workspace status: %w", err)
	}

	if statusOutput == "" {
		return types.Clean, nil
	}

	return types.Dirty, nil
}

// getSyncStatus gets the sync status with remote
func (g *Manager) getSyncStatus(path string) (types.SyncStatus, error) {
	// Fetch latest from remote (silently)
	g.RunCommand(path, "fetch", "--quiet")

	// Get current branch
	branch, err := g.getCurrentBranch(path)
	if err != nil {
		return types.SyncStatus{}, err
	}

	// Get remote tracking branch
	remoteRef := fmt.Sprintf("origin/%s", branch)

	// Check if remote branch exists
	_, err = g.RunCommand(path, "rev-parse", "--verify", remoteRef)
	if err != nil {
		// No remote tracking branch
		return types.SyncStatus{}, nil
	}

	// Get ahead/behind counts
	aheadOutput, err := g.RunCommand(path, "rev-list", "--count", remoteRef+"..HEAD")
	if err != nil {
		return types.SyncStatus{}, fmt.Errorf("failed to get ahead count: %w", err)
	}

	behindOutput, err := g.RunCommand(path, "rev-list", "--count", "HEAD.."+remoteRef)
	if err != nil {
		return types.SyncStatus{}, fmt.Errorf("failed to get behind count: %w", err)
	}

	ahead, _ := strconv.Atoi(aheadOutput)
	behind, _ := strconv.Atoi(behindOutput)

	return types.SyncStatus{
		Ahead:  ahead,
		Behind: behind,
	}, nil
}

// getLastCommit gets the last commit message
func (g *Manager) getLastCommit(path string) (string, error) {
	output, err := g.RunCommand(path, "log", "-1", "--pretty=format:%h %s")
	if err != nil {
		return "", fmt.Errorf("failed to get last commit: %w", err)
	}
	return output, nil
}

// buildSyncCommand builds the appropriate sync command based on mode
func (g *Manager) buildSyncCommand(mode string) []string {
	switch mode {
	case "rebase":
		return []string{"pull", "--rebase"}
	case "ff-only":
		return []string{"pull", "--ff-only"}
	case "autostash":
		return []string{"pull", "--autostash"}
	default:
		return []string{"pull", "--ff-only"}
	}
}

// runGitCommand runs a git command in the specified directory
func (g *Manager) runGitCommand(path string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	return cmd.Run()
}