package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmdutils "gman/internal/cmd"
	"gman/internal/interactive"
	"gman/pkg/types"
	"gopkg.in/yaml.v3"
)

// TestFullWorkflow tests a complete workflow from repository setup to operations
func TestFullWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test scenario: Set up multiple repositories with different states
	repos := map[string]string{
		"project-backend":  filepath.Join(tempDir, "backend"),
		"project-frontend": filepath.Join(tempDir, "frontend"),
		"shared-utils":     filepath.Join(tempDir, "utils"),
	}

	// Initialize all repositories
	for alias, path := range repos {
		if err := initIntegrationTestRepo(t, path, alias); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}
	}

	// Create and setup configuration
	configPath := filepath.Join(tempDir, "config.yml")
	mgrs := cmdutils.GetManagers()
	cfg := mgrs.Config.GetConfig()
	cfg.Repositories = repos

	// Add some groups for testing
	cfg.Groups = map[string]types.Group{
		"project": {
			Name:         "project",
			Description:  "Main project repositories",
			Repositories: []string{"project-backend", "project-frontend"},
			CreatedAt:    time.Now(),
		},
		"all": {
			Name:         "all",
			Description:  "All repositories",
			Repositories: []string{"project-backend", "project-frontend", "shared-utils"},
			CreatedAt:    time.Now(),
		},
	}

	// Save config manually to the specified path  
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Set environment
	os.Setenv("GMAN_CONFIG", configPath)
	defer os.Unsetenv("GMAN_CONFIG")

	// Test workflow steps
	t.Run("1. Configuration Loading", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		reloadedMgr := mgrs.Config
		os.Setenv("GMAN_CONFIG", configPath)
		reloadedMgr.Load()
		reloadedCfg := reloadedMgr.GetConfig()

		if len(reloadedCfg.Repositories) != 3 {
			t.Errorf("Expected 3 repositories, got %d", len(reloadedCfg.Repositories))
		}

		if len(reloadedCfg.Groups) != 2 {
			t.Errorf("Expected 2 groups, got %d", len(reloadedCfg.Groups))
		}
	})

	t.Run("2. Git Operations", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git

		// Test status checking
		for alias, path := range repos {
			status := gitMgr.GetRepoStatus(alias, path)
			if status.Error != nil {
				t.Errorf("Failed to get status for %s: %v", alias, status.Error)
				continue
			}

			if status.Workspace != types.Clean {
				t.Errorf("Expected clean workspace for %s, got %s", alias, status.Workspace)
			}
		}
	})

	t.Run("3. Worktree Operations", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git
		backendPath := repos["project-backend"]

		// Create worktrees
		worktrees := []struct {
			path   string
			branch string
		}{
			{filepath.Join(tempDir, "backend-feature"), "feature-xyz"},
			{filepath.Join(tempDir, "backend-hotfix"), "hotfix-123"},
		}

		for _, wt := range worktrees {
			if err := gitMgr.AddWorktree(backendPath, wt.path, wt.branch); err != nil {
				t.Errorf("Failed to create worktree %s: %v", wt.path, err)
			}
		}

		// List worktrees
		wtList, err := gitMgr.ListWorktrees(backendPath)
		if err != nil {
			t.Errorf("Failed to list worktrees: %v", err)
		}

		// Should have main repo + 2 worktrees = 3 total
		if len(wtList) < 3 {
			t.Errorf("Expected at least 3 worktrees, got %d", len(wtList))
		}

		// Remove worktrees
		for _, wt := range worktrees {
			if err := gitMgr.RemoveWorktree(backendPath, wt.path, true); err != nil {
				t.Errorf("Failed to remove worktree %s: %v", wt.path, err)
			}
		}
	})

	t.Run("4. Diff Operations", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git
		backendPath := repos["project-backend"]

		// Create a feature branch with changes
		if err := createBranchWithChanges(t, backendPath, "diff-test", "Modified content for diff testing"); err != nil {
			t.Fatalf("Failed to create branch with changes: %v", err)
		}

		// Test diff between branches
		diff, err := gitMgr.DiffFileBetweenBranches(backendPath, "main", "diff-test", "README.md")
		if err != nil {
			t.Errorf("Failed to diff between branches: %v", err)
		}

		if diff == "" {
			t.Error("Expected diff output but got empty string")
		}

		// Test cross-repo diff
		frontendPath := repos["project-frontend"]
		diff, err = gitMgr.DiffFileBetweenRepos(backendPath, frontendPath, "README.md")
		if err != nil {
			t.Errorf("Failed to diff between repos: %v", err)
		}
	})

	t.Run("5. Interactive Components", func(t *testing.T) {
		// Test repository selector creation
		selector := interactive.NewRepositorySelector(repos)
		if selector == nil {
			t.Error("Failed to create repository selector")
		}

		// Test switch target collection
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")
		mgrs := cmdutils.GetManagers()
		err := mgrs.Config.Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// This would normally collect both repos and worktrees
		// We can't easily test the interactive part, but we can test the data structures
		targets := []types.SwitchTarget{
			{
				Alias:     "project-backend",
				Path:      repos["project-backend"],
				Type:      "repository",
				RepoAlias: "project-backend",
			},
			{
				Alias:     "project-frontend",
				Path:      repos["project-frontend"],
				Type:      "repository",
				RepoAlias: "project-frontend",
			},
		}

		targetSelector := interactive.NewSwitchTargetSelector(targets)
		if targetSelector == nil {
			t.Error("Failed to create switch target selector")
		}
	})

	t.Run("6. Cross-Package Integration", func(t *testing.T) {
		// Test that git manager and config manager work together
		os.Setenv("GMAN_CONFIG", configPath)
		defer os.Unsetenv("GMAN_CONFIG")
		mgrs := cmdutils.GetManagers()
		err := mgrs.Config.Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		cfg := mgrs.Config.GetConfig()

		gitMgr := mgrs.Git

		// Test operations across all configured repositories
		statusResults := make(map[string]*types.RepoStatus)
		for alias, path := range cfg.Repositories {
			status := gitMgr.GetRepoStatus(alias, path)
			if status.Error != nil {
				t.Errorf("Failed to get status for %s: %v", alias, status.Error)
				continue
			}
			statusResults[alias] = &status
		}

		if len(statusResults) != len(cfg.Repositories) {
			t.Errorf("Expected status for %d repos, got %d", len(cfg.Repositories), len(statusResults))
		}

		// Verify all repositories are in clean state
		for alias, status := range statusResults {
			if status.Workspace != types.Clean {
				t.Errorf("Repository %s should be clean, got %s", alias, status.Workspace)
			}
		}
	})
}

// TestConcurrentOperations tests operations across multiple repositories concurrently
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "gman_concurrent_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple repositories for concurrent testing
	numRepos := 10
	repos := make(map[string]string)
	for i := 0; i < numRepos; i++ {
		alias := fmt.Sprintf("repo-%02d", i)
		path := filepath.Join(tempDir, alias)
		repos[alias] = path

		if err := initIntegrationTestRepo(t, path, alias); err != nil {
			t.Fatalf("Failed to initialize %s: %v", alias, err)
		}
	}

	configPath := filepath.Join(tempDir, "config.yml")
	mgrs := cmdutils.GetManagers()
	cfg := mgrs.Config.GetConfig()
	cfg.Repositories = repos

	// Save config manually to the specified path  
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	t.Run("concurrent status checks", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git

		// Test concurrent status checking
		statusChan := make(chan error, len(repos))

		for alias, path := range repos {
			go func(a, p string) {
				status := gitMgr.GetRepoStatus("test", p)
				statusChan <- status.Error
			}(alias, path)
		}

		// Collect results
		var errors []error
		for i := 0; i < len(repos); i++ {
			if err := <-statusChan; err != nil {
				errors = append(errors, err)
			}
		}

		if len(errors) > 0 {
			t.Errorf("Got %d errors in concurrent status checks: %v", len(errors), errors[0])
		}
	})

	t.Run("concurrent worktree operations", func(t *testing.T) {
		mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git

		// Test concurrent worktree creation (limited to first 3 repos to avoid conflicts)
		testRepos := 3
		wtChan := make(chan error, testRepos)

		for i := 0; i < testRepos; i++ {
			alias := fmt.Sprintf("repo-%02d", i)
			repoPath := repos[alias]
			wtPath := filepath.Join(tempDir, fmt.Sprintf("wt-%02d", i))
			branch := fmt.Sprintf("concurrent-branch-%02d", i)

			go func(rp, wp, b string) {
				err := gitMgr.AddWorktree(rp, wp, b)
				wtChan <- err
			}(repoPath, wtPath, branch)
		}

		// Collect results
		var errors []error
		for i := 0; i < testRepos; i++ {
			if err := <-wtChan; err != nil {
				errors = append(errors, err)
			}
		}

		// Some concurrent worktree operations might fail due to Git locks, which is expected
		if len(errors) == testRepos {
			t.Error("All concurrent worktree operations failed, expected at least some to succeed")
		}

		t.Logf("Concurrent worktree test: %d/%d operations succeeded", testRepos-len(errors), testRepos)
	})
}

// TestErrorRecovery tests error recovery and resilience
func TestErrorRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gman_error_recovery_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgrs := cmdutils.GetManagers()
	gitMgr := mgrs.Git

	t.Run("non-existent repository", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		status := gitMgr.GetRepoStatus("test", nonExistentPath)
		if status.Error == nil {
			t.Error("Expected error for non-existent repository")
		}
	})

	t.Run("corrupted repository", func(t *testing.T) {
		corruptedPath := filepath.Join(tempDir, "corrupted")
		if err := os.MkdirAll(corruptedPath, 0755); err != nil {
			t.Fatalf("Failed to create corrupted repo dir: %v", err)
		}

		// Create a fake .git directory without proper git structure
		gitDir := filepath.Join(corruptedPath, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatalf("Failed to create fake git dir: %v", err)
		}

		status := gitMgr.GetRepoStatus("test", corruptedPath)
		if status.Error == nil {
			t.Error("Expected error for corrupted repository")
		}
	})

	t.Run("invalid configuration", func(t *testing.T) {
		invalidConfigPath := filepath.Join(tempDir, "invalid.yml")
		if err := os.WriteFile(invalidConfigPath, []byte("invalid: yaml: content:\n  - malformed"), 0644); err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		os.Setenv("GMAN_CONFIG", invalidConfigPath)
		defer os.Unsetenv("GMAN_CONFIG")
		mgrs := cmdutils.GetManagers()
		err := mgrs.Config.Load()
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}
	})
}

// TestPerformanceCharacteristics tests performance with various workloads
func TestPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "gman_performance_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with varying repository sizes
	testCases := []struct {
		name     string
		numRepos int
		numFiles int
		fileSize int
	}{
		{"small_scale", 5, 10, 1024},
		{"medium_scale", 20, 50, 10240},
		{"large_scale", 50, 100, 51200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repos := make(map[string]string)

			// Create repositories with specified characteristics
			for i := 0; i < tc.numRepos; i++ {
				alias := fmt.Sprintf("%s-repo-%02d", tc.name, i)
				path := filepath.Join(tempDir, alias)
				repos[alias] = path

				if err := createPerformanceTestRepo(t, path, tc.numFiles, tc.fileSize); err != nil {
					t.Fatalf("Failed to create performance test repo: %v", err)
				}
			}

			mgrs := cmdutils.GetManagers()
		gitMgr := mgrs.Git

			// Measure status checking performance
			start := time.Now()
			for _, path := range repos {
				status := gitMgr.GetRepoStatus("test", path)
				if status.Error != nil {
					t.Errorf("Status check failed: %v", status.Error)
				}
			}
			duration := time.Since(start)

			avgPerRepo := duration / time.Duration(tc.numRepos)
			t.Logf("%s: %d repos processed in %v (avg: %v per repo)",
				tc.name, tc.numRepos, duration, avgPerRepo)

			// Performance assertion: should not exceed reasonable time limits
			maxPerRepo := 500 * time.Millisecond
			if avgPerRepo > maxPerRepo {
				t.Errorf("Performance degraded: avg %v per repo exceeds limit %v", avgPerRepo, maxPerRepo)
			}
		})
	}
}

// Helper functions for integration testing

// initIntegrationTestRepo creates a test repository with realistic content
func initIntegrationTestRepo(t *testing.T, repoPath, alias string) error {
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
		{"git", "config", "user.name", "Integration Test User"},
		{"git", "config", "user.email", "integration@test.com"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure git: %w", err)
		}
	}

	// Create realistic project structure
	files := map[string]string{
		"README.md":  fmt.Sprintf("# %s\n\nThis is the %s repository for integration testing.\n\n## Features\n\n- Feature 1\n- Feature 2\n", alias, alias),
		"go.mod":     fmt.Sprintf("module %s\n\ngo 1.19\n", alias),
		"main.go":    fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello from %s\")\n}\n", alias),
		"Makefile":   "build:\n\tgo build -o bin/app .\n\ntest:\n\tgo test ./...\n\nclean:\n\trm -rf bin/\n",
		".gitignore": "bin/\n*.log\n.env\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Create subdirectories with files
	subdirs := []string{"cmd", "internal", "pkg"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(repoPath, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			return err
		}

		// Add a simple file in each subdirectory
		subFile := filepath.Join(subdirPath, "example.go")
		content := fmt.Sprintf("package %s\n\n// Example file in %s/%s\n", subdir, alias, subdir)
		if err := os.WriteFile(subFile, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Commit everything
	cmds = [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", fmt.Sprintf("Initial commit for %s", alias)},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit initial files: %w", err)
		}
	}

	return nil
}

// createBranchWithChanges creates a new branch with modified content
func createBranchWithChanges(t *testing.T, repoPath, branchName, newContent string) error {
	t.Helper()

	cmds := [][]string{
		{"git", "checkout", "-b", branchName},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Modify README.md
	readmePath := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(newContent), 0644); err != nil {
		return err
	}

	cmds = [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", fmt.Sprintf("Changes in %s branch", branchName)},
		{"git", "checkout", "main"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	return nil
}

// createPerformanceTestRepo creates a repository with specified file count and sizes
func createPerformanceTestRepo(t *testing.T, repoPath string, numFiles, fileSize int) error {
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
		{"git", "config", "user.name", "Perf Test User"},
		{"git", "config", "user.email", "perf@test.com"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to configure git: %w", err)
		}
	}

	// Create files with specified size
	content := strings.Repeat("a", fileSize)
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file_%04d.txt", i)
		filePath := filepath.Join(repoPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Commit all files
	cmds = [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "Performance test files"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit performance files: %w", err)
		}
	}

	return nil
}
