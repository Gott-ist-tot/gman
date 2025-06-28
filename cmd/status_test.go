package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gman/internal/di"
)

func TestStatusCommand(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	// Set environment variable to use our temp config
	originalConfigPath := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	defer func() {
		if originalConfigPath != "" {
			os.Setenv("GMAN_CONFIG", originalConfigPath)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	}()

	// Create test Git repositories
	repo1Dir := filepath.Join(tempDir, "repo1")
	repo2Dir := filepath.Join(tempDir, "repo2")
	
	// Initialize test Git repositories with basic setup
	if err := createBasicTestRepo(repo1Dir); err != nil {
		t.Fatalf("Failed to create test repo1: %v", err)
	}
	if err := createBasicTestRepo(repo2Dir); err != nil {
		t.Fatalf("Failed to create test repo2: %v", err)
	}

	// Create test config with real paths
	configData := fmt.Sprintf(`
repositories:
  test-repo1: %s
  test-repo2: %s
`, repo1Dir, repo2Dir)
	
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reset DI container for testing
	di.Reset()

	// Skip repository check for testing
	os.Setenv("GMAN_SKIP_REPO_CHECK", "true")
	defer os.Unsetenv("GMAN_SKIP_REPO_CHECK")

	// Manually load the configuration for testing
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		flags       map[string]bool
		wantErr     bool
		contains    []string
		notContains []string
	}{
		{
			name:     "basic status",
			args:     []string{},
			flags:    map[string]bool{"extended": false},
			wantErr:  false,
			contains: []string{"test-repo1", "test-repo2"},
		},
		{
			name:     "extended status",
			args:     []string{},
			flags:    map[string]bool{"extended": true},
			wantErr:  false,
			contains: []string{"test-repo1", "test-repo2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			verboseStatus = false
			if ext, ok := tt.flags["extended"]; ok {
				verboseStatus = ext
			}

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the command
			err := runStatus(statusCmd, tt.args)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)
			output := buf.String()

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("runStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that output contains expected strings
			for _, want := range tt.contains {
				if !bytes.Contains(buf.Bytes(), []byte(want)) {
					t.Errorf("runStatus() output should contain %q, got: %s", want, output)
				}
			}

			// Check that output does not contain unexpected strings
			for _, notWant := range tt.notContains {
				if bytes.Contains(buf.Bytes(), []byte(notWant)) {
					t.Errorf("runStatus() output should not contain %q, got: %s", notWant, output)
				}
			}
		})
	}
}

func TestStatusCommandWithNoRepositories(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	// Set environment variable to use our temp config
	originalConfigPath := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	defer func() {
		if originalConfigPath != "" {
			os.Setenv("GMAN_CONFIG", originalConfigPath)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	}()

	// Create empty config
	configData := `repositories: {}`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reset DI container for testing
	di.Reset()

	// Skip repository check for testing
	os.Setenv("GMAN_SKIP_REPO_CHECK", "true")
	defer os.Unsetenv("GMAN_SKIP_REPO_CHECK")

	// Manually load the configuration for testing
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the command
	err := runStatus(statusCmd, []string{})

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// Should not error but should indicate no repositories
	if err != nil {
		t.Errorf("runStatus() should not error with empty config, got: %v", err)
	}

	expectedMsg := "No repositories configured"
	if !bytes.Contains(buf.Bytes(), []byte(expectedMsg)) {
		t.Errorf("runStatus() should indicate no repositories, got: %s", output)
	}
}

// createBasicTestRepo creates a basic Git repository for testing
func createBasicTestRepo(repoPath string) error {
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Initialize Git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Configure git user (required for commits)
	userCmd := exec.Command("git", "config", "user.name", "Test User")
	userCmd.Dir = repoPath
	if err := userCmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user: %w", err)
	}

	emailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	emailCmd.Dir = repoPath
	if err := emailCmd.Run(); err != nil {
		return fmt.Errorf("failed to set git email: %w", err)
	}

	// Create a test file and initial commit
	testFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repository\n"), 0644); err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoPath
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = repoPath
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}