package index

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gman/internal/config"
)

// Searcher provides high-level search functionality
type Searcher struct {
	indexer *Indexer
	storage *Storage
}

// SearchResult represents a search result
type SearchResult struct {
	Type        string      `json:"type"` // "file" or "commit"
	RepoAlias   string      `json:"repo_alias"`
	DisplayText string      `json:"display_text"` // Formatted text for display
	Path        string      `json:"path"`         // For files
	Hash        string      `json:"hash"`         // For commits
	Data        interface{} `json:"data"`         // Full FileEntry or CommitEntry
}

// NewSearcher creates a new searcher instance
func NewSearcher(configMgr *config.Manager) (*Searcher, error) {
	indexer, err := NewIndexer(configMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &Searcher{
		indexer: indexer,
		storage: indexer.GetStorage(),
	}, nil
}

// Close closes the searcher and releases resources
func (s *Searcher) Close() error {
	return s.indexer.Close()
}

// EnsureIndex ensures that the search index exists and is up to date
func (s *Searcher) EnsureIndex(repos map[string]string, force bool) error {
	needsIndexing, err := s.indexer.NeedsIndexing(repos)
	if err != nil {
		return fmt.Errorf("failed to check index status: %w", err)
	}

	if needsIndexing || force {
		return s.indexer.BuildIndex(repos, nil)
	}

	return nil
}

// SearchFiles searches for files across all repositories
func (s *Searcher) SearchFiles(query string, groupFilter string, repos map[string]string) ([]SearchResult, error) {
	// Apply group filter if specified
	repoFilter := s.getRepositoriesFromGroup(groupFilter, repos)

	var files []FileEntry
	var err error

	if query == "" {
		// No query, return all files
		files, err = s.storage.GetAllFiles(repoFilter)
	} else {
		// Search with query
		files, err = s.storage.SearchFiles(query, repoFilter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	// Convert to SearchResults
	results := make([]SearchResult, len(files))
	for i, file := range files {
		results[i] = SearchResult{
			Type:        "file",
			RepoAlias:   file.RepoAlias,
			DisplayText: fmt.Sprintf("%s: %s", file.RepoAlias, file.RelativePath),
			Path:        file.AbsolutePath,
			Data:        file,
		}
	}

	return results, nil
}

// SearchCommits searches for commits across all repositories
func (s *Searcher) SearchCommits(query string, groupFilter string, repos map[string]string) ([]SearchResult, error) {
	// Apply group filter if specified
	repoFilter := s.getRepositoriesFromGroup(groupFilter, repos)

	var commits []CommitEntry
	var err error

	if query == "" {
		// No query, return all commits
		commits, err = s.storage.GetAllCommits(repoFilter)
	} else {
		// Search with query
		commits, err = s.storage.SearchCommits(query, repoFilter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search commits: %w", err)
	}

	// Convert to SearchResults
	results := make([]SearchResult, len(commits))
	for i, commit := range commits {
		// Format commit display text
		shortHash := commit.Hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}

		// Truncate subject if too long
		subject := commit.Subject
		if len(subject) > 60 {
			subject = subject[:57] + "..."
		}

		results[i] = SearchResult{
			Type:        "commit",
			RepoAlias:   commit.RepoAlias,
			DisplayText: fmt.Sprintf("%s: %s %s %s", commit.RepoAlias, shortHash, commit.Author, subject),
			Hash:        commit.Hash,
			Data:        commit,
		}
	}

	return results, nil
}

// getRepositoriesFromGroup converts a group name to a list of repository aliases
func (s *Searcher) getRepositoriesFromGroup(groupFilter string, repos map[string]string) []string {
	if groupFilter == "" {
		return nil // No filter, search all repos
	}

	groupRepos, err := s.indexer.configMgr.GetGroupRepositories(groupFilter)
	if err != nil {
		return nil // Group not found, search all repos
	}

	var repoFilter []string
	for alias := range groupRepos {
		repoFilter = append(repoFilter, alias)
	}

	return repoFilter
}

// UpdateIndex updates the search index
func (s *Searcher) UpdateIndex(repos map[string]string) error {
	return s.indexer.UpdateIndex(repos)
}

// RebuildIndex completely rebuilds the search index
func (s *Searcher) RebuildIndex(repos map[string]string, progress func(string, int, int)) error {
	return s.indexer.RebuildIndex(repos, progress)
}

// GetIndexStats returns statistics about the search index
func (s *Searcher) GetIndexStats() (map[string]interface{}, error) {
	return s.storage.GetStats()
}

// FormatFileSearchResults formats file search results for fzf consumption
func (s *Searcher) FormatFileSearchResults(results []SearchResult) []string {
	formatted := make([]string, len(results))
	for i, result := range results {
		formatted[i] = result.DisplayText
	}
	return formatted
}

// FormatCommitSearchResults formats commit search results for fzf consumption
func (s *Searcher) FormatCommitSearchResults(results []SearchResult) []string {
	formatted := make([]string, len(results))
	for i, result := range results {
		formatted[i] = result.DisplayText
	}
	return formatted
}

// ParseFileSelection parses a selected line from fzf back to file information
func (s *Searcher) ParseFileSelection(selection string, results []SearchResult) (*SearchResult, error) {
	for _, result := range results {
		if result.DisplayText == selection {
			return &result, nil
		}
	}
	return nil, fmt.Errorf("selection not found: %s", selection)
}

// ParseCommitSelection parses a selected line from fzf back to commit information
func (s *Searcher) ParseCommitSelection(selection string, results []SearchResult) (*SearchResult, error) {
	for _, result := range results {
		if result.DisplayText == selection {
			return &result, nil
		}
	}
	return nil, fmt.Errorf("selection not found: %s", selection)
}

// GetFilePreview generates preview content for a file
func (s *Searcher) GetFilePreview(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "File not found", nil
	}

	// Try to use bat for syntax highlighting, fallback to cat
	batCmd := []string{"bat", "--style=numbers", "--color=always", filePath}
	catCmd := []string{"cat", filePath}

	// Check if bat is available
	if _, err := exec.LookPath("bat"); err == nil {
		// Use bat
		return s.runCommandForPreview(batCmd)
	}

	// Fallback to cat
	return s.runCommandForPreview(catCmd)
}

// GetCommitPreview generates preview content for a commit
func (s *Searcher) GetCommitPreview(repoPath, hash string) (string, error) {
	cmd := []string{"git", "-C", repoPath, "show", "--color=always", hash}
	return s.runCommandForPreview(cmd)
}

// runCommandForPreview runs a command and returns its output for preview
func (s *Searcher) runCommandForPreview(cmd []string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Execute the command
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error executing preview command: %v", err), nil
	}

	return string(output), nil
}

// ValidateSearchTerm validates a search term for safety and effectiveness
func (s *Searcher) ValidateSearchTerm(term string) error {
	if term == "" {
		return nil // Empty terms are allowed (show all)
	}

	if len(term) < 1 {
		return fmt.Errorf("search term too short")
	}

	if len(term) > 200 {
		return fmt.Errorf("search term too long")
	}

	// Basic validation for SQL injection prevention (FTS5 is generally safe, but be cautious)
	dangerous := []string{"'", "\"", ";", "--", "/*", "*/"}
	for _, d := range dangerous {
		if strings.Contains(term, d) {
			return fmt.Errorf("search term contains invalid characters")
		}
	}

	return nil
}

// GetRecentSearches returns recently used search terms (if implemented)
func (s *Searcher) GetRecentSearches() []string {
	// This could be implemented to track and return recent search terms
	// For now, return empty slice
	return []string{}
}

// SaveRecentSearch saves a search term to recent searches (if implemented)
func (s *Searcher) SaveRecentSearch(term string) error {
	// This could be implemented to save recent search terms
	// For now, do nothing
	return nil
}
