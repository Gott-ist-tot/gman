package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorktreeAddCommand tests the worktree add functionality
func TestWorktreeAddCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_worktree_add_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize test repository
	repoPath := filepath.Join(tempDir, "test-repo")
	if err := initWorktreeTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create basic config (simplified for testing)
	configPath := filepath.Join(tempDir, "config.yml")
	configContent := fmt.Sprintf("repositories:\n  test-repo: %s\n", repoPath)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		checkWorktree bool
		wtPath        string
	}{
		{
			name:          "valid worktree creation",
			args:          []string{"test-repo", filepath.Join(tempDir, "feature-wt"), "--branch", "feature-branch"},
			expectError:   false,
			checkWorktree: true,
			wtPath:        filepath.Join(tempDir, "feature-wt"),
		},
		{
			name:          "missing repository",
			args:          []string{"nonexistent", filepath.Join(tempDir, "feature-wt"), "--branch", "feature"},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
		{
			name:          "missing branch flag",
			args:          []string{"test-repo", filepath.Join(tempDir, "feature-wt2")},
			expectError:   true,
			errorContains: "required flag",
		},
		{
			name:          "existing path",
			args:          []string{"test-repo", repoPath, "--branch", "another-feature"},
			expectError:   true,
			errorContains: "already exists",
		},
		{
			name:          "new branch creation",
			args:          []string{"test-repo", filepath.Join(tempDir, "new-feature-wt"), "--branch", "new-feature"},
			expectError:   false,
			checkWorktree: true,
			wtPath:        filepath.Join(tempDir, "new-feature-wt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			// Reset flags and command
			worktreeAddCmd.ResetFlags()
			worktreeAddCmd.Flags().StringVarP(&worktreeBranch, "branch", "b", "", "Branch to checkout in the new worktree (required)")
			worktreeAddCmd.MarkFlagRequired("branch")

			var stdout, stderr bytes.Buffer
			worktreeAddCmd.SetOut(&stdout)
			worktreeAddCmd.SetErr(&stderr)

			worktreeAddCmd.SetArgs(tt.args)
			err := worktreeAddCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.checkWorktree {
					// Verify worktree was created
					if _, err := os.Stat(tt.wtPath); os.IsNotExist(err) {
						t.Errorf("Worktree directory was not created: %s", tt.wtPath)
					}

					// Verify it's a valid git worktree
					gitDir := filepath.Join(tt.wtPath, ".git")
					if _, err := os.Stat(gitDir); os.IsNotExist(err) {
						t.Errorf("Worktree .git file was not created: %s", gitDir)
					}
				}
			}
		})
	}
}

// TestWorktreeListCommand tests the worktree list functionality
func TestWorktreeListCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_worktree_list_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "test-repo")
	if err := initWorktreeTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create a worktree for testing
	wtPath := filepath.Join(tempDir, "test-worktree")
	if err := createTestWorktree(t, repoPath, wtPath, "test-branch"); err != nil {
		t.Fatalf("Failed to create test worktree: %v", err)
	}

	// Create basic config (simplified for testing)
	configPath := filepath.Join(tempDir, "config.yml")
	configContent := fmt.Sprintf("repositories:\n  test-repo: %s\n", repoPath)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		errorContains  string
		expectOutput   bool
		outputContains []string
	}{
		{
			name:           "list existing worktrees",
			args:           []string{"test-repo"},
			expectError:    false,
			expectOutput:   true,
			outputContains: []string{"test-worktree", "test-branch"},
		},
		{
			name:          "missing repository",
			args:          []string{"nonexistent"},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
		{
			name:        "empty worktree list",
			args:        []string{"empty-repo"},
			expectError: false,
		},
	}

	// Add empty repo for testing
	emptyRepoPath := filepath.Join(tempDir, "empty-repo")
	if err := initEmptyTestRepository(t, emptyRepoPath); err != nil {
		t.Fatalf("Failed to create empty repo: %v", err)
	}

	// Update config to include empty repo
	if err := createTestConfig(t, configPath, map[string]string{
		"test-repo":  repoPath,
		"empty-repo": emptyRepoPath,
	}); err != nil {
		t.Fatalf("Failed to update test config: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			worktreeListCmd.ResetFlags()

			var stdout, stderr bytes.Buffer
			worktreeListCmd.SetOut(&stdout)
			worktreeListCmd.SetErr(&stderr)

			worktreeListCmd.SetArgs(tt.args)
			err := worktreeListCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				output := stdout.String()
				if tt.expectOutput && output == "" {
					t.Error("Expected output but got none")
				}

				for _, expected := range tt.outputContains {
					if !strings.Contains(output, expected) {
						t.Errorf("Expected output to contain '%s', got: %s", expected, output)
					}
				}
			}
		})
	}
}

// TestWorktreeRemoveCommand tests the worktree remove functionality
func TestWorktreeRemoveCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_worktree_remove_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "test-repo")
	if err := initWorktreeTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create basic config (simplified for testing)
	configPath := filepath.Join(tempDir, "config.yml")
	configContent := fmt.Sprintf("repositories:\n  test-repo: %s\n", repoPath)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name          string
		setupWorktree bool
		wtPath        string
		addChanges    bool
		args          []string
		expectError   bool
		errorContains string
		checkRemoved  bool
	}{
		{
			name:          "remove clean worktree",
			setupWorktree: true,
			wtPath:        filepath.Join(tempDir, "clean-wt"),
			args:          []string{"test-repo", filepath.Join(tempDir, "clean-wt")},
			expectError:   false,
			checkRemoved:  true,
		},
		{
			name:          "remove worktree with changes (should fail)",
			setupWorktree: true,
			wtPath:        filepath.Join(tempDir, "dirty-wt"),
			addChanges:    true,
			args:          []string{"test-repo", filepath.Join(tempDir, "dirty-wt")},
			expectError:   true,
			errorContains: "uncommitted changes",
		},
		{
			name:          "force remove worktree with changes",
			setupWorktree: true,
			wtPath:        filepath.Join(tempDir, "force-wt"),
			addChanges:    true,
			args:          []string{"test-repo", filepath.Join(tempDir, "force-wt"), "--force"},
			expectError:   false,
			checkRemoved:  true,
		},
		{
			name:          "remove nonexistent worktree",
			args:          []string{"test-repo", filepath.Join(tempDir, "nonexistent-wt")},
			expectError:   true,
			errorContains: "failed to remove worktree",
		},
		{
			name:          "missing repository",
			args:          []string{"nonexistent", filepath.Join(tempDir, "some-wt")},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup worktree if needed
			if tt.setupWorktree {
				branchName := fmt.Sprintf("test-branch-%s", strings.ReplaceAll(tt.name, " ", "-"))
				if err := createTestWorktree(t, repoPath, tt.wtPath, branchName); err != nil {
					t.Fatalf("Failed to create test worktree: %v", err)
				}

				// Add uncommitted changes if requested
				if tt.addChanges {
					testFile := filepath.Join(tt.wtPath, "dirty.txt")
					if err := os.WriteFile(testFile, []byte("uncommitted content"), 0644); err != nil {
						t.Fatalf("Failed to create dirty file: %v", err)
					}
				}
			}

			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			// Reset flags and add force flag
			worktreeRemoveCmd.ResetFlags()
			worktreeRemoveCmd.Flags().BoolVarP(&worktreeForce, "force", "f", false, "Force removal even with uncommitted changes")

			var stdout, stderr bytes.Buffer
			worktreeRemoveCmd.SetOut(&stdout)
			worktreeRemoveCmd.SetErr(&stderr)

			worktreeRemoveCmd.SetArgs(tt.args)
			err := worktreeRemoveCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.checkRemoved {
					// Verify worktree was removed
					if _, err := os.Stat(tt.wtPath); !os.IsNotExist(err) {
						t.Errorf("Worktree directory was not removed: %s", tt.wtPath)
					}
				}
			}

			// Reset force flag for next test
			worktreeForce = false
		})
	}
}

// TestWorktreeComplexScenarios tests complex worktree management scenarios
func TestWorktreeComplexScenarios(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_worktree_complex_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "test-repo")
	if err := initWorktreeTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create basic config (simplified for testing)
	configPath := filepath.Join(tempDir, "config.yml")
	configContent := fmt.Sprintf("repositories:\n  test-repo: %s\n", repoPath)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	os.Setenv("GMAN_CONFIG", configPath)
	defer os.Unsetenv("GMAN_CONFIG")

	t.Run("multiple worktree lifecycle", func(t *testing.T) {
		// Create multiple worktrees
		worktrees := []struct {
			name   string
			path   string
			branch string
		}{
			{"feature-1", filepath.Join(tempDir, "feature-1"), "feature-1"},
			{"feature-2", filepath.Join(tempDir, "feature-2"), "feature-2"},
			{"hotfix", filepath.Join(tempDir, "hotfix"), "hotfix"},
		}

		// Create all worktrees
		for _, wt := range worktrees {
			worktreeAddCmd.ResetFlags()
			worktreeAddCmd.Flags().StringVarP(&worktreeBranch, "branch", "b", "", "Branch to checkout in the new worktree (required)")
			worktreeAddCmd.MarkFlagRequired("branch")

			args := []string{"test-repo", wt.path, "--branch", wt.branch}
			worktreeAddCmd.SetArgs(args)

			if err := worktreeAddCmd.Execute(); err != nil {
				t.Errorf("Failed to create worktree %s: %v", wt.name, err)
			}
		}

		// List all worktrees
		worktreeListCmd.ResetFlags()
		var stdout bytes.Buffer
		worktreeListCmd.SetOut(&stdout)
		worktreeListCmd.SetArgs([]string{"test-repo"})

		if err := worktreeListCmd.Execute(); err != nil {
			t.Errorf("Failed to list worktrees: %v", err)
		}

		output := stdout.String()
		for _, wt := range worktrees {
			if !strings.Contains(output, wt.branch) {
				t.Errorf("Expected to find worktree branch %s in output", wt.branch)
			}
		}

		// Remove all worktrees
		for _, wt := range worktrees {
			worktreeRemoveCmd.ResetFlags()
			worktreeRemoveCmd.Flags().BoolVarP(&worktreeForce, "force", "f", false, "Force removal even with uncommitted changes")

			args := []string{"test-repo", wt.path}
			worktreeRemoveCmd.SetArgs(args)

			if err := worktreeRemoveCmd.Execute(); err != nil {
				t.Errorf("Failed to remove worktree %s: %v", wt.name, err)
			}
		}
	})

	t.Run("worktree with existing branch", func(t *testing.T) {
		// Create a branch in main repo
		cmd := exec.Command("git", "checkout", "-b", "existing-branch")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create existing branch: %v", err)
		}

		// Switch back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to switch back to main: %v", err)
		}

		// Create worktree with existing branch
		worktreeAddCmd.ResetFlags()
		worktreeAddCmd.Flags().StringVarP(&worktreeBranch, "branch", "b", "", "Branch to checkout in the new worktree (required)")
		worktreeAddCmd.MarkFlagRequired("branch")

		wtPath := filepath.Join(tempDir, "existing-branch-wt")
		args := []string{"test-repo", wtPath, "--branch", "existing-branch"}
		worktreeAddCmd.SetArgs(args)

		if err := worktreeAddCmd.Execute(); err != nil {
			t.Errorf("Failed to create worktree with existing branch: %v", err)
		}

		// Verify worktree was created and is on correct branch
		cmd = exec.Command("git", "branch", "--show-current")
		cmd.Dir = wtPath
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to get current branch in worktree: %v", err)
		}

		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != "existing-branch" {
			t.Errorf("Expected worktree to be on 'existing-branch', got '%s'", currentBranch)
		}
	})
}

// Helper functions specific to worktree testing

// initWorktreeTestRepository creates a test repository suitable for worktree operations
func initWorktreeTestRepository(t *testing.T, repoPath string) error {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Configure git user
	cmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure git: %w", err)
		}
	}

	// Create initial file and commit
	testFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repository\n\nThis is a test repository for worktree operations.\n"), 0644); err != nil {
		return err
	}

	cmds = [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "Initial commit"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create initial commit: %w", err)
		}
	}

	return nil
}

// createTestWorktree creates a worktree for testing purposes
func createTestWorktree(t *testing.T, repoPath, wtPath, branchName string) error {
	t.Helper()

	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", branchName)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create test worktree: %w", err)
	}

	return nil
}

// initEmptyTestRepository creates an empty test repository with no worktrees
func initEmptyTestRepository(t *testing.T, repoPath string) error {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init empty git repo: %w", err)
	}

	// Configure git user
	cmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure empty repo: %w", err)
		}
	}

	return nil
}
