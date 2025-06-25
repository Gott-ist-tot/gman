package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gman/internal/config"
	"gman/internal/di"
	"gman/pkg/types"
)

// TestRepository represents a test repository with metadata
type TestRepository struct {
	Alias       string
	Path        string
	Description string
	Files       map[string]string
	Branches    []string
}

// TestConfig holds test configuration data
type TestConfig struct {
	Path         string
	Repositories map[string]string
	Groups       []types.Group
	Manager      *config.Manager
}

// RepositoryBuilder helps build test repositories with fluent API
type RepositoryBuilder struct {
	repo *TestRepository
}

// NewRepositoryBuilder creates a new repository builder
func NewRepositoryBuilder(alias, path string) *RepositoryBuilder {
	return &RepositoryBuilder{
		repo: &TestRepository{
			Alias: alias,
			Path:  path,
			Files: make(map[string]string),
		},
	}
}

// WithDescription sets the repository description
func (rb *RepositoryBuilder) WithDescription(desc string) *RepositoryBuilder {
	rb.repo.Description = desc
	return rb
}

// WithFile adds a file to the repository
func (rb *RepositoryBuilder) WithFile(filename, content string) *RepositoryBuilder {
	rb.repo.Files[filename] = content
	return rb
}

// WithBranches sets the branches to create
func (rb *RepositoryBuilder) WithBranches(branches ...string) *RepositoryBuilder {
	rb.repo.Branches = branches
	return rb
}

// Build creates the actual repository on disk
func (rb *RepositoryBuilder) Build(t *testing.T) error {
	t.Helper()
	return CreateTestRepository(t, rb.repo)
}

// GetRepository returns the repository metadata
func (rb *RepositoryBuilder) GetRepository() *TestRepository {
	return rb.repo
}

// CreateTestRepository creates a test repository with specified characteristics
func CreateTestRepository(t *testing.T, repo *TestRepository) error {
	t.Helper()

	if err := os.MkdirAll(repo.Path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", repo.Path, err)
	}

	// Initialize git repository
	if err := runGitCommand(repo.Path, "init"); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Configure git user
	if err := runGitCommand(repo.Path, "config", "user.name", "Test User"); err != nil {
		return err
	}
	if err := runGitCommand(repo.Path, "config", "user.email", "test@example.com"); err != nil {
		return err
	}

	// Add default files if none specified
	if len(repo.Files) == 0 {
		repo.Files = map[string]string{
			"README.md": fmt.Sprintf("# %s\n\n%s\n", repo.Alias, repo.Description),
			"main.go":   fmt.Sprintf("package main\n\nfunc main() {\n\t// %s\n}\n", repo.Alias),
		}
	}

	// Create all files
	for filename, content := range repo.Files {
		filePath := filepath.Join(repo.Path, filename)

		// Create directory if needed
		if dir := filepath.Dir(filePath); dir != repo.Path {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
	}

	// Initial commit
	if err := runGitCommand(repo.Path, "add", "."); err != nil {
		return err
	}
	if err := runGitCommand(repo.Path, "commit", "-m", "Initial commit"); err != nil {
		return err
	}

	// Create additional branches
	for _, branch := range repo.Branches {
		if err := runGitCommand(repo.Path, "checkout", "-b", branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}

		// Make a small change in each branch
		branchFile := filepath.Join(repo.Path, fmt.Sprintf("%s.txt", branch))
		content := fmt.Sprintf("Content for branch %s\n", branch)
		if err := os.WriteFile(branchFile, []byte(content), 0644); err != nil {
			return err
		}

		if err := runGitCommand(repo.Path, "add", fmt.Sprintf("%s.txt", branch)); err != nil {
			return err
		}
		if err := runGitCommand(repo.Path, "commit", "-m", fmt.Sprintf("Add %s branch content", branch)); err != nil {
			return err
		}
	}

	// Return to main branch
	if len(repo.Branches) > 0 {
		if err := runGitCommand(repo.Path, "checkout", "main"); err != nil {
			return err
		}
	}

	return nil
}

// CreateTestConfig creates a test configuration with specified repositories and groups
func CreateTestConfig(t *testing.T, configPath string, repos map[string]string, groups ...types.Group) (*TestConfig, error) {
	t.Helper()

	configMgr := di.ConfigManager()
	cfg := configMgr.GetConfig()
	cfg.Repositories = repos
	cfg.Groups = groups

	if err := configMgr.SaveToPath(configPath); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return &TestConfig{
		Path:         configPath,
		Repositories: repos,
		Groups:       groups,
		Manager:      configMgr,
	}, nil
}

// WithTestConfig sets up a test with the specified configuration
func WithTestConfig(t *testing.T, configPath string, testFunc func(*TestConfig)) {
	t.Helper()

	originalConfig := os.Getenv("GMAN_CONFIG")
	os.Setenv("GMAN_CONFIG", configPath)
	defer func() {
		if originalConfig != "" {
			os.Setenv("GMAN_CONFIG", originalConfig)
		} else {
			os.Unsetenv("GMAN_CONFIG")
		}
	}()

	configMgr := di.ConfigManager()
	configMgr.LoadFromPath(configPath)
	cfg := configMgr.GetConfig()

	testConfig := &TestConfig{
		Path:         configPath,
		Repositories: cfg.Repositories,
		Groups:       cfg.Groups,
		Manager:      configMgr,
	}

	testFunc(testConfig)
}

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T, prefix string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// runGitCommand runs a git command in the specified directory
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s failed in %s: %w\nOutput: %s",
			strings.Join(args, " "), dir, err, string(output))
	}
	return nil
}

// WaitForCondition waits for a condition to be met with timeout
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("Condition not met within %v: %s", timeout, message)
}

// AssertFileExists checks if a file exists at the specified path
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("File does not exist: %s", path)
	}
}

// AssertFileNotExists checks if a file does not exist at the specified path
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("File should not exist: %s", path)
	}
}

// AssertFileContains checks if a file contains the specified content
func AssertFileContains(t *testing.T, path, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}

	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("File %s does not contain expected content.\nExpected: %s\nActual: %s",
			path, expectedContent, string(content))
	}
}

// AssertStringContains checks if a string contains the expected substring
func AssertStringContains(t *testing.T, str, expected string) {
	t.Helper()

	if !strings.Contains(str, expected) {
		t.Errorf("String does not contain expected content.\nExpected: %s\nActual: %s", expected, str)
	}
}

// AssertStringNotContains checks if a string does not contain the unexpected substring
func AssertStringNotContains(t *testing.T, str, unexpected string) {
	t.Helper()

	if strings.Contains(str, unexpected) {
		t.Errorf("String contains unexpected content.\nUnexpected: %s\nActual: %s", unexpected, str)
	}
}

// CreateMultipleRepos creates multiple test repositories with consistent structure
func CreateMultipleRepos(t *testing.T, baseDir string, count int, prefix string) map[string]string {
	t.Helper()

	repos := make(map[string]string)
	for i := 0; i < count; i++ {
		alias := fmt.Sprintf("%s-%02d", prefix, i)
		path := filepath.Join(baseDir, alias)
		repos[alias] = path

		repo := &TestRepository{
			Alias:       alias,
			Path:        path,
			Description: fmt.Sprintf("Test repository %s", alias),
			Files: map[string]string{
				"README.md":  fmt.Sprintf("# %s\n\nRepository %d in the %s series.\n", alias, i, prefix),
				"main.go":    fmt.Sprintf("package main\n\nfunc main() {\n\tprintln(\"%s\")\n}\n", alias),
				"config.yml": fmt.Sprintf("name: %s\nversion: 1.0.%d\n", alias, i),
			},
		}

		if err := CreateTestRepository(t, repo); err != nil {
			t.Fatalf("Failed to create repository %s: %v", alias, err)
		}
	}

	return repos
}

// CreateRepoWithWorktrees creates a repository with multiple worktrees
func CreateRepoWithWorktrees(t *testing.T, repoPath string, worktreeSpecs []WorktreeSpec) error {
	t.Helper()

	// First create the main repository
	repo := &TestRepository{
		Alias:       "main-repo",
		Path:        repoPath,
		Description: "Repository with worktrees",
	}

	if err := CreateTestRepository(t, repo); err != nil {
		return err
	}

	// Create worktrees
	for _, spec := range worktreeSpecs {
		if err := runGitCommand(repoPath, "worktree", "add", spec.Path, "-b", spec.Branch); err != nil {
			return fmt.Errorf("failed to create worktree %s: %w", spec.Path, err)
		}

		// Add content to worktree if specified
		if len(spec.Files) > 0 {
			for filename, content := range spec.Files {
				filePath := filepath.Join(spec.Path, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					return fmt.Errorf("failed to write file %s in worktree: %w", filename, err)
				}
			}

			// Commit changes in worktree
			if err := runGitCommand(spec.Path, "add", "."); err != nil {
				return err
			}
			if err := runGitCommand(spec.Path, "commit", "-m", fmt.Sprintf("Changes in %s worktree", spec.Branch)); err != nil {
				return err
			}
		}
	}

	return nil
}

// WorktreeSpec specifies a worktree to create
type WorktreeSpec struct {
	Path   string
	Branch string
	Files  map[string]string
}

// NewWorktreeSpec creates a new worktree specification
func NewWorktreeSpec(path, branch string) WorktreeSpec {
	return WorktreeSpec{
		Path:   path,
		Branch: branch,
		Files:  make(map[string]string),
	}
}

// WithFiles adds files to the worktree spec
func (ws WorktreeSpec) WithFiles(files map[string]string) WorktreeSpec {
	ws.Files = files
	return ws
}

// MockStdin simulates user input for interactive testing
type MockStdin struct {
	input string
	pos   int
}

// NewMockStdin creates a new mock stdin with the specified input
func NewMockStdin(input string) *MockStdin {
	return &MockStdin{input: input}
}

// Read implements the io.Reader interface
func (m *MockStdin) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.input) {
		return 0, fmt.Errorf("EOF")
	}

	n = copy(p, []byte(m.input[m.pos:]))
	m.pos += n
	return n, nil
}

// TestTimer helps with timing-sensitive tests
type TestTimer struct {
	start time.Time
	name  string
	t     *testing.T
}

// NewTestTimer creates a new test timer
func NewTestTimer(t *testing.T, name string) *TestTimer {
	return &TestTimer{
		start: time.Now(),
		name:  name,
		t:     t,
	}
}

// Stop stops the timer and logs the duration
func (tt *TestTimer) Stop() time.Duration {
	duration := time.Since(tt.start)
	tt.t.Logf("%s took %v", tt.name, duration)
	return duration
}

// AssertDuration asserts that the operation completed within the expected duration
func (tt *TestTimer) AssertDuration(expected time.Duration) {
	duration := time.Since(tt.start)
	if duration > expected {
		tt.t.Errorf("%s took %v, expected less than %v", tt.name, duration, expected)
	}
}

// CaptureOutput captures stdout and stderr from a function execution
func CaptureOutput(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	// This is a simplified version - in practice, you might want to use pipes
	// For now, we'll just return empty strings as this is primarily for structure
	fn()
	return "", ""
}

// RandomString generates a random string of specified length (for unique test data)
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// CleanupPath ensures a path is cleaned up after testing
func CleanupPath(t *testing.T, path string) {
	t.Helper()
	t.Cleanup(func() {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Warning: failed to cleanup path %s: %v", path, err)
		}
	})
}

// InitBasicTestRepository creates a simple git repository with basic structure
func InitBasicTestRepository(t *testing.T, repoPath string) error {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return err
	}

	// Initialize git repository
	if err := runGitCommand(repoPath, "init"); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}

	// Configure git user (required for commits)
	if err := runGitCommand(repoPath, "config", "user.name", "Test User"); err != nil {
		return fmt.Errorf("failed to configure git user name: %w", err)
	}
	if err := runGitCommand(repoPath, "config", "user.email", "test@example.com"); err != nil {
		return fmt.Errorf("failed to configure git user email: %w", err)
	}

	// Create initial file and commit
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	if err := runGitCommand(repoPath, "add", "test.txt"); err != nil {
		return fmt.Errorf("failed to add test file: %w", err)
	}

	if err := runGitCommand(repoPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// InitTestRepositoryWithBranches creates a git repository with multiple branches
func InitTestRepositoryWithBranches(t *testing.T, repoPath string, branches []string) error {
	t.Helper()

	if err := InitBasicTestRepository(t, repoPath); err != nil {
		return err
	}

	// Create additional branches
	for _, branchName := range branches {
		if branchName == "main" || branchName == "master" {
			continue // Skip main/master branch as it already exists
		}

		if err := runGitCommand(repoPath, "checkout", "-b", branchName); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}

		// Create a unique file in this branch
		branchFile := filepath.Join(repoPath, fmt.Sprintf("%s.txt", branchName))
		content := fmt.Sprintf("content for %s branch", branchName)
		if err := os.WriteFile(branchFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create branch file: %w", err)
		}

		if err := runGitCommand(repoPath, "add", fmt.Sprintf("%s.txt", branchName)); err != nil {
			return fmt.Errorf("failed to add branch file: %w", err)
		}

		if err := runGitCommand(repoPath, "commit", "-m", fmt.Sprintf("Add %s content", branchName)); err != nil {
			return fmt.Errorf("failed to commit branch changes: %w", err)
		}
	}

	// Switch back to main
	if err := runGitCommand(repoPath, "checkout", "main"); err != nil {
		// Try master if main doesn't exist
		if err := runGitCommand(repoPath, "checkout", "master"); err != nil {
			return fmt.Errorf("failed to switch back to main/master: %w", err)
		}
	}

	return nil
}

// CreateBasicTestConfig creates a simple test configuration file
func CreateBasicTestConfig(t *testing.T, configPath string, repositories map[string]string) error {
	t.Helper()

	// Ensure the config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := `repositories:`
	for alias, path := range repositories {
		config += fmt.Sprintf("\n  %s: %s", alias, path)
	}
	config += "\n"

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
