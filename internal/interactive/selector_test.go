package interactive

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"gman/pkg/types"
)

// TestRepositorySelector_SelectRepository tests the basic repository selection functionality
func TestRepositorySelector_SelectRepository(t *testing.T) {
	tests := []struct {
		name          string
		repos         map[string]string
		input         string
		expectError   bool
		errorContains string
		expectedRepo  string
		description   string
	}{
		{
			name: "numeric selection",
			repos: map[string]string{
				"repo1": "/path/to/repo1",
				"repo2": "/path/to/repo2",
			},
			input:        "1\n",
			expectError:  false,
			expectedRepo: "repo1", // First in alphabetical order
			description:  "Should select first repository with numeric input",
		},
		{
			name: "exact alias match",
			repos: map[string]string{
				"backend":  "/path/to/backend",
				"frontend": "/path/to/frontend",
			},
			input:        "backend\n",
			expectError:  false,
			expectedRepo: "backend",
			description:  "Should select repository by exact alias",
		},
		{
			name: "fuzzy match single result",
			repos: map[string]string{
				"my-awesome-backend":  "/path/to/backend",
				"my-awesome-frontend": "/path/to/frontend",
			},
			input:        "back\n",
			expectError:  false,
			expectedRepo: "my-awesome-backend",
			description:  "Should fuzzy match and select unique result",
		},
		{
			name: "fuzzy match multiple results",
			repos: map[string]string{
				"backend-api":    "/path/to/api",
				"backend-worker": "/path/to/worker",
				"frontend-web":   "/path/to/web",
			},
			input:         "backend\n",
			expectError:   true,
			errorContains: "ambiguous selection",
			description:   "Should fail with multiple fuzzy matches",
		},
		{
			name: "invalid numeric selection",
			repos: map[string]string{
				"repo1": "/path/to/repo1",
			},
			input:         "99\n",
			expectError:   true,
			errorContains: "invalid selection number",
			description:   "Should fail with out-of-range numeric selection",
		},
		{
			name: "no match found",
			repos: map[string]string{
				"repo1": "/path/to/repo1",
			},
			input:         "nonexistent\n",
			expectError:   true,
			errorContains: "not found",
			description:   "Should fail when no repository matches input",
		},
		{
			name:          "empty input",
			repos:         map[string]string{"repo1": "/path/to/repo1"},
			input:         "\n",
			expectError:   true,
			errorContains: "no selection made",
			description:   "Should fail with empty input",
		},
		{
			name:          "empty repository list",
			repos:         map[string]string{},
			input:         "anything\n",
			expectError:   true,
			errorContains: "no repositories configured",
			description:   "Should fail with empty repository list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate user input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			selector := NewRepositorySelector(tt.repos)
			result, err := selector.SelectRepository()

			// Restore stdin
			os.Stdin = oldStdin
			r.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for: %s", tt.description)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}
				if result != tt.expectedRepo {
					t.Errorf("Expected repo '%s', got '%s' for: %s", tt.expectedRepo, result, tt.description)
				}
			}
		})
	}
}

// TestRepositorySelector_FuzzyMatch tests the fuzzy matching functionality in isolation
func TestRepositorySelector_FuzzyMatch(t *testing.T) {
	repos := map[string]string{
		"backend-api":      "/path/to/api",
		"backend-worker":   "/path/to/worker",
		"backend-database": "/path/to/db",
		"frontend-web":     "/path/to/web",
		"frontend-mobile":  "/path/to/mobile",
		"shared-utils":     "/path/to/utils",
	}

	selector := NewRepositorySelector(repos)

	tests := []struct {
		input       string
		expected    []string
		description string
	}{
		{
			input:       "backend",
			expected:    []string{"backend-api", "backend-worker", "backend-database"},
			description: "Should match all backend repositories",
		},
		{
			input:       "front",
			expected:    []string{"frontend-web", "frontend-mobile"},
			description: "Should match all frontend repositories",
		},
		{
			input:       "api",
			expected:    []string{"backend-api"},
			description: "Should match single repository",
		},
		{
			input:       "nonexistent",
			expected:    []string{},
			description: "Should return empty for no matches",
		},
		{
			input:       "",
			expected:    []string{"backend-api", "backend-worker", "backend-database", "frontend-web", "frontend-mobile", "shared-utils"}, // Empty string matches all
			description: "Empty input matches all (contains behavior)",
		},
		{
			input:       "BACKEND",
			expected:    []string{"backend-api", "backend-worker", "backend-database"},
			description: "Should be case insensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Convert repos map to aliases slice for fuzzyMatch
			var aliases []string
			for alias := range repos {
				aliases = append(aliases, alias)
			}

			// Sort for consistent testing
			for i := 0; i < len(aliases)-1; i++ {
				for j := i + 1; j < len(aliases); j++ {
					if aliases[i] > aliases[j] {
						aliases[i], aliases[j] = aliases[j], aliases[i]
					}
				}
			}

			matches := selector.fuzzyMatch(tt.input, aliases)

			if len(matches) != len(tt.expected) {
				t.Errorf("Expected %d matches, got %d for input '%s'", len(tt.expected), len(matches), tt.input)
				return
			}

			// Check that all expected matches are present (order might differ)
			for _, expectedMatch := range tt.expected {
				found := false
				for _, actualMatch := range matches {
					if actualMatch == expectedMatch {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected match '%s' not found in results for input '%s'", expectedMatch, tt.input)
				}
			}
		})
	}
}

// TestSwitchTargetSelector_SelectTarget tests the enhanced switch target selection
func TestSwitchTargetSelector_SelectTarget(t *testing.T) {
	targets := []types.SwitchTarget{
		{
			Alias:     "backend",
			Path:      "/path/to/backend",
			Type:      "repository",
			RepoAlias: "backend",
		},
		{
			Alias:     "frontend",
			Path:      "/path/to/frontend",
			Type:      "repository",
			RepoAlias: "frontend",
		},
		{
			Alias:       "backend-feature",
			Path:        "/path/to/backend-feature",
			Type:        "worktree",
			RepoAlias:   "backend",
			Branch:      "feature-xyz",
			Description: "Worktree of backend",
		},
		{
			Alias:       "backend-hotfix",
			Path:        "/path/to/backend-hotfix",
			Type:        "worktree",
			RepoAlias:   "backend",
			Branch:      "hotfix-123",
			Description: "Worktree of backend",
		},
	}

	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
		expectedAlias string
		expectedType  string
		description   string
	}{
		{
			name:          "select repository by number",
			input:         "1\n",
			expectError:   false,
			expectedAlias: "backend", // Should be first after sorting (repos before worktrees)
			expectedType:  "repository",
			description:   "Should select repository by numeric input",
		},
		{
			name:          "select worktree by number",
			input:         "3\n", // After sorting: backend(1), frontend(2), backend-feature(3), backend-hotfix(4)
			expectError:   false,
			expectedAlias: "backend-feature",
			expectedType:  "worktree",
			description:   "Should select worktree by numeric input",
		},
		{
			name:          "select by exact alias",
			input:         "backend-hotfix\n",
			expectError:   false,
			expectedAlias: "backend-hotfix",
			expectedType:  "worktree",
			description:   "Should select target by exact alias match",
		},
		{
			name:          "fuzzy match unique result",
			input:         "hotfix\n",
			expectError:   false,
			expectedAlias: "backend-hotfix",
			expectedType:  "worktree",
			description:   "Should fuzzy match and select unique worktree",
		},
		{
			name:          "fuzzy match multiple results",
			input:         "back\n",
			expectError:   true,
			errorContains: "ambiguous selection",
			description:   "Should fail with multiple matches (backend, backend-feature, backend-hotfix)",
		},
		{
			name:          "invalid numeric selection",
			input:         "99\n",
			expectError:   true,
			errorContains: "invalid selection number",
			description:   "Should fail with out-of-range selection",
		},
		{
			name:          "no match found",
			input:         "nonexistent\n",
			expectError:   true,
			errorContains: "not found",
			description:   "Should fail when no target matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate user input
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			selector := NewSwitchTargetSelector(targets)
			result, err := selector.SelectTarget()

			// Restore stdin
			os.Stdin = oldStdin
			r.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for: %s", tt.description)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil for: %s", tt.description)
					return
				}
				if result.Alias != tt.expectedAlias {
					t.Errorf("Expected alias '%s', got '%s' for: %s", tt.expectedAlias, result.Alias, tt.description)
				}
				if result.Type != tt.expectedType {
					t.Errorf("Expected type '%s', got '%s' for: %s", tt.expectedType, result.Type, tt.description)
				}
			}
		})
	}
}

// TestSwitchTargetSelector_Sorting tests the sorting behavior of switch targets
func TestSwitchTargetSelector_Sorting(t *testing.T) {
	targets := []types.SwitchTarget{
		{Alias: "z-worktree", Type: "worktree", RepoAlias: "z-repo"},
		{Alias: "a-repo", Type: "repository", RepoAlias: "a-repo"},
		{Alias: "m-worktree", Type: "worktree", RepoAlias: "m-repo"},
		{Alias: "b-repo", Type: "repository", RepoAlias: "b-repo"},
		{Alias: "a-worktree", Type: "worktree", RepoAlias: "a-repo"},
	}

	// Simulate selection to trigger sorting
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		w.Write([]byte("1\n")) // Select first item
	}()

	selector := NewSwitchTargetSelector(targets)
	result, err := selector.SelectTarget()

	os.Stdin = oldStdin
	r.Close()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// First item after sorting should be a repository (repositories come first)
	if result.Type != "repository" {
		t.Errorf("Expected first item to be a repository, got: %s", result.Type)
	}

	// Should be the alphabetically first repository (a-repo)
	if result.Alias != "a-repo" {
		t.Errorf("Expected first repository to be 'a-repo', got: %s", result.Alias)
	}
}

// TestSwitchTargetSelector_EmptyTargets tests behavior with empty target list
func TestSwitchTargetSelector_EmptyTargets(t *testing.T) {
	selector := NewSwitchTargetSelector([]types.SwitchTarget{})

	result, err := selector.SelectTarget()

	if err == nil {
		t.Error("Expected error with empty targets")
	}
	if result != nil {
		t.Error("Expected nil result with empty targets")
	}
	if !strings.Contains(err.Error(), "no repositories or worktrees available") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestInteractiveOutputCapture tests that interactive output is properly formatted
func TestInteractiveOutputCapture(t *testing.T) {
	repos := map[string]string{
		"test-repo": "/very/long/path/to/repository/that/exceeds/normal/display/width/test-repo",
		"short":     "/short/path",
	}

	// Capture output by redirecting stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simulate user input
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR

	done := make(chan bool)
	var output string

	// Capture output in chunks to handle buffering
	go func() {
		defer func() { done <- true }()
		var totalOutput []byte
		buf := make([]byte, 512)
		
		for {
			n, err := r.Read(buf)
			if n > 0 {
				totalOutput = append(totalOutput, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		output = string(totalOutput)
	}()

	// Run selector in a goroutine to allow output capture
	go func() {
		// Provide input after a small delay to ensure display is rendered
		time.Sleep(10 * time.Millisecond)
		stdinW.Write([]byte("1\n"))
		stdinW.Close()
	}()

	selector := NewRepositorySelector(repos)
	_, _ = selector.SelectRepository() // We don't care about the result, just the output

	// Close output pipe to signal EOF to reader
	w.Close()
	os.Stdout = oldStdout
	stdinR.Close()
	os.Stdin = oldStdin

	<-done

	// Verify output contains expected elements
	t.Logf("Captured output: %q", output)
	if !strings.Contains(output, "Select a repository:") {
		t.Error("Expected selection prompt in output")
	}
	if !strings.Contains(output, "test-repo") {
		t.Error("Expected repository name in output")
	}
	if !strings.Contains(output, "...") {
		t.Error("Expected path truncation for long paths")
	}
}

// TestConcurrentSelectorAccess tests thread safety of selector methods
func TestConcurrentSelectorAccess(t *testing.T) {
	repos := map[string]string{
		"repo1": "/path/to/repo1",
		"repo2": "/path/to/repo2",
		"repo3": "/path/to/repo3",
	}

	selector := NewRepositorySelector(repos)
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	// Test concurrent access to selector's internal methods (fuzzyMatch)
	// This tests thread safety without relying on global state
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() {
				if r := recover(); r != nil {
					results <- fmt.Errorf("panic in goroutine %d: %v", index, r)
					return
				}
				results <- nil
			}()

			// Test concurrent access to internal state via fuzzyMatch
			aliases := []string{"repo1", "repo2", "repo3"}
			query := fmt.Sprintf("repo%d", (index%3)+1)
			
			// Call fuzzyMatch multiple times to stress test
			for j := 0; j < 100; j++ {
				matches := selector.fuzzyMatch(query, aliases)
				if len(matches) == 0 {
					results <- fmt.Errorf("unexpected empty matches for query '%s'", query)
					return
				}
			}
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	// With proper thread safety, no operations should fail
	if len(errors) > 0 {
		t.Errorf("Concurrent access test failed with %d errors:", len(errors))
		for _, err := range errors {
			t.Errorf("  - %v", err)
		}
	}

	t.Logf("Concurrent access test: %d/%d operations succeeded", numGoroutines-len(errors), numGoroutines)
}

// BenchmarkRepositorySelection benchmarks the selection performance
func BenchmarkRepositorySelection(b *testing.B) {
	// Create a large repository map
	repos := make(map[string]string)
	for i := 0; i < 1000; i++ {
		alias := fmt.Sprintf("repo-%d", i)
		path := fmt.Sprintf("/very/long/path/to/repository/number/%d/with/many/subdirectories", i)
		repos[alias] = path
	}

	selector := NewRepositorySelector(repos)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate input without actual I/O
		aliases := make([]string, 0, len(repos))
		for alias := range repos {
			aliases = append(aliases, alias)
		}

		// Test fuzzy matching performance
		matches := selector.fuzzyMatch("repo-5", aliases)
		if len(matches) == 0 {
			b.Error("Expected at least one match")
		}
	}
}

// TestRepositorySelector_LargeDataset tests performance with large number of repositories
func TestRepositorySelector_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	// Create a large repository map
	repos := make(map[string]string)
	for i := 0; i < 10000; i++ {
		alias := fmt.Sprintf("repository-%04d", i)
		path := fmt.Sprintf("/path/to/repository-%04d", i)
		repos[alias] = path
	}

	selector := NewRepositorySelector(repos)

	// Test fuzzy matching with large dataset
	tests := []struct {
		input   string
		maxTime int // milliseconds
	}{
		{"repo", 100}, // Should be fast even with many matches
		{"5000", 50},  // Specific match should be very fast
		{"xyz", 10},   // No matches should be fastest
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("search_%s", tt.input), func(t *testing.T) {
			// Convert to aliases slice
			var aliases []string
			for alias := range repos {
				aliases = append(aliases, alias)
			}

			start := time.Now()
			matches := selector.fuzzyMatch(tt.input, aliases)
			duration := time.Since(start)

			if duration.Milliseconds() > int64(tt.maxTime) {
				t.Errorf("Search took %v, expected less than %dms", duration, tt.maxTime)
			}

			t.Logf("Search for '%s' found %d matches in %v", tt.input, len(matches), duration)
		})
	}
}

// Helper function to create a test stdin reader
func createTestStdin(input string) (io.Reader, func()) {
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin

	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	return r, func() {
		os.Stdin = oldStdin
		r.Close()
	}
}
