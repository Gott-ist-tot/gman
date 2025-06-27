package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gman/internal/errors"
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

// GetRepoStatus gets the status of a single repository with fetch
func (g *Manager) GetRepoStatus(alias, path string) types.RepoStatus {
	return g.getRepoStatusInternal(alias, path, true)
}

// GetRepoStatusNoFetch gets the status of a single repository without fetch (for fast loading)
func (g *Manager) GetRepoStatusNoFetch(alias, path string) types.RepoStatus {
	return g.getRepoStatusInternal(alias, path, false)
}

// getRepoStatusInternal gets the status of a single repository
func (g *Manager) getRepoStatusInternal(alias, path string, withFetch bool) types.RepoStatus {
	status := types.RepoStatus{
		Alias: alias,
		Path:  path,
	}

	// Check if path exists and is a git repository
	if !g.isGitRepository(path) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			status.Error = errors.NewRepoNotFoundError(path)
		} else {
			status.Error = errors.NewNotGitRepoError(path)
		}
		return status
	}

	// Check if this is the current working directory
	status.IsCurrent = g.isCurrentRepository(path)

	// Get current branch
	branch, err := g.getCurrentBranch(path)
	if err != nil {
		status.Error = errors.Wrap(err, errors.ErrTypeInternal, 
			fmt.Sprintf("failed to get current branch for repository: %s", alias)).
			WithContext("repository", path).
			WithSuggestion("Verify the repository is in a valid state")
		return status
	}
	status.Branch = branch

	// Get workspace status
	workspaceStatus, err := g.getWorkspaceStatus(path)
	if err != nil {
		status.Error = errors.Wrap(err, errors.ErrTypeInternal,
			fmt.Sprintf("failed to get workspace status for repository: %s", alias)).
			WithContext("repository", path).
			WithSuggestion("Check if the repository working directory is accessible")
		return status
	}
	status.Workspace = workspaceStatus

	// Get sync status
	syncStatus, err := g.getSyncStatusInternal(path, withFetch)
	if err != nil {
		// Sync errors might be network-related, so they're often recoverable
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection") {
			status.Error = errors.NewNetworkTimeoutError("sync status check", "30s").
				WithCause(err).
				WithContext("repository", path)
		} else {
			status.Error = errors.Wrap(err, errors.ErrTypeRemoteUnreachable,
				fmt.Sprintf("failed to get sync status for repository: %s", alias)).
				WithContext("repository", path)
		}
		return status
	}
	status.SyncStatus = syncStatus

	// Get last commit
	lastCommit, err := g.getLastCommit(path)
	if err != nil {
		status.Error = errors.Wrap(err, errors.ErrTypeInternal,
			fmt.Sprintf("failed to get last commit for repository: %s", alias)).
			WithContext("repository", path).
			WithSuggestion("Verify the repository has at least one commit")
		return status
	}
	status.LastCommit = lastCommit

	// Get files changed count
	filesChanged, err := g.getFilesChangedCount(path)
	if err != nil {
		// Don't fail for this, just set to 0
		filesChanged = 0
	}
	status.FilesChanged = filesChanged

	// Get commit time
	commitTime, err := g.getLastCommitTime(path)
	if err != nil {
		// Don't fail for this, use zero time
		commitTime = time.Time{}
	}
	status.CommitTime = commitTime

	// Get enhanced status information (non-blocking)
	// Remote URL
	if remoteURL, err := g.GetRemoteURL(path); err == nil {
		status.RemoteURL = remoteURL
	}

	// Remote branch
	if remoteBranch, err := g.GetRemoteBranch(path); err == nil {
		status.RemoteBranch = remoteBranch
	}

	// Stash count
	if stashCount, err := g.GetStashCount(path); err == nil {
		status.StashCount = stashCount
	}

	// Branch counts
	if local, remote, total, err := g.GetBranchCounts(path); err == nil {
		status.LocalBranches = local
		status.RemoteBranches = remote
		status.TotalBranches = total
	}

	// Last fetch time
	if fetchTime, err := g.GetLastFetchTime(path); err == nil {
		status.LastFetchTime = fetchTime
	}

	return status
}

// GetAllRepoStatus gets status for multiple repositories concurrently
func (g *Manager) GetAllRepoStatus(repositories map[string]string) ([]types.RepoStatus, error) {
	return g.getAllRepoStatusInternal(repositories, true)
}

// GetAllRepoStatusNoFetch gets status for multiple repositories concurrently without fetch
func (g *Manager) GetAllRepoStatusNoFetch(repositories map[string]string) ([]types.RepoStatus, error) {
	return g.getAllRepoStatusInternal(repositories, false)
}

// getAllRepoStatusInternal gets status for multiple repositories with configurable fetch behavior
func (g *Manager) getAllRepoStatusInternal(repositories map[string]string, withFetch bool) ([]types.RepoStatus, error) {
	repoCount := len(repositories)
	if repoCount == 0 {
		return []types.RepoStatus{}, nil
	}

	var wg sync.WaitGroup
	// Use buffered channel with capacity equal to repo count to avoid blocking
	statusChan := make(chan types.RepoStatus, repoCount)

	// Dynamic semaphore size based on repository count
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
				status = g.GetRepoStatus(alias, path)
			} else {
				status = g.GetRepoStatusNoFetch(alias, path)
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
	// Validate path to prevent directory traversal
	if err := g.validatePath(path); err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Validate git arguments to prevent command injection
	if err := g.validateGitArgs(args); err != nil {
		return "", fmt.Errorf("invalid git arguments: %w", err)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	
	// Force English locale to ensure consistent Git output parsing
	// This prevents issues with localized Git messages
	cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C")
	
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// validatePath validates that the path is safe and absolute
func (g *Manager) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Convert to absolute path for safety
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check for directory traversal attempts
	cleanPath := filepath.Clean(absPath)
	if cleanPath != absPath {
		return fmt.Errorf("path contains directory traversal elements")
	}

	// Verify path exists and is a directory
	stat, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	return nil
}

// validateGitArgs validates git command arguments to prevent injection
func (g *Manager) validateGitArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no git arguments provided")
	}

	// Whitelist of allowed git commands for security
	allowedCommands := map[string]bool{
		"status":    true,
		"rev-parse": true,
		"log":       true,
		"fetch":     true,
		"pull":      true,
		"push":      true,
		"checkout":  true,
		"branch":    true,
		"commit":    true,
		"add":       true,
		"diff":      true,
		"show":      true,
		"stash":     true,
		"rev-list":  true,
		"worktree":  true,
		"merge":     true,
		"reset":     true,
		"config":    true, // For test environments only
	}

	if !allowedCommands[args[0]] {
		return fmt.Errorf("git command '%s' is not allowed", args[0])
	}

	// Check for suspicious characters in arguments
	for i, arg := range args {
		if err := g.validateArgument(arg); err != nil {
			return fmt.Errorf("invalid argument at position %d: %w", i, err)
		}
	}

	return nil
}

// validateArgument validates a single command argument
func (g *Manager) validateArgument(arg string) error {
	// Check for shell metacharacters that could be used for injection
	dangerousChars := []string{";", "|", "&", "$", "`", "(", ")", "<", ">", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(arg, char) {
			return fmt.Errorf("argument contains dangerous character: %s", char)
		}
	}

	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("argument contains null byte")
	}

	return nil
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

// IsGitRepository is a public version of isGitRepository for external use
func (g *Manager) IsGitRepository(path string) bool {
	return g.isGitRepository(path)
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
	return g.getSyncStatusInternal(path, true)
}

// getSyncStatusInternal gets the sync status, optionally with remote fetch
func (g *Manager) getSyncStatusInternal(path string, withFetch bool) (types.SyncStatus, error) {
	syncStatus := types.SyncStatus{}

	// Attempt to fetch latest from remote if requested
	if withFetch {
		_, err := g.RunCommand(path, "fetch", "--quiet")
		if err != nil {
			// If fetch fails, mark the sync error but continue with local state
			syncStatus.SyncError = fmt.Errorf("failed to fetch from remote: %w", err)
		}
	}

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

	syncStatus.Ahead = ahead
	syncStatus.Behind = behind

	return syncStatus, nil
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

// getFilesChangedCount gets the number of changed files in the workspace
func (g *Manager) getFilesChangedCount(path string) (int, error) {
	output, err := g.RunCommand(path, "status", "--porcelain")
	if err != nil {
		return 0, fmt.Errorf("failed to get files changed count: %w", err)
	}

	if output == "" {
		return 0, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	return len(lines), nil
}

// getLastCommitTime gets the timestamp of the last commit
func (g *Manager) getLastCommitTime(path string) (time.Time, error) {
	output, err := g.RunCommand(path, "log", "-1", "--pretty=format:%ct")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last commit time: %w", err)
	}

	if output == "" {
		return time.Time{}, nil
	}

	timestamp, err := strconv.ParseInt(output, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

// GetCurrentBranch gets the current branch name for a repository
func (g *Manager) GetCurrentBranch(path string) (string, error) {
	return g.getCurrentBranch(path)
}

// GetBranches gets all branches for a repository
func (g *Manager) GetBranches(path string, includeRemote bool) ([]string, error) {
	var args []string
	if includeRemote {
		args = []string{"branch", "-a"}
	} else {
		args = []string{"branch"}
	}

	output, err := g.RunCommand(path, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	var branches []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove current branch indicator
		if strings.HasPrefix(line, "* ") {
			line = line[2:]
		}

		// Skip remote HEAD references
		if strings.Contains(line, "remotes/origin/HEAD") {
			continue
		}

		// Clean up remote branch names
		if strings.HasPrefix(line, "remotes/origin/") {
			line = strings.TrimPrefix(line, "remotes/origin/")
		}

		branches = append(branches, line)
	}

	// Remove duplicates
	uniqueBranches := make(map[string]bool)
	var result []string
	for _, branch := range branches {
		if !uniqueBranches[branch] {
			uniqueBranches[branch] = true
			result = append(result, branch)
		}
	}

	return result, nil
}

// CreateBranch creates a new branch in the repository
func (g *Manager) CreateBranch(path, branchName string) error {
	// Check if branch already exists
	branches, err := g.GetBranches(path, false)
	if err != nil {
		return fmt.Errorf("failed to check existing branches: %w", err)
	}

	for _, branch := range branches {
		if branch == branchName {
			return fmt.Errorf("branch '%s' already exists", branchName)
		}
	}

	// Create the branch
	return g.runGitCommand(path, "checkout", "-b", branchName)
}

// SwitchBranch switches to the specified branch
func (g *Manager) SwitchBranch(path, branchName string) error {
	// Check if branch exists
	branches, err := g.GetBranches(path, true)
	if err != nil {
		return fmt.Errorf("failed to check existing branches: %w", err)
	}

	branchExists := false
	for _, branch := range branches {
		if branch == branchName {
			branchExists = true
			break
		}
	}

	if !branchExists {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}

	// Switch to the branch
	return g.runGitCommand(path, "checkout", branchName)
}

// CleanMergedBranches removes local branches that have been merged
func (g *Manager) CleanMergedBranches(path, mainBranch string) ([]string, error) {
	// Auto-detect main branch if not specified
	if mainBranch == "" {
		var err error
		mainBranch, err = g.detectMainBranch(path)
		if err != nil {
			return nil, fmt.Errorf("failed to detect main branch: %w", err)
		}
	}

	// Get current branch
	currentBranch, err := g.getCurrentBranch(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get merged branches
	output, err := g.RunCommand(path, "branch", "--merged", mainBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	var cleanedBranches []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove current branch indicator
		if strings.HasPrefix(line, "* ") {
			line = line[2:]
		}

		// Don't delete main branch or current branch
		if line == mainBranch || line == currentBranch {
			continue
		}

		// Delete the branch
		err := g.runGitCommand(path, "branch", "-d", line)
		if err != nil {
			// If force delete is needed, try with -D
			err = g.runGitCommand(path, "branch", "-D", line)
		}

		if err == nil {
			cleanedBranches = append(cleanedBranches, line)
		}
	}

	return cleanedBranches, nil
}

// detectMainBranch tries to detect the main branch (main, master, develop)
func (g *Manager) detectMainBranch(path string) (string, error) {
	possibleMains := []string{"main", "master", "develop"}

	branches, err := g.GetBranches(path, true)
	if err != nil {
		return "", err
	}

	for _, main := range possibleMains {
		for _, branch := range branches {
			if branch == main {
				return main, nil
			}
		}
	}

	// Default to current branch if no common main branch found
	return g.getCurrentBranch(path)
}

// HasUncommittedChanges checks if repository has uncommitted changes
func (g *Manager) HasUncommittedChanges(path string) (bool, error) {
	output, err := g.RunCommand(path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}
	return strings.TrimSpace(output) != "", nil
}

// HasUnpushedCommits checks if repository has unpushed commits
func (g *Manager) HasUnpushedCommits(path string) (bool, error) {
	// Get current branch
	branch, err := g.getCurrentBranch(path)
	if err != nil {
		return false, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if remote tracking branch exists
	remoteRef := fmt.Sprintf("origin/%s", branch)
	_, err = g.RunCommand(path, "rev-parse", "--verify", remoteRef)
	if err != nil {
		// No remote tracking branch, consider as having unpushed commits
		return true, nil
	}

	// Check for commits ahead of remote
	output, err := g.RunCommand(path, "rev-list", "--count", remoteRef+"..HEAD")
	if err != nil {
		return false, fmt.Errorf("failed to check unpushed commits: %w", err)
	}

	ahead, _ := strconv.Atoi(strings.TrimSpace(output))
	return ahead > 0, nil
}

// CommitChanges commits changes in the repository
func (g *Manager) CommitChanges(path, message string, addAll bool) error {
	if addAll {
		// Add all changes
		if err := g.runGitCommand(path, "add", "."); err != nil {
			return fmt.Errorf("failed to add changes: %w", err)
		}
	}

	// Check if there are staged changes
	output, err := g.RunCommand(path, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no staged changes to commit")
	}

	// Commit changes
	return g.runGitCommand(path, "commit", "-m", message)
}

// PushChanges pushes local commits to remote
func (g *Manager) PushChanges(path string, force, setUpstream bool) error {
	args := []string{"push"}

	if force {
		args = append(args, "--force")
	}

	if setUpstream {
		// Get current branch for setting upstream
		branch, err := g.getCurrentBranch(path)
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		args = append(args, "--set-upstream", "origin", branch)
	}

	return g.runGitCommand(path, args...)
}

// StashSave saves current changes to stash
func (g *Manager) StashSave(path, message string) error {
	// Check if there are any changes to stash
	hasChanges, err := g.HasUncommittedChanges(path)
	if err != nil {
		return err
	}

	if !hasChanges {
		return fmt.Errorf("no changes to stash")
	}

	args := []string{"stash", "save"}
	if message != "" {
		args = append(args, message)
	}

	return g.runGitCommand(path, args...)
}

// StashPop applies and removes the latest stash
func (g *Manager) StashPop(path string) error {
	// Check if there are any stashes
	stashes, err := g.StashList(path)
	if err != nil {
		return err
	}

	if len(stashes) == 0 {
		return fmt.Errorf("no stashes to pop")
	}

	return g.runGitCommand(path, "stash", "pop")
}

// StashList returns list of stashes
func (g *Manager) StashList(path string) ([]string, error) {
	output, err := g.RunCommand(path, "stash", "list", "--oneline")
	if err != nil {
		return nil, fmt.Errorf("failed to list stashes: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var stashes []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			// Remove the stash@{n}: prefix for cleaner display
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				stashes = append(stashes, parts[1])
			} else {
				stashes = append(stashes, line)
			}
		}
	}

	return stashes, nil
}

// StashClear removes all stashes
func (g *Manager) StashClear(path string) error {
	return g.runGitCommand(path, "stash", "clear")
}

// GetRemoteURL returns the URL of the origin remote
func (g *Manager) GetRemoteURL(path string) (string, error) {
	output, err := g.RunCommand(path, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// GetRemoteBranch returns the tracking remote branch for the current branch
func (g *Manager) GetRemoteBranch(path string) (string, error) {
	// Get current branch
	currentBranch, err := g.GetCurrentBranch(path)
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	// Get the upstream branch
	output, err := g.RunCommand(path, "rev-parse", "--abbrev-ref", currentBranch+"@{upstream}")
	if err != nil {
		// No upstream configured
		return "", nil
	}
	return strings.TrimSpace(output), nil
}

// GetStashCount returns the number of stashes
func (g *Manager) GetStashCount(path string) (int, error) {
	stashes, err := g.StashList(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get stash count: %w", err)
	}
	return len(stashes), nil
}

// GetBranchCounts returns counts of local and remote branches
func (g *Manager) GetBranchCounts(path string) (local, remote, total int, err error) {
	// Get all branches including remotes
	allBranches, err := g.GetBranches(path, true)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get branches: %w", err)
	}
	
	localCount := 0
	remoteCount := 0
	
	for _, branch := range allBranches {
		if strings.HasPrefix(branch, "remotes/") {
			remoteCount++
		} else {
			localCount++
		}
	}
	
	return localCount, remoteCount, localCount + remoteCount, nil
}

// GetLastFetchTime returns the time of the last fetch operation
func (g *Manager) GetLastFetchTime(path string) (time.Time, error) {
	// Check the modification time of .git/FETCH_HEAD
	fetchHeadPath := filepath.Join(path, ".git", "FETCH_HEAD")
	info, err := os.Stat(fetchHeadPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No fetch has been performed yet
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to check FETCH_HEAD: %w", err)
	}
	return info.ModTime(), nil
}

// DiffFileBetweenBranches compares a specific file between two branches
func (g *Manager) DiffFileBetweenBranches(repoPath, branch1, branch2, filePath string) (string, error) {
	// Verify that both branches exist
	if err := g.verifyBranchExists(repoPath, branch1); err != nil {
		return "", fmt.Errorf("branch '%s' does not exist: %w", branch1, err)
	}
	if err := g.verifyBranchExists(repoPath, branch2); err != nil {
		return "", fmt.Errorf("branch '%s' does not exist: %w", branch2, err)
	}

	// Run git diff command
	diffRef := fmt.Sprintf("%s..%s", branch1, branch2)
	output, err := g.RunCommand(repoPath, "diff", diffRef, "--", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to run git diff: %w", err)
	}

	return output, nil
}

// DiffFileBetweenRepos compares the same file between two different repositories
func (g *Manager) DiffFileBetweenRepos(repo1Path, repo2Path, filePath string) (string, error) {
	// Get absolute paths to the files
	file1Path := filepath.Join(repo1Path, filePath)
	file2Path := filepath.Join(repo2Path, filePath)

	// Check if files exist
	if _, err := os.Stat(file1Path); os.IsNotExist(err) {
		return "", fmt.Errorf("file '%s' does not exist in first repository", filePath)
	}
	if _, err := os.Stat(file2Path); os.IsNotExist(err) {
		return "", fmt.Errorf("file '%s' does not exist in second repository", filePath)
	}

	// Use system diff command to compare files
	cmd := exec.Command("diff", "-u", file1Path, file2Path)
	output, err := cmd.CombinedOutput()

	// diff returns non-zero exit code when files differ, which is expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// diff returns 1 when files differ, 2 for errors
			if exitErr.ExitCode() == 1 {
				return string(output), nil // Files differ, this is normal
			}
			return "", fmt.Errorf("diff command failed: %w", err)
		}
		return "", fmt.Errorf("failed to run diff command: %w", err)
	}

	return string(output), nil
}

// GetFileContentFromBranch retrieves the content of a file from a specific branch
func (g *Manager) GetFileContentFromBranch(repoPath, branch, filePath string) (string, error) {
	// Verify that the branch exists
	if err := g.verifyBranchExists(repoPath, branch); err != nil {
		return "", fmt.Errorf("branch '%s' does not exist: %w", branch, err)
	}

	// Get file content from the specified branch
	fileRef := fmt.Sprintf("%s:%s", branch, filePath)
	content, err := g.RunCommand(repoPath, "show", fileRef)
	if err != nil {
		return "", fmt.Errorf("failed to get file content from branch '%s': %w", branch, err)
	}

	return content, nil
}

// verifyBranchExists checks if a branch exists in the repository
func (g *Manager) verifyBranchExists(repoPath, branch string) error {
	_, err := g.RunCommand(repoPath, "rev-parse", "--verify", branch)
	return err
}

// AddWorktree creates a new Git worktree for the specified branch
func (g *Manager) AddWorktree(repoPath, worktreePath, branch string) error {
	// Check if the worktree path already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("path '%s' already exists", worktreePath)
	}

	// Check if branch exists, if not create it with worktree
	branchExists := g.verifyBranchExists(repoPath, branch) == nil
	
	var cmd []string
	if branchExists {
		// Branch exists, create worktree for existing branch
		cmd = []string{"worktree", "add", worktreePath, branch}
	} else {
		// Branch doesn't exist, create new branch with worktree
		cmd = []string{"worktree", "add", "-b", branch, worktreePath}
	}
	
	args := append([]string{cmd[0]}, cmd[1:]...)
	_, err := g.RunCommand(repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	return nil
}

// ListWorktrees returns all worktrees for the specified repository
func (g *Manager) ListWorktrees(repoPath string) ([]types.Worktree, error) {
	// Get worktree list in porcelain format
	output, err := g.RunCommand(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	if output == "" {
		return []types.Worktree{}, nil
	}

	return g.parseWorktreeList(output)
}

// RemoveWorktree removes a Git worktree
func (g *Manager) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	_, err := g.RunCommand(repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	return nil
}

// parseWorktreeList parses the output of 'git worktree list --porcelain'
func (g *Manager) parseWorktreeList(output string) ([]types.Worktree, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var worktrees []types.Worktree
	var current types.Worktree

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line indicates end of worktree entry
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = types.Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		var key, value string

		if len(parts) == 1 {
			// Single word lines (like "detached", "bare")
			key = parts[0]
			value = ""
		} else {
			key, value = parts[0], parts[1]
		}

		switch key {
		case "worktree":
			current.Path = value
		case "HEAD":
			current.Commit = value
		case "branch":
			// Branch format: "refs/heads/branch-name"
			if strings.HasPrefix(value, "refs/heads/") {
				current.Branch = strings.TrimPrefix(value, "refs/heads/")
			} else {
				current.Branch = value
			}
		case "bare":
			current.IsBare = true
		case "detached":
			current.IsDetached = true
		}
	}

	// Add the last worktree if there's no trailing empty line
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}
