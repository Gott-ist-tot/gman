package external

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gman/internal/di"
)

// FileResult represents a file search result from fd
type FileResult struct {
	RepoAlias   string
	RelativePath string
	FullPath    string
	DisplayText string
}

// FDSearcher performs file searches using the fd tool
type FDSearcher struct {
	timeout time.Duration
}

// NewFDSearcher creates a new fd-based file searcher
func NewFDSearcher() *FDSearcher {
	return &FDSearcher{
		timeout: 10 * time.Second,
	}
}

// SearchFiles searches for files across multiple repositories using fd
func (fs *FDSearcher) SearchFiles(pattern string, repositories map[string]string, groupFilter string) ([]FileResult, error) {
	if !FD.IsAvailable() {
		return nil, fmt.Errorf("fd not available: %s", FD.GetInstallInstructions())
	}

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
		return results, fmt.Errorf("search timed out after %v", fs.timeout)
	}

	return results, nil
}

// searchInRepository searches for files in a single repository
func (fs *FDSearcher) searchInRepository(ctx context.Context, alias, repoPath, pattern string) ([]FileResult, error) {
	// Construct fd command
	args := []string{
		"--type", "f",        // files only
		"--hidden",           // include hidden files
		"--follow",           // follow symlinks
		"--exclude", ".git",  // exclude .git directories
		"--color", "never",   // no color output
	}

	// Add pattern if provided
	if pattern != "" {
		args = append(args, pattern)
	}

	// Add search path
	args = append(args, repoPath)

	cmd := exec.CommandContext(ctx, "fd", args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fd command: %w", err)
	}

	var results []FileResult
	scanner := bufio.NewScanner(stdout)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Calculate relative path from repository root
		relPath, err := filepath.Rel(repoPath, line)
		if err != nil {
			// If we can't get relative path, use the full path
			relPath = line
		}

		// Create display text: "alias:path"
		displayText := fmt.Sprintf("%s:%s", alias, relPath)

		results = append(results, FileResult{
			RepoAlias:    alias,
			RelativePath: relPath,
			FullPath:     line,
			DisplayText:  displayText,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading fd output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// fd returns non-zero exit code when no matches found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No matches found, return empty results
			return results, nil
		}
		return nil, fmt.Errorf("fd command failed: %w", err)
	}

	return results, nil
}

// FormatForFZF formats file results for fzf input
// Format: "absolute_path:0:display_text" (line number 0 for files, consistent with content search)
func (fs *FDSearcher) FormatForFZF(results []FileResult) string {
	var lines []string
	for _, result := range results {
		// Format: absolute_path:0:display_text
		// Using 0 as line number for files (consistent with content search format)
		line := fmt.Sprintf("%s:0:%s", result.FullPath, result.DisplayText)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ParseFZFSelection parses fzf selection and returns the corresponding file result
func (fs *FDSearcher) ParseFZFSelection(selection string, results []FileResult) (*FileResult, error) {
	// New format is: "absolute_path:0:display_text"
	// We need to extract the display_text part and match it
	parts := strings.SplitN(selection, ":", 3)
	if len(parts) < 3 {
		// Fallback: try to match the entire selection as display text
		for _, result := range results {
			if result.DisplayText == selection {
				return &result, nil
			}
		}
		return nil, fmt.Errorf("selection not found in results")
	}

	// Extract display text (everything after the second colon)
	displayText := parts[2]
	
	for _, result := range results {
		if result.DisplayText == displayText {
			return &result, nil
		}
	}
	return nil, fmt.Errorf("selection not found in results")
}