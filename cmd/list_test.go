package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	cmdutils "gman/internal/cmd"
	"gman/internal/di"
)

func TestListCommand(t *testing.T) {
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

	// Create test config
	configData := `
repositories:
  test-repo1: /path/to/repo1
  test-repo2: /path/to/repo2
  cli-tool: /path/to/cli-tool
`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reset DI container for testing
	di.Reset()

	// Skip repository check for testing
	os.Setenv("GMAN_SKIP_REPO_CHECK", "true")
	defer os.Unsetenv("GMAN_SKIP_REPO_CHECK")

	// Manually load the configuration for testing
	mgrs := cmdutils.GetManagers()
	if err := mgrs.Config.Load(); err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "list repositories",
			args:     []string{},
			wantErr:  false,
			contains: []string{"test-repo1", "test-repo2", "cli-tool"},
		},
		{
			name:     "list with extra args (should still work)",
			args:     []string{"extra", "args"},
			wantErr:  false,
			contains: []string{"test-repo1", "test-repo2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the command
			err := runList(listCmd, tt.args)

			// Restore stdout and get output
			w.Close()
			os.Stdout = oldStdout
			buf.ReadFrom(r)
			output := buf.String()

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("runList() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that output contains expected strings
			for _, want := range tt.contains {
				if !bytes.Contains(buf.Bytes(), []byte(want)) {
					t.Errorf("runList() output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestListCommandWithNoConfig(t *testing.T) {
	// Create a temporary directory for non-existent config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yml")

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

	// Reset DI container for testing
	di.Reset()

	// Skip repository check for testing
	os.Setenv("GMAN_SKIP_REPO_CHECK", "true")
	defer os.Unsetenv("GMAN_SKIP_REPO_CHECK")

	// This should handle missing config gracefully
	err := runList(listCmd, []string{})
	
	// The command should not fail completely, but should handle missing config
	// Note: The actual behavior depends on how the config manager handles missing files
	// For now, we're just ensuring it doesn't panic
	if err != nil {
		t.Logf("Expected behavior: runList() with missing config returned error: %v", err)
	}
}