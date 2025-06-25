package git

import (
	"fmt"
	"os"
	"strings"
)

// BranchManagerImpl implements BranchManager interface
type BranchManagerImpl struct {
	currentDir string
}

// NewBranchManager creates a new branch manager
func NewBranchManager() *BranchManagerImpl {
	currentDir, _ := os.Getwd()
	return &BranchManagerImpl{
		currentDir: currentDir,
	}
}

// GetCurrentBranch returns the current branch name
func (b *BranchManagerImpl) GetCurrentBranch(path string) (string, error) {
	output, err := b.runGitCommand(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// GetBranches returns all branches in the repository
func (b *BranchManagerImpl) GetBranches(path string, includeRemote bool) ([]string, error) {
	args := []string{"branch"}
	if includeRemote {
		args = append(args, "-a")
	}

	output, err := b.runGitCommand(path, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	lines := strings.Split(output, "\n")
	var branches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove the current branch marker (*)
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		} else if strings.HasPrefix(line, "  ") {
			line = strings.TrimPrefix(line, "  ")
		}

		// Skip HEAD pointer for remote branches
		if strings.Contains(line, "HEAD ->") {
			continue
		}

		// Clean up remote branch names
		if includeRemote && strings.HasPrefix(line, "remotes/") {
			line = strings.TrimPrefix(line, "remotes/")
		}

		if line != "" {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// CreateBranch creates a new branch from the current commit
func (b *BranchManagerImpl) CreateBranch(path, branchName string) error {
	// Verify the branch name is valid
	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check if branch already exists
	branches, err := b.GetBranches(path, false)
	if err != nil {
		return fmt.Errorf("failed to check existing branches: %w", err)
	}

	for _, branch := range branches {
		if branch == branchName {
			return fmt.Errorf("branch '%s' already exists", branchName)
		}
	}

	// Create the branch
	_, err = b.runGitCommand(path, "checkout", "-b", branchName)
	if err != nil {
		return fmt.Errorf("failed to create branch '%s': %w", branchName, err)
	}

	return nil
}

// SwitchBranch switches to the specified branch
func (b *BranchManagerImpl) SwitchBranch(path, branchName string) error {
	// Verify the branch exists
	err := b.verifyBranchExists(path, branchName)
	if err != nil {
		return err
	}

	// Check for uncommitted changes
	hasChanges, err := b.hasUncommittedChanges(path)
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if hasChanges {
		return fmt.Errorf("cannot switch branches with uncommitted changes. Please commit or stash your changes first")
	}

	// Switch to the branch
	_, err = b.runGitCommand(path, "checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to switch to branch '%s': %w", branchName, err)
	}

	return nil
}

// CleanMergedBranches removes branches that have been merged into the main branch
func (b *BranchManagerImpl) CleanMergedBranches(path, mainBranch string) ([]string, error) {
	// If no main branch specified, try to detect it
	if mainBranch == "" {
		detectedMain, err := b.detectMainBranch(path)
		if err != nil {
			return nil, fmt.Errorf("failed to detect main branch: %w", err)
		}
		mainBranch = detectedMain
	}

	// Get current branch to avoid deleting it
	currentBranch, err := b.GetCurrentBranch(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get merged branches
	output, err := b.runGitCommand(path, "branch", "--merged", mainBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged branches: %w", err)
	}

	lines := strings.Split(output, "\n")
	var mergedBranches []string
	var deletedBranches []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove the current branch marker (*)
		if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		} else if strings.HasPrefix(line, "  ") {
			line = strings.TrimPrefix(line, "  ")
		}

		// Skip main branch and current branch
		if line == mainBranch || line == currentBranch {
			continue
		}

		mergedBranches = append(mergedBranches, line)
	}

	// Delete merged branches
	for _, branch := range mergedBranches {
		_, err := b.runGitCommand(path, "branch", "-d", branch)
		if err != nil {
			// Log error but continue with other branches
			fmt.Printf("Warning: Failed to delete branch '%s': %v\n", branch, err)
			continue
		}
		deletedBranches = append(deletedBranches, branch)
	}

	return deletedBranches, nil
}

// Private helper methods

func (b *BranchManagerImpl) verifyBranchExists(repoPath, branch string) error {
	// Check local branches first
	branches, err := b.GetBranches(repoPath, false)
	if err != nil {
		return fmt.Errorf("failed to get local branches: %w", err)
	}

	for _, b := range branches {
		if b == branch {
			return nil // Branch exists locally
		}
	}

	// Check remote branches
	remoteBranches, err := b.GetBranches(repoPath, true)
	if err != nil {
		return fmt.Errorf("failed to get remote branches: %w", err)
	}

	for _, b := range remoteBranches {
		if b == fmt.Sprintf("origin/%s", branch) {
			return nil // Branch exists on remote
		}
	}

	return fmt.Errorf("branch '%s' does not exist", branch)
}

func (b *BranchManagerImpl) detectMainBranch(path string) (string, error) {
	// Try common main branch names
	mainBranches := []string{"main", "master", "develop", "dev"}

	branches, err := b.GetBranches(path, false)
	if err != nil {
		return "", err
	}

	for _, mainCandidate := range mainBranches {
		for _, branch := range branches {
			if branch == mainCandidate {
				return mainCandidate, nil
			}
		}
	}

	// If no common main branch found, use the first branch
	if len(branches) > 0 {
		return branches[0], nil
	}

	return "", fmt.Errorf("no branches found in repository")
}

func (b *BranchManagerImpl) hasUncommittedChanges(path string) (bool, error) {
	output, err := b.runGitCommand(path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	return strings.TrimSpace(output) != "", nil
}

func (b *BranchManagerImpl) runGitCommand(path string, args ...string) (string, error) {
	// Delegate to Manager implementation for now
	manager := &Manager{currentDir: b.currentDir}
	output, err := manager.RunCommand(path, args...)
	return output, err
}
