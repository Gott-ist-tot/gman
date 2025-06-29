package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	cmdutils "gman/internal/cmd"
	"gman/internal/di"
)

// BenchmarkStatusChecking benchmarks repository status checking performance
func BenchmarkStatusChecking(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "gman_benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test repositories
	numRepos := 10
	repos := make(map[string]string)
	for i := 0; i < numRepos; i++ {
		alias := "bench-repo-" + string(rune('0'+i))
		repoPath := filepath.Join(tempDir, alias)
		repos[alias] = repoPath

		if err := createSimpleBenchmarkRepo(b, repoPath, alias); err != nil {
			b.Fatalf("Failed to initialize repo %s: %v", alias, err)
		}
	}

	mgrs := cmdutils.GetManagers()
	gitMgr := mgrs.Git
	
	// Reset timer before benchmark
	b.ResetTimer()

	// Benchmark status checking
	for i := 0; i < b.N; i++ {
		for alias, path := range repos {
			status := gitMgr.GetRepoStatus(alias, path)
			if status.Error != nil {
				b.Errorf("Status check failed for %s: %v", alias, status.Error)
			}
		}
	}
}

// BenchmarkConfigLoading benchmarks configuration loading performance
func BenchmarkConfigLoading(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "gman_config_benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yml")
	
	// Create a large configuration
	configContent := `repositories:
`
	for i := 0; i < 100; i++ {
		configContent += "  repo-" + string(rune('0'+(i%10))) + string(rune('0'+(i/10))) + ": /path/to/repo-" + string(rune('0'+(i%10))) + string(rune('0'+(i/10))) + "\n"
	}
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset DI container for clean state
		di.Reset()
		
		os.Setenv("GMAN_CONFIG", configPath)
		mgrs := cmdutils.GetManagers()
		err := mgrs.Config.Load()
		os.Unsetenv("GMAN_CONFIG")
		
		if err != nil {
			b.Errorf("Config loading failed: %v", err)
		}
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent git operations
func BenchmarkConcurrentOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "gman_concurrent_benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test repositories
	numRepos := 5
	repos := make(map[string]string)
	for i := 0; i < numRepos; i++ {
		alias := "concurrent-repo-" + string(rune('0'+i))
		repoPath := filepath.Join(tempDir, alias)
		repos[alias] = repoPath

		if err := createSimpleBenchmarkRepo(b, repoPath, alias); err != nil {
			b.Fatalf("Failed to initialize repo %s: %v", alias, err)
		}
	}

	mgrs := cmdutils.GetManagers()
	gitMgr := mgrs.Git
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Channel to collect results
		resultsChan := make(chan error, len(repos))
		
		// Start concurrent operations
		for alias, path := range repos {
			go func(a, p string) {
				status := gitMgr.GetRepoStatus(a, p)
				resultsChan <- status.Error
			}(alias, path)
		}
		
		// Wait for all operations to complete
		for j := 0; j < len(repos); j++ {
			if err := <-resultsChan; err != nil {
				b.Errorf("Concurrent operation failed: %v", err)
			}
		}
	}
}

// Helper function for simple benchmark repositories
func createSimpleBenchmarkRepo(b testing.TB, repoPath, alias string) error {
	b.Helper()
	
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	// Initialize git repository quickly for benchmarking
	if err := runBenchmarkGitCommand(repoPath, "init"); err != nil {
		return err
	}
	
	if err := runBenchmarkGitCommand(repoPath, "config", "user.name", "Benchmark User"); err != nil {
		return err
	}
	
	if err := runBenchmarkGitCommand(repoPath, "config", "user.email", "benchmark@test.com"); err != nil {
		return err
	}

	// Create a simple file
	readmePath := filepath.Join(repoPath, "README.md")
	content := "# " + alias + "\nBenchmark repository"
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return err
	}

	// Commit
	if err := runBenchmarkGitCommand(repoPath, "add", "."); err != nil {
		return err
	}
	
	if err := runBenchmarkGitCommand(repoPath, "commit", "-m", "Initial commit"); err != nil {
		return err
	}

	return nil
}

// Helper function to run git commands for benchmarks
func runBenchmarkGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}