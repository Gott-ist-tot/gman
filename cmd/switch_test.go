package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmdutils "gman/internal/cmd"
	"gman/test"
)

// TestSwitchCommand tests the switch command with various scenarios
func TestSwitchCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize test repository
	repoPath := filepath.Join(tempDir, "test-repo")
	if err := test.InitBasicTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create basic config
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
		outputCheck   func(output string) bool
	}{
		{
			name:         "switch to existing repository",
			args:         []string{"test-repo"},
			expectError:  false,
			expectOutput: true,
			outputCheck: func(output string) bool {
				return strings.Contains(output, "GMAN_CD:")
			},
		},
		{
			name:          "switch to nonexistent repository",
			args:          []string{"nonexistent"},
			expectError:   true,
			errorContains: "repository 'nonexistent' not found",
		},
		{
			name:         "switch with fuzzy matching",
			args:         []string{"test"},
			expectError:  false,
			expectOutput: true,
			outputCheck: func(output string) bool {
				return strings.Contains(output, "GMAN_CD:") && strings.Contains(output, repoPath)
			},
		},
		{
			name:         "list all targets (no args)",
			args:         []string{},
			expectError:  false,
			expectOutput: true,
			outputCheck: func(output string) bool {
				// When no args, should show interactive menu
				return strings.Contains(output, "Select a target:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			var stdout, stderr bytes.Buffer
			switchCmd.SetOut(&stdout)
			switchCmd.SetErr(&stderr)

			switchCmd.SetArgs(tt.args)
			err := switchCmd.Execute()

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

				if tt.expectOutput {
					output := stdout.String()
					if output == "" {
						t.Error("Expected output but got none")
					} else if tt.outputCheck != nil && !tt.outputCheck(output) {
						t.Errorf("Output check failed for output: %s", output)
					}
				}
			}
		})
	}
}

// TestSwitchWithWorktrees tests switch command integration with worktrees
func TestSwitchWithWorktrees(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_worktree_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repoPath := filepath.Join(tempDir, "test-repo")
	if err := test.InitBasicTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create worktrees for testing
	worktrees := []struct {
		path   string
		branch string
	}{
		{filepath.Join(tempDir, "feature-wt"), "feature-branch"},
		{filepath.Join(tempDir, "hotfix-wt"), "hotfix-branch"},
	}

	mgrs := cmdutils.GetManagers()
	gitManager := mgrs.Git
	for _, wt := range worktrees {
		if err := gitManager.AddWorktree(repoPath, wt.path, wt.branch); err != nil {
			t.Fatalf("Failed to create worktree %s: %v", wt.path, err)
		}
	}

	// Create basic config (simplified for testing)
	configPath := filepath.Join(tempDir, "config.yml")
	configContent := fmt.Sprintf("repositories:\n  test-repo: %s\n", repoPath)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("switch to worktree by path", func(t *testing.T) {
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		// Use the worktree path as argument
		args := []string{filepath.Join(tempDir, "feature-wt")}
		switchCmd.SetArgs(args)
		err := switchCmd.Execute()

		if err != nil {
			t.Errorf("Unexpected error switching to worktree: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "GMAN_CD:") {
			t.Error("Expected GMAN_CD output for worktree switch")
		}
		if !strings.Contains(output, "feature-wt") {
			t.Error("Expected worktree path in output")
		}
	})

	t.Run("switch to worktree by fuzzy match", func(t *testing.T) {
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		// Use fuzzy matching for worktree
		args := []string{"hotfix"}
		switchCmd.SetArgs(args)
		err := switchCmd.Execute()

		if err != nil {
			t.Errorf("Unexpected error switching to worktree by fuzzy match: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "GMAN_CD:") {
			t.Error("Expected GMAN_CD output for worktree fuzzy switch")
		}
		if !strings.Contains(output, "hotfix-wt") {
			t.Error("Expected hotfix worktree path in output")
		}
	})

	t.Run("ambiguous worktree match", func(t *testing.T) {
		// Create another worktree with similar name
		ambiguousWtPath := filepath.Join(tempDir, "feature-v2-wt")
		if err := gitManager.AddWorktree(repoPath, ambiguousWtPath, "feature-v2"); err != nil {
			t.Fatalf("Failed to create ambiguous worktree: %v", err)
		}

		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		// This should now be ambiguous
		args := []string{"feature"}
		switchCmd.SetArgs(args)
		err := switchCmd.Execute()

		if err == nil {
			t.Error("Expected error for ambiguous worktree match")
		}
		if !strings.Contains(err.Error(), "multiple matches") {
			t.Errorf("Expected ambiguous match error, got: %v", err)
		}
	})
}

// TestWorktreeAliasCollisionResolution tests the improved worktree alias generation
func TestWorktreeAliasCollisionResolution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_worktree_alias_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create two repositories that might have conflicting worktree names
	repos := map[string]string{
		"backend":  filepath.Join(tempDir, "backend"),
		"frontend": filepath.Join(tempDir, "frontend"),
	}

	mgrs := cmdutils.GetManagers()
	gitManager := mgrs.Git

	for alias, path := range repos {
		if err := test.InitBasicTestRepository(t, path); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}

		// Create worktrees with same names but in different repos
		hotfixPath := filepath.Join(tempDir, fmt.Sprintf("%s-hotfix", alias))
		if err := gitManager.AddWorktree(path, hotfixPath, "hotfix-branch"); err != nil {
			t.Fatalf("Failed to create hotfix worktree for %s: %v", alias, err)
		}
	}

	// Test the collectSwitchTargets function directly
	targets, err := collectSwitchTargets(repos)
	if err != nil {
		t.Fatalf("Failed to collect switch targets: %v", err)
	}

	// Verify that worktree aliases are prefixed with repo names
	var backendHotfixFound, frontendHotfixFound bool
	worktreeAliases := make(map[string]bool)

	for _, target := range targets {
		if target.Type == "worktree" {
			// Check for duplicate aliases
			if worktreeAliases[target.Alias] {
				t.Errorf("Duplicate worktree alias found: %s", target.Alias)
			}
			worktreeAliases[target.Alias] = true

			// Verify aliases are prefixed with repo names using "/" separator
			if strings.HasPrefix(target.Alias, "backend/") && strings.Contains(target.Alias, "hotfix") {
				backendHotfixFound = true
				if target.RepoAlias != "backend" {
					t.Errorf("Expected backend worktree to have RepoAlias 'backend', got '%s'", target.RepoAlias)
				}
			}
			if strings.HasPrefix(target.Alias, "frontend/") && strings.Contains(target.Alias, "hotfix") {
				frontendHotfixFound = true
				if target.RepoAlias != "frontend" {
					t.Errorf("Expected frontend worktree to have RepoAlias 'frontend', got '%s'", target.RepoAlias)
				}
			}

			// Verify description includes branch and path info for disambiguation
			if target.Branch != "" && !strings.Contains(target.Description, target.Branch) {
				t.Errorf("Expected description to include branch '%s', got '%s'", target.Branch, target.Description)
			}
			// Verify description includes path for disambiguation
			if !strings.Contains(target.Description, target.Path) {
				t.Errorf("Expected description to include path '%s' for disambiguation, got '%s'", target.Path, target.Description)
			}
		}
	}

	if !backendHotfixFound {
		t.Error("Expected to find backend hotfix worktree with proper alias")
	}
	if !frontendHotfixFound {
		t.Error("Expected to find frontend hotfix worktree with proper alias")
	}

	// Verify we can distinguish between the two hotfix worktrees using new format
	expectedAliases := []string{"backend/backend-hotfix", "frontend/frontend-hotfix"}
	for _, expectedAlias := range expectedAliases {
		found := false
		for _, target := range targets {
			if target.Type == "worktree" && target.Alias == expectedAlias {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find worktree with alias '%s'", expectedAlias)
		}
	}
}

// TestSwitchTargetCollection tests the target collection functionality
func TestSwitchTargetCollection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_targets_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple repositories and worktrees
	repos := map[string]string{
		"repo1": filepath.Join(tempDir, "repo1"),
		"repo2": filepath.Join(tempDir, "repo2"),
	}

	for alias, path := range repos {
		if err := test.InitBasicTestRepository(t, path); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}
	}

	// Create worktrees
	mgrs := cmdutils.GetManagers()
	gitManager := mgrs.Git
	worktreeInfo := []struct {
		repoPath string
		wtPath   string
		branch   string
	}{
		{repos["repo1"], filepath.Join(tempDir, "repo1-feature"), "feature-1"},
		{repos["repo1"], filepath.Join(tempDir, "repo1-hotfix"), "hotfix-1"},
		{repos["repo2"], filepath.Join(tempDir, "repo2-feature"), "feature-2"},
	}

	for _, wt := range worktreeInfo {
		if err := gitManager.AddWorktree(wt.repoPath, wt.wtPath, wt.branch); err != nil {
			t.Fatalf("Failed to create worktree %s: %v", wt.wtPath, err)
		}
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := test.CreateBasicTestConfig(t, configPath, repos); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("collect all targets", func(t *testing.T) {
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		// Test the collectSwitchTargets function indirectly by checking switch output
		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		// No args should trigger interactive mode showing all targets
		switchCmd.SetArgs([]string{})
		_ = switchCmd.Execute() // Ignore error as it may fail without input

		// The command should show the interactive menu (might fail due to no input, but we check output)
		output := stdout.String()

		// Should show repositories
		if !strings.Contains(output, "repo1") || !strings.Contains(output, "repo2") {
			t.Error("Expected repositories in target list")
		}

		// Should show worktrees (based on directory names)
		if !strings.Contains(output, "repo1-feature") || !strings.Contains(output, "repo2-feature") {
			t.Error("Expected worktrees in target list")
		}

		// Should show type indicators
		if !strings.Contains(output, "repo") || !strings.Contains(output, "worktree") {
			t.Error("Expected type indicators in target list")
		}
	})
}

// TestSwitchFuzzyMatching tests fuzzy matching behavior
func TestSwitchFuzzyMatching(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_fuzzy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repos := map[string]string{
		"backend-api":     filepath.Join(tempDir, "backend-api"),
		"backend-worker":  filepath.Join(tempDir, "backend-worker"),
		"frontend-web":    filepath.Join(tempDir, "frontend-web"),
		"frontend-mobile": filepath.Join(tempDir, "frontend-mobile"),
	}

	for alias, path := range repos {
		if err := test.InitBasicTestRepository(t, path); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := test.CreateBasicTestConfig(t, configPath, repos); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
		shouldMatch   string
	}{
		{
			name:        "unique fuzzy match",
			input:       "api",
			expectError: false,
			shouldMatch: "backend-api",
		},
		{
			name:          "ambiguous fuzzy match",
			input:         "backend",
			expectError:   true,
			errorContains: "multiple matches",
		},
		{
			name:        "unique partial match",
			input:       "mobile",
			expectError: false,
			shouldMatch: "frontend-mobile",
		},
		{
			name:          "no match",
			input:         "nonexistent",
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "empty input",
			input:       "",
			expectError: false, // Should trigger interactive mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GMAN_CONFIG", configPath)
			defer os.Unsetenv("GMAN_CONFIG")

			var stdout, stderr bytes.Buffer
			switchCmd.SetOut(&stdout)
			switchCmd.SetErr(&stderr)

			var args []string
			if tt.input != "" {
				args = []string{tt.input}
			}

			switchCmd.SetArgs(args)
			err := switchCmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil && tt.input != "" {
					t.Errorf("Unexpected error: %v", err)
				}

				if tt.shouldMatch != "" {
					output := stdout.String()
					if !strings.Contains(output, tt.shouldMatch) {
						t.Errorf("Expected output to contain '%s', got: %s", tt.shouldMatch, output)
					}
				}
			}
		})
	}
}

// TestSwitchPerformance tests switch command performance with many targets
func TestSwitchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "gman_switch_perf_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create many repositories
	numRepos := 100
	repos := make(map[string]string)
	for i := 0; i < numRepos; i++ {
		alias := fmt.Sprintf("repo-%03d", i)
		path := filepath.Join(tempDir, alias)
		repos[alias] = path

		if err := test.InitBasicTestRepository(t, path); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}
	}

	configPath := filepath.Join(tempDir, "config.yml")
	if err := test.CreateBasicTestConfig(t, configPath, repos); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("performance with many repositories", func(t *testing.T) {
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		// Test specific match
		args := []string{"repo-050"}
		switchCmd.SetArgs(args)

		err := switchCmd.Execute()
		if err != nil {
			t.Errorf("Performance test failed: %v", err)
		}

		output := stdout.String()
		if !strings.Contains(output, "repo-050") {
			t.Error("Performance test should find specific repository")
		}
	})
}

// TestSwitchErrorHandling tests error handling scenarios
func TestSwitchErrorHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_error_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		setupConfig   bool
		configContent map[string]string
		args          []string
		expectedError string
	}{
		{
			name:          "no config file",
			setupConfig:   false,
			args:          []string{"any-repo"},
			expectedError: "failed to load configuration",
		},
		{
			name:          "empty configuration",
			setupConfig:   true,
			configContent: map[string]string{},
			args:          []string{"any-repo"},
			expectedError: "no repositories configured",
		},
		{
			name:          "repository path doesn't exist",
			setupConfig:   true,
			configContent: map[string]string{"test": "/nonexistent/path"},
			args:          []string{"test"},
			expectedError: "", // Should still work, just switch to non-existent path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, fmt.Sprintf("config-%s.yml", tt.name))

			if tt.setupConfig {
				if err := test.CreateBasicTestConfig(t, configPath, tt.configContent); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
				os.Setenv("GMAN_CONFIG", configPath)
			} else {
				os.Setenv("GMAN_CONFIG", "/nonexistent/config.yml")
			}
			defer os.Unsetenv("GMAN_CONFIG")

			var stdout, stderr bytes.Buffer
			switchCmd.SetOut(&stdout)
			switchCmd.SetErr(&stderr)

			switchCmd.SetArgs(tt.args)
			err := switchCmd.Execute()

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s' but got none", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}


// TestSwitchShellIntegrationWarning tests that the shell integration warning is shown
func TestSwitchShellIntegrationWarning(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_switch_shell_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize test repository
	repoPath := filepath.Join(tempDir, "test-repo")
	if err := test.InitBasicTestRepository(t, repoPath); err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	// Create basic config
	configPath := filepath.Join(tempDir, "config.yml")
	if err := test.CreateBasicTestConfig(t, configPath, map[string]string{"test-repo": repoPath}); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("shell integration warning shown when not detected", func(t *testing.T) {
		// Clear shell integration environment variables
		os.Unsetenv("GMAN_SHELL_INTEGRATION")
		os.Unsetenv("GMAN_SKIP_SHELL_CHECK")
		defer func() {
			os.Unsetenv("GMAN_SHELL_INTEGRATION")
			os.Unsetenv("GMAN_SKIP_SHELL_CHECK")
		}()

		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		switchCmd.SetArgs([]string{"test-repo"})
		err := switchCmd.Execute()

		// Should get an error about shell integration
		if err == nil {
			t.Error("Expected error about shell integration but got none")
		}

		if !strings.Contains(err.Error(), "shell integration required") {
			t.Errorf("Expected shell integration error, got: %v", err)
		}

		// Check that warning message is displayed
		output := stdout.String()
		if !strings.Contains(output, "Shell integration not detected") {
			t.Error("Expected shell integration warning in output")
		}
		if !strings.Contains(output, "gman() {") {
			t.Error("Expected shell function example in output")
		}
	})

	t.Run("no warning when shell integration is active", func(t *testing.T) {
		// Set shell integration environment variable
		os.Setenv("GMAN_SHELL_INTEGRATION", "1")
		defer os.Unsetenv("GMAN_SHELL_INTEGRATION")

		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")

		var stdout, stderr bytes.Buffer
		switchCmd.SetOut(&stdout)
		switchCmd.SetErr(&stderr)

		switchCmd.SetArgs([]string{"test-repo"})
		err := switchCmd.Execute()

		// Should not get shell integration error
		if err != nil {
			t.Errorf("Unexpected error when shell integration is active: %v", err)
		}

		// Should get GMAN_CD output
		output := stdout.String()
		if !strings.Contains(output, "GMAN_CD:") {
			t.Error("Expected GMAN_CD output when shell integration is active")
		}
	})
}

// TestShellIntegrationDetection tests the improved shell integration detection
func TestShellIntegrationDetection(t *testing.T) {
	// Save original environment
	originalIntegration := os.Getenv("GMAN_SHELL_INTEGRATION")
	originalSkipCheck := os.Getenv("GMAN_SKIP_SHELL_CHECK")
	defer func() {
		if originalIntegration != "" {
			os.Setenv("GMAN_SHELL_INTEGRATION", originalIntegration)
		} else {
			os.Unsetenv("GMAN_SHELL_INTEGRATION")
		}
		if originalSkipCheck != "" {
			os.Setenv("GMAN_SKIP_SHELL_CHECK", originalSkipCheck)
		} else {
			os.Unsetenv("GMAN_SKIP_SHELL_CHECK")
		}
	}()

	tests := []struct {
		name                   string
		shellIntegrationEnv    string
		skipCheckEnv           string
		expectedActive         bool
		expectedDiagnosticsContain string
		description            string
	}{
		{
			name:                   "shell integration active",
			shellIntegrationEnv:    "1",
			skipCheckEnv:           "",
			expectedActive:         true,
			expectedDiagnosticsContain: "GMAN_SHELL_INTEGRATION properly set",
			description:            "Should detect active shell integration",
		},
		{
			name:                   "shell integration inactive",
			shellIntegrationEnv:    "",
			skipCheckEnv:           "",
			expectedActive:         false,
			expectedDiagnosticsContain: "GMAN_SHELL_INTEGRATION not set",
			description:            "Should detect inactive shell integration",
		},
		{
			name:                   "skip check bypass",
			shellIntegrationEnv:    "",
			skipCheckEnv:           "1",
			expectedActive:         true,
			expectedDiagnosticsContain: "Shell check bypass is active",
			description:            "Should honor skip check bypass",
		},
		{
			name:                   "shell integration overrides skip",
			shellIntegrationEnv:    "1",
			skipCheckEnv:           "1",
			expectedActive:         true,
			expectedDiagnosticsContain: "GMAN_SHELL_INTEGRATION properly set",
			description:            "Shell integration should be primary indicator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.shellIntegrationEnv != "" {
				os.Setenv("GMAN_SHELL_INTEGRATION", tt.shellIntegrationEnv)
			} else {
				os.Unsetenv("GMAN_SHELL_INTEGRATION")
			}
			
			if tt.skipCheckEnv != "" {
				os.Setenv("GMAN_SKIP_SHELL_CHECK", tt.skipCheckEnv)
			} else {
				os.Unsetenv("GMAN_SKIP_SHELL_CHECK")
			}

			// Test detection
			isActive := isShellIntegrationActive()
			if isActive != tt.expectedActive {
				t.Errorf("Expected isShellIntegrationActive() to return %v, got %v (%s)", 
					tt.expectedActive, isActive, tt.description)
			}

			// Test diagnostics
			diagnostics := getShellIntegrationDiagnostics()
			if !strings.Contains(diagnostics, tt.expectedDiagnosticsContain) {
				t.Errorf("Expected diagnostics to contain '%s', got: %s (%s)", 
					tt.expectedDiagnosticsContain, diagnostics, tt.description)
			}
		})
	}
}

// TestShellIntegrationDiagnostics tests the diagnostic information quality
func TestShellIntegrationDiagnostics(t *testing.T) {
	diagnostics := getShellIntegrationDiagnostics()
	
	// Should contain basic diagnostic elements
	expectedElements := []string{
		"GMAN_SHELL_INTEGRATION",
		"Shell:",
		"config file",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(diagnostics, element) {
			t.Errorf("Expected diagnostics to contain '%s', got: %s", element, diagnostics)
		}
	}
	
	// Should contain emojis for visual clarity
	expectedEmojis := []string{"❌", "📝", "💡"}
	for _, emoji := range expectedEmojis {
		if !strings.Contains(diagnostics, emoji) {
			t.Errorf("Expected diagnostics to contain emoji '%s' for visual clarity, got: %s", emoji, diagnostics)
		}
	}
}

// Helper functions for switch command testing

// initSwitchTestRepository creates a test repository suitable for switch operations
