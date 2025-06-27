package external

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gman/internal/di"
)

// FallbackSearcher provides basic file search functionality when external tools are not available
type FallbackSearcher struct {
	timeout time.Duration
}

// NewFallbackSearcher creates a new fallback file searcher
func NewFallbackSearcher() *FallbackSearcher {
	return &FallbackSearcher{
		timeout: 30 * time.Second, // Longer timeout since this is slower
	}
}

// SearchFiles searches for files using basic Go file walking (fallback for fd)
func (fs *FallbackSearcher) SearchFiles(pattern string, repositories map[string]string, groupFilter string) ([]FileResult, error) {
	// Filter repositories by group if specified
	reposToSearch := repositories
	if groupFilter != "" {
		configMgr := di.ConfigManager()
		groupRepos, err := configMgr.GetGroupRepositories(groupFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get group repositories: %w", err)
		}
		if len(groupRepos) == 0 {
			return []FileResult{}, nil // Empty group, return empty results
		}
		reposToSearch = groupRepos
	}

	ctx, cancel := context.WithTimeout(context.Background(), fs.timeout)
	defer cancel()

	var results []FileResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Search each repository concurrently
	for alias, path := range reposToSearch {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			
			repoResults, err := fs.searchInRepository(ctx, alias, path, pattern)
			if err != nil {
				// Log error but continue with other repositories
				fmt.Printf("Warning: Failed to search %s: %v\n", alias, err)
				return
			}

			mu.Lock()
			results = append(results, repoResults...)
			mu.Unlock()
		}(alias, path)
	}

	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		return results, fmt.Errorf("search timed out after %v (using fallback search)", fs.timeout)
	}

	return results, nil
}

// searchInRepository walks the repository directory tree to find matching files
func (fs *FallbackSearcher) searchInRepository(ctx context.Context, alias, repoPath, pattern string) ([]FileResult, error) {
	var results []FileResult
	
	// Convert pattern to lowercase for case-insensitive matching
	lowerPattern := strings.ToLower(pattern)
	
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if err != nil {
			// Skip directories we can't read
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Skip directories and .git directories
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and directories (except if explicitly looking for them)
		if strings.HasPrefix(info.Name(), ".") && pattern != "" && !strings.HasPrefix(pattern, ".") {
			return nil
		}

		// Check if filename matches pattern (case-insensitive)
		filename := strings.ToLower(info.Name())
		if pattern == "" || strings.Contains(filename, lowerPattern) {
			// Calculate relative path from repository root
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				// If we can't get relative path, use the full path
				relPath = path
			}

			// Create display text: "alias:path"
			displayText := fmt.Sprintf("%s:%s", alias, relPath)

			results = append(results, FileResult{
				RepoAlias:    alias,
				RelativePath: relPath,
				FullPath:     path,
				DisplayText:  displayText,
			})
		}

		return nil
	})

	if err != nil && err != ctx.Err() {
		return nil, fmt.Errorf("error walking directory %s: %w", repoPath, err)
	}

	return results, nil
}

// FormatForFZF formats file results for fzf input (same interface as FDSearcher)
func (fs *FallbackSearcher) FormatForFZF(results []FileResult) string {
	var lines []string
	for _, result := range results {
		lines = append(lines, result.DisplayText)
	}
	return strings.Join(lines, "\n")
}

// ParseFZFSelection parses fzf selection and returns the corresponding file result
func (fs *FallbackSearcher) ParseFZFSelection(selection string, results []FileResult) (*FileResult, error) {
	for _, result := range results {
		if result.DisplayText == selection {
			return &result, nil
		}
	}
	return nil, fmt.Errorf("selection not found in results")
}

// BasicSelector provides a simple numbered selection interface when fzf is not available
type BasicSelector struct{}

// NewBasicSelector creates a new basic selector
func NewBasicSelector() *BasicSelector {
	return &BasicSelector{}
}

// SelectFromResults provides a basic numbered selection interface
func (bs *BasicSelector) SelectFromResults(results []FileResult, prompt string) (*FileResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	// Display results with numbers
	fmt.Printf("\n%s\n", prompt)
	fmt.Println(strings.Repeat("-", len(prompt)))
	
	const maxDisplayResults = 20
	displayCount := len(results)
	if displayCount > maxDisplayResults {
		displayCount = maxDisplayResults
	}
	
	for i := 0; i < displayCount; i++ {
		fmt.Printf("%3d. %s\n", i+1, results[i].DisplayText)
	}
	
	if len(results) > maxDisplayResults {
		fmt.Printf("     ... and %d more results\n", len(results)-maxDisplayResults)
	}
	
	// Get user selection
	fmt.Printf("\nEnter selection (1-%d) or 'q' to quit: ", displayCount)
	var input string
	fmt.Scanln(&input)
	
	if input == "q" || input == "quit" {
		return nil, fmt.Errorf("selection cancelled")
	}
	
	// Parse selection
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}
	
	if selection < 1 || selection > displayCount {
		return nil, fmt.Errorf("selection out of range: %d (valid range: 1-%d)", selection, displayCount)
	}
	
	return &results[selection-1], nil
}