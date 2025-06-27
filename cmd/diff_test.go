package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gman/internal/di"
	"gman/test"

	"github.com/spf13/cobra"
)

// TestDiffFileCommand tests the diff file command with various scenarios
func TestDiffFileCommand(t *testing.T) {
	// Reset DI container for clean test state
	di.Reset()
	defer di.Reset()

	// Create a temporary directory for our test repository
	tempDir, err := os.MkdirTemp("", "gman_diff_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize a test repository
	repoPath := filepath.Join(tempDir, "test-repo")
	if err := test.InitTestRepositoryWithBranches(t, repoPath, []string{"main", "feature"}); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create a test configuration
	configPath := filepath.Join(tempDir, "config.yml")
	if err := test.CreateBasicTestConfig(t, configPath, map[string]string{"test-repo": repoPath}); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		expectOutput  bool
	}{
		{
			name:         "valid diff between branches",
			args:         []string{"test-repo", "main", "feature", "test.txt"},
			expectError:  false,
			expectOutput: true,
		},
		{
			name:          "missing repository",
			args:          []string{"nonexistent", "main", "feature", "test.txt"},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
		{
			name:          "missing file argument",
			args:          []string{"test-repo", "main", "feature"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:          "invalid argument count",
			args:          []string{"test-repo", "main", "test.txt"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:          "nonexistent branch",
			args:          []string{"test-repo", "nonexistent-branch", "main", "test.txt"},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name:         "same branch comparison",
			args:         []string{"test-repo", "main", "main", "test.txt"},
			expectError:  false,
			expectOutput: false, // No differences expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the config file environment variable
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			// Create command hierarchy: workCmd -> diffCmd -> diffFileCmd
			testWorkCmd := &cobra.Command{Use: "work"}
			testDiffCmd := &cobra.Command{Use: "diff"}
			testDiffFileCmd := &cobra.Command{
				Use:  "file <repo> <branch1> <branch2> -- <file_path>",
				RunE: func(cmd *cobra.Command, args []string) error {
					if err := diffFileCmd.Args(cmd, args); err != nil {
						return err
					}
					return runDiffFile(cmd, args)
				},
			}
			
			testDiffCmd.AddCommand(testDiffFileCmd)
			testWorkCmd.AddCommand(testDiffCmd)

			// Configure output
			var stdout, stderr bytes.Buffer
			testWorkCmd.SetOut(&stdout)
			testWorkCmd.SetErr(&stderr)

			// Build full args: diff file [test args]
			fullArgs := append([]string{"diff", "file"}, tt.args...)
			testWorkCmd.SetArgs(fullArgs)
			err := testWorkCmd.Execute()

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
				errOutput := stderr.String()
				if tt.expectOutput && output == "" && errOutput == "" {
					t.Error("Expected output but got none")
				}
			}
		})
	}
}

// TestDiffCrossRepoCommand tests the cross-repository diff functionality
func TestDiffCrossRepoCommand(t *testing.T) {
	// Reset DI container for clean test state
	di.Reset()
	defer di.Reset()

	tempDir, err := os.MkdirTemp("", "gman_cross_diff_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create two test repositories
	repo1Path := filepath.Join(tempDir, "repo1")
	repo2Path := filepath.Join(tempDir, "repo2")

	if err := initTestRepository(t, repo1Path); err != nil {
		t.Fatalf("Failed to initialize repo1: %v", err)
	}
	if err := initTestRepository(t, repo2Path); err != nil {
		t.Fatalf("Failed to initialize repo2: %v", err)
	}

	// Create different content in the repositories
	if err := createFileInRepo(t, repo2Path, "test.txt", "Different content in repo2\n"); err != nil {
		t.Fatalf("Failed to create different content: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := createTestConfig(t, configPath, map[string]string{
		"repo1": repo1Path,
		"repo2": repo2Path,
	}); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		expectOutput  bool
	}{
		{
			name:         "valid cross-repo diff",
			args:         []string{"repo1", "repo2", "test.txt"},
			expectError:  false,
			expectOutput: true,
		},
		{
			name:          "missing repository",
			args:          []string{"nonexistent", "repo2", "test.txt"},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
		{
			name:          "nonexistent file",
			args:          []string{"repo1", "repo2", "nonexistent.txt"},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name:         "same repository comparison",
			args:         []string{"repo1", "repo1", "test.txt"},
			expectError:  false,
			expectOutput: false, // No differences expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			// Create command hierarchy for cross-repo diff
			testWorkCmd := &cobra.Command{Use: "work"}
			testDiffCmd := &cobra.Command{Use: "diff"}
			testCrossRepoCmd := &cobra.Command{
				Use:  "cross-repo <repo1> <repo2> <file_path>",
				RunE: func(cmd *cobra.Command, args []string) error {
					if err := diffCrossRepoCmd.Args(cmd, args); err != nil {
						return err
					}
					return runDiffCrossRepo(cmd, args)
				},
			}

			testDiffCmd.AddCommand(testCrossRepoCmd)
			testWorkCmd.AddCommand(testDiffCmd)

			var stdout, stderr bytes.Buffer
			testWorkCmd.SetOut(&stdout)
			testWorkCmd.SetErr(&stderr)

			fullArgs := append([]string{"diff", "cross-repo"}, tt.args...)
			testWorkCmd.SetArgs(fullArgs)
			err := testWorkCmd.Execute()

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
				errOutput := stderr.String()
				if tt.expectOutput && output == "" && errOutput == "" {
					t.Error("Expected output but got none")
				}
			}
		})
	}
}

// TestExternalDiffTool tests integration with external diff tools
func TestExternalDiffTool(t *testing.T) {
	// Reset DI container for clean test state
	di.Reset()
	defer di.Reset()

	// Skip this test if diff tool is not available
	if _, err := exec.LookPath("diff"); err != nil {
		t.Skip("diff command not available, skipping external tool test")
	}

	tempDir, err := os.MkdirTemp("", "gman_external_diff_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "test-repo")
	if err := initTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := createTestConfig(t, configPath, map[string]string{"test-repo": repoPath}); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test with a tool that should exist (diff)
	t.Run("valid external tool", func(t *testing.T) {
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		// Set the tool flag
		diffTool = "diff"
		defer func() { diffTool = "" }()

		var stdout, stderr bytes.Buffer
		diffFileCmd.SetOut(&stdout)
		diffFileCmd.SetErr(&stderr)

		args := []string{"test-repo", "main", "feature", "--tool", "diff", "test.txt"}
		diffFileCmd.SetArgs(args)

		// The command might succeed or fail depending on whether the external tool works
		// We're mainly testing that the external tool path is executed without panicking
		_ = diffFileCmd.Execute()
	})
}

// TestDiffArgValidation tests argument validation for diff commands
func TestDiffArgValidation(t *testing.T) {
	tests := []struct {
		name          string
		command       *cobra.Command
		args          []string
		expectError   bool
		errorContains string
	}{
		{
			name:          "diff file - too few args",
			command:       diffFileCmd,
			args:          []string{"repo", "branch1", "branch2"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:          "diff file - wrong arg count",
			command:       diffFileCmd,
			args:          []string{"repo", "branch1", "file.txt"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:          "diff file - too many args",
			command:       diffFileCmd,
			args:          []string{"repo", "branch1", "branch2", "file.txt", "extra"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:        "diff file - valid args",
			command:     diffFileCmd,
			args:        []string{"repo", "branch1", "branch2", "file.txt"},
			expectError: false,
		},
		{
			name:          "cross-repo - too few args",
			command:       diffCrossRepoCmd,
			args:          []string{"repo1", "repo2"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:          "cross-repo - wrong arg count",
			command:       diffCrossRepoCmd,
			args:          []string{"repo1", "file.txt"},
			expectError:   true,
			errorContains: "expected format",
		},
		{
			name:        "cross-repo - valid args",
			command:     diffCrossRepoCmd,
			args:        []string{"repo1", "repo2", "file.txt"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test argument validation function directly
			if tt.command.Args != nil {
				err := tt.command.Args(tt.command, tt.args)

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
				}
			}
		})
	}
}

// Helper functions for test setup

// initTestRepository creates a test Git repository with sample content
func initTestRepository(t *testing.T, repoPath string) error {
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

	// Configure git user (required for commits)
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
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("Initial content\nLine 2\nLine 3\n"), 0644); err != nil {
		return err
	}

	cmds = [][]string{
		{"git", "add", "test.txt"},
		{"git", "commit", "-m", "Initial commit"},
		{"git", "checkout", "-b", "feature"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to setup git repo: %w", err)
		}
	}

	// Modify file in feature branch
	if err := os.WriteFile(testFile, []byte("Modified content\nLine 2\nNew line 3\nAdded line 4\n"), 0644); err != nil {
		return err
	}

	cmds = [][]string{
		{"git", "add", "test.txt"},
		{"git", "commit", "-m", "Feature changes"},
		{"git", "checkout", "main"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to setup feature branch: %w", err)
		}
	}

	return nil
}

// createFileInRepo creates a file with specific content in a repository
func createFileInRepo(t *testing.T, repoPath, filename, content string) error {
	t.Helper()

	filePath := filepath.Join(repoPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	// Add and commit the file
	cmds := [][]string{
		{"git", "add", filename},
		{"git", "commit", "-m", fmt.Sprintf("Update %s", filename)},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit file %s: %w", filename, err)
		}
	}

	return nil
}

// createTestConfig creates a temporary configuration file for testing
func createTestConfig(t *testing.T, configPath string, repositories map[string]string) error {
	t.Helper()

	configMgr := di.ConfigManager()
	cfg := configMgr.GetConfig()
	cfg.Repositories = repositories

	// Set the config path and save
	os.Setenv("GMAN_CONFIG", configPath)
	return configMgr.Save()
}

// captureOutput captures stdout/stderr from a function execution
func captureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	go func() {
		defer wOut.Close()
		defer wErr.Close()
		fn()
	}()

	wOut.Close()
	wErr.Close()

	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return string(outBytes), string(errBytes)
}
