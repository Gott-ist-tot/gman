package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManager_verifyBranchExists(t *testing.T) {
	// This is a unit test for the verifyBranchExists method
	// We'll test with the current repository since we know it exists
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to the repository root
	repoRoot := filepath.Join(wd, "..", "..")
	
	manager := NewManager()

	// Test with a branch that should exist (main or master)
	err = manager.verifyBranchExists(repoRoot, "main")
	if err != nil {
		// Try master if main doesn't exist
		err = manager.verifyBranchExists(repoRoot, "master")
		if err != nil {
			t.Logf("Neither 'main' nor 'master' branch found, skipping test")
			t.Skip("No main/master branch found")
		}
	}

	// Test with a branch that definitely doesn't exist
	err = manager.verifyBranchExists(repoRoot, "non-existent-branch-12345")
	if err == nil {
		t.Error("Expected error for non-existent branch, but got nil")
	}
}

func TestManager_DiffFileBetweenBranches(t *testing.T) {
	// This test requires a git repository with multiple branches
	// We'll check if we're in a git repo first
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	repoRoot := filepath.Join(wd, "..", "..")
	manager := NewManager()

	// Check if this is a git repository
	if !manager.isGitRepository(repoRoot) {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Test with same branch should return empty diff
	_, err = manager.DiffFileBetweenBranches(repoRoot, "HEAD", "HEAD", "README.md")
	if err != nil {
		t.Errorf("Expected no error for same branch diff, got: %v", err)
	}

	// Test with non-existent file should return empty output (no error if branches exist)
	_, err = manager.DiffFileBetweenBranches(repoRoot, "HEAD", "HEAD", "non-existent-file.txt")
	if err != nil {
		t.Logf("Got expected error for non-existent file: %v", err)
	}
}

func TestManager_GetFileContentFromBranch(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	repoRoot := filepath.Join(wd, "..", "..")
	manager := NewManager()

	// Check if this is a git repository
	if !manager.isGitRepository(repoRoot) {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Test getting content of a file that should exist
	content, err := manager.GetFileContentFromBranch(repoRoot, "HEAD", "go.mod")
	if err != nil {
		t.Errorf("Failed to get file content: %v", err)
	}

	if content == "" {
		t.Error("Expected non-empty content for go.mod")
	}

	// Test with non-existent file should return error
	_, err = manager.GetFileContentFromBranch(repoRoot, "HEAD", "non-existent-file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, but got nil")
	}
}

func TestManager_DiffFileBetweenRepos(t *testing.T) {
	// This test would require two separate repositories
	// For now, we'll just test the basic functionality with same repo
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	repoRoot := filepath.Join(wd, "..", "..")
	manager := NewManager()

	// Test with same repository should work
	_, err = manager.DiffFileBetweenRepos(repoRoot, repoRoot, "go.mod")
	if err != nil {
		t.Errorf("Expected no error for same repo diff, got: %v", err)
	}

	// Test with non-existent file should return error
	_, err = manager.DiffFileBetweenRepos(repoRoot, repoRoot, "non-existent-file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, but got nil")
	}
}

func TestManager_parseWorktreeList(t *testing.T) {
	manager := NewManager()
	
	// Test with empty output
	worktrees, err := manager.parseWorktreeList("")
	if err != nil {
		t.Errorf("Expected no error for empty input, got: %v", err)
	}
	if len(worktrees) != 0 {
		t.Errorf("Expected 0 worktrees for empty input, got: %d", len(worktrees))
	}

	// Test with sample worktree output
	sampleOutput := `worktree /Users/test/repo
HEAD 1234567890abcdef1234567890abcdef12345678
branch refs/heads/main

worktree /Users/test/repo-feature
HEAD abcdef1234567890abcdef1234567890abcdef12
branch refs/heads/feature-branch

worktree /Users/test/repo-detached
HEAD fedcba0987654321fedcba0987654321fedcba09
detached`

	worktrees, err = manager.parseWorktreeList(sampleOutput)
	if err != nil {
		t.Errorf("Expected no error for valid input, got: %v", err)
	}
	
	if len(worktrees) != 3 {
		t.Errorf("Expected 3 worktrees, got: %d", len(worktrees))
		return
	}

	// Verify first worktree
	if worktrees[0].Path != "/Users/test/repo" {
		t.Errorf("Expected path '/Users/test/repo', got: %s", worktrees[0].Path)
	}
	if worktrees[0].Branch != "main" {
		t.Errorf("Expected branch 'main', got: %s", worktrees[0].Branch)
	}
	if worktrees[0].IsDetached {
		t.Error("Expected not detached")
	}

	// Verify detached worktree
	if !worktrees[2].IsDetached {
		t.Errorf("Expected detached worktree, got IsDetached=%v", worktrees[2].IsDetached)
	}
}

func TestManager_ListWorktrees(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	repoRoot := filepath.Join(wd, "..", "..")
	manager := NewManager()

	// Check if this is a git repository
	if !manager.isGitRepository(repoRoot) {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// List worktrees (should work even if there are no additional worktrees)
	worktrees, err := manager.ListWorktrees(repoRoot)
	if err != nil {
		t.Errorf("Failed to list worktrees: %v", err)
	}

	// Should have at least the main worktree
	if len(worktrees) == 0 {
		t.Error("Expected at least one worktree (the main repository)")
	}
}