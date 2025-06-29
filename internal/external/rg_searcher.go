package external

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gman/internal/di"
	"gman/internal/repository"
)

// ContentResult represents a content search result from rg
type ContentResult struct {
	RepoAlias    string
	FilePath     string
	FullPath     string // Absolute path to the file
	LineNumber   int
	LineContent  string
	MatchColumn  int
	DisplayText  string
}

// RGSearcher performs content searches using the rg (ripgrep) tool
type RGSearcher struct {
	timeout time.Duration
}

// NewRGSearcher creates a new rg-based content searcher
func NewRGSearcher() *RGSearcher {
	return &RGSearcher{
		timeout: 15 * time.Second, // Content search might take longer
	}
}

// SearchContent searches for content across multiple repositories using rg
func (rs *RGSearcher) SearchContent(pattern string, repositories map[string]string, groupFilter string) ([]ContentResult, error) {
	if !RipGrep.IsAvailable() {
		return nil, fmt.Errorf("rg not available: %s", RipGrep.GetInstallInstructions())
	}

	if pattern == "" {
		return nil, fmt.Errorf("search pattern is required for content search")
	}

	// Use consolidated repository filtering
	filter := repository.NewFilter(di.ConfigManager())
	reposToSearch, err := filter.FilterByGroup(repositories, groupFilter)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), rs.timeout)
	defer cancel()

	var results []ContentResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Search each repository concurrently
	for alias, path := range reposToSearch {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			
			repoResults, err := rs.searchInRepository(ctx, alias, path, pattern)
			if err != nil {
				// Log error but continue with other repositories
				fmt.Printf("Warning: Failed to search content in %s: %v\n", alias, err)
				return
			}

			mu.Lock()
			results = append(results, repoResults...)
			mu.Unlock()
		}(alias, path)
	}

	wg.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		return results, fmt.Errorf("content search timed out after %v", rs.timeout)
	}

	return results, nil
}

// searchInRepository searches for content in a single repository
func (rs *RGSearcher) searchInRepository(ctx context.Context, alias, repoPath, pattern string) ([]ContentResult, error) {
	// Construct rg command
	args := []string{
		"--line-number",      // show line numbers
		"--column",           // show column numbers
		"--no-heading",       // don't group by file
		"--with-filename",    // show filenames
		"--color", "never",   // no color output
		"--hidden",           // search hidden files
		"--follow",           // follow symlinks
		"--text",         // treat all files as text (skip binary detection)
		"-g", "!.git/**",     // exclude .git directory
		"--max-count", "50",  // limit matches per file to prevent overwhelming results
		pattern,              // search pattern
		repoPath,             // search path
	}

	cmd := exec.CommandContext(ctx, "rg", args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rg command: %w", err)
	}

	var results []ContentResult
	scanner := bufio.NewScanner(stdout)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse rg output: "file:line:column:content"
		result, err := rs.parseRGLine(alias, repoPath, line)
		if err != nil {
			// Skip malformed lines
			continue
		}

		results = append(results, *result)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rg output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// rg returns non-zero exit code when no matches found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No matches found, return empty results
			return results, nil
		}
		return nil, fmt.Errorf("rg command failed: %w", err)
	}

	return results, nil
}

// parseRGLine parses a single line of rg output
func (rs *RGSearcher) parseRGLine(alias, repoPath, line string) (*ContentResult, error) {
	// Split by colon, but be careful with file paths that might contain colons
	parts := strings.SplitN(line, ":", 4)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid rg output format: %s", line)
	}

	filePath := parts[0]
	lineNumStr := parts[1]
	colStr := parts[2]
	content := parts[3]

	// Parse line number
	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		return nil, fmt.Errorf("invalid line number: %s", lineNumStr)
	}

	// Parse column number
	column, err := strconv.Atoi(colStr)
	if err != nil {
		return nil, fmt.Errorf("invalid column number: %s", colStr)
	}

	// Calculate relative path from repository root
	relPath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		// If we can't get relative path, use the full path
		relPath = filePath
	}

	// Store the absolute path (filePath is already absolute from rg output)
	fullPath := filePath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(repoPath, filePath)
	}

	// Create display text: "alias:path:line: content"
	displayText := fmt.Sprintf("%s:%s:%d: %s", alias, relPath, lineNum, strings.TrimSpace(content))

	return &ContentResult{
		RepoAlias:   alias,
		FilePath:    relPath,
		FullPath:    fullPath,
		LineNumber:  lineNum,
		LineContent: content,
		MatchColumn: column,
		DisplayText: displayText,
	}, nil
}

// FormatForFZF formats content results for fzf input
// Format: "absolute_path:line_number:display_text"
func (rs *RGSearcher) FormatForFZF(results []ContentResult) string {
	var lines []string
	for _, result := range results {
		// Format: absolute_path:line_number:display_text
		// This allows fzf key bindings to extract the path and line number easily
		line := fmt.Sprintf("%s:%d:%s", result.FullPath, result.LineNumber, result.DisplayText)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// ParseFZFSelection parses fzf selection and returns the corresponding content result
func (rs *RGSearcher) ParseFZFSelection(selection string, results []ContentResult) (*ContentResult, error) {
	// New format is: "absolute_path:line_number:display_text"
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