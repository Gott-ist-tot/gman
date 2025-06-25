package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gman/internal/config"
	"gman/internal/di"
	"gman/internal/git"
)

// Indexer manages the creation and maintenance of search indexes
type Indexer struct {
	storage   *Storage
	gitMgr    *git.Manager
	configMgr *config.Manager
}

// NewIndexer creates a new indexer instance
func NewIndexer(configMgr *config.Manager) (*Indexer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "gman")
	storage, err := NewStorage(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return &Indexer{
		storage:   storage,
		gitMgr:    di.GitManager(),
		configMgr: configMgr,
	}, nil
}

// Close closes the indexer and releases resources
func (idx *Indexer) Close() error {
	return idx.storage.Close()
}

// BuildIndex builds the complete index for all repositories
func (idx *Indexer) BuildIndex(repos map[string]string, progress func(string, int, int)) error {
	total := len(repos) * 2 // files + commits for each repo
	current := 0

	for alias, path := range repos {
		if progress != nil {
			progress(fmt.Sprintf("Indexing files in %s", alias), current, total)
		}

		// Index files
		if err := idx.indexRepoFiles(alias, path); err != nil {
			return fmt.Errorf("failed to index files for %s: %w", alias, err)
		}
		current++

		if progress != nil {
			progress(fmt.Sprintf("Indexing commits in %s", alias), current, total)
		}

		// Index commits
		if err := idx.indexRepoCommits(alias, path); err != nil {
			return fmt.Errorf("failed to index commits for %s: %w", alias, err)
		}
		current++
	}

	if progress != nil {
		progress("Index build complete", total, total)
	}

	return nil
}

// UpdateIndex performs incremental updates to the index
func (idx *Indexer) UpdateIndex(repos map[string]string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(repos)*2)

	for alias, path := range repos {
		wg.Add(2)

		// Update files in parallel
		go func(alias, path string) {
			defer wg.Done()
			if err := idx.indexRepoFiles(alias, path); err != nil {
				errChan <- fmt.Errorf("failed to update files for %s: %w", alias, err)
			}
		}(alias, path)

		// Update commits in parallel
		go func(alias, path string) {
			defer wg.Done()
			if err := idx.indexRepoCommits(alias, path); err != nil {
				errChan <- fmt.Errorf("failed to update commits for %s: %w", alias, err)
			}
		}(alias, path)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("index update failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// indexRepoFiles scans and indexes all files in a repository
func (idx *Indexer) indexRepoFiles(repoAlias, repoPath string) error {
	var files []FileEntry

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip .git directory and its contents
		if strings.Contains(path, ".git/") || strings.Contains(path, ".git\\") {
			return nil
		}

		// Skip hidden files and common ignore patterns
		if shouldIgnoreFile(path, info) {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil // Skip if we can't get relative path
		}

		files = append(files, FileEntry{
			RepoAlias:    repoAlias,
			RelativePath: relPath,
			AbsolutePath: path,
			ModTime:      info.ModTime(),
			FileSize:     info.Size(),
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk repository %s: %w", repoPath, err)
	}

	// Insert files in batches to improve performance
	const batchSize = 1000
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		if err := idx.storage.InsertFiles(files[i:end]); err != nil {
			return fmt.Errorf("failed to insert file batch: %w", err)
		}
	}

	return nil
}

// indexRepoCommits scans and indexes commit history from a repository
func (idx *Indexer) indexRepoCommits(repoAlias, repoPath string) error {
	// Get recent commits (last 1000 commits to balance completeness and performance)
	output, err := idx.gitMgr.RunCommand(repoPath, "log", "--oneline", "--format=%H|%an|%s|%ct|", "-1000")
	if err != nil {
		// Repository might not have any commits, which is okay
		return nil
	}

	if strings.TrimSpace(output) == "" {
		return nil // No commits
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var commits []CommitEntry

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue // Skip malformed lines
		}

		hash := parts[0]
		author := parts[1]
		subject := parts[2]
		timestampStr := parts[3]

		// Parse timestamp
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue // Skip commits with invalid timestamps
		}

		// Get files changed for this commit
		filesChanged := idx.getCommitFilesChanged(repoPath, hash)

		commits = append(commits, CommitEntry{
			RepoAlias:    repoAlias,
			Hash:         hash,
			Author:       author,
			Subject:      subject,
			Date:         time.Unix(timestamp, 0),
			FilesChanged: filesChanged,
		})
	}

	// Insert commits in batches
	const batchSize = 100
	for i := 0; i < len(commits); i += batchSize {
		end := i + batchSize
		if end > len(commits) {
			end = len(commits)
		}

		if err := idx.storage.InsertCommits(commits[i:end]); err != nil {
			return fmt.Errorf("failed to insert commit batch: %w", err)
		}
	}

	return nil
}

// getCommitFilesChanged returns the number of files changed in a commit
func (idx *Indexer) getCommitFilesChanged(repoPath, hash string) int {
	output, err := idx.gitMgr.RunCommand(repoPath, "show", "--name-only", "--format=", hash)
	if err != nil {
		return 0
	}

	if strings.TrimSpace(output) == "" {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

// shouldIgnoreFile determines if a file should be ignored during indexing
func shouldIgnoreFile(path string, info os.FileInfo) bool {
	name := info.Name()

	// Skip hidden files
	if strings.HasPrefix(name, ".") && name != "." && name != ".." {
		return true
	}

	// Skip common binary extensions
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".a", ".o", ".obj",
		".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg",
		".mp3", ".mp4", ".avi", ".mov", ".wav", ".flac",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	}

	ext := strings.ToLower(filepath.Ext(name))
	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}

	// Skip common build/cache directories
	skipDirs := []string{
		"node_modules", "vendor", "target", "build", "dist", ".cache",
		"__pycache__", ".pytest_cache", ".mypy_cache", ".tox",
		".gradle", ".idea", ".vscode", ".vs",
	}

	for _, skipDir := range skipDirs {
		if strings.Contains(path, string(filepath.Separator)+skipDir+string(filepath.Separator)) ||
			strings.HasSuffix(filepath.Dir(path), string(filepath.Separator)+skipDir) {
			return true
		}
	}

	// Skip very large files (> 10MB)
	if info.Size() > 10*1024*1024 {
		return true
	}

	return false
}

// RebuildIndex completely rebuilds the index, clearing existing data
func (idx *Indexer) RebuildIndex(repos map[string]string, progress func(string, int, int)) error {
	// Clear existing data
	for alias := range repos {
		if err := idx.storage.ClearRepository(alias); err != nil {
			return fmt.Errorf("failed to clear existing data for %s: %w", alias, err)
		}
	}

	// Rebuild from scratch
	return idx.BuildIndex(repos, progress)
}

// UpdateSingleRepository updates the index for a single repository
func (idx *Indexer) UpdateSingleRepository(alias, path string) error {
	// Clear existing data for this repository
	if err := idx.storage.ClearRepository(alias); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Reindex files and commits
	if err := idx.indexRepoFiles(alias, path); err != nil {
		return fmt.Errorf("failed to index files: %w", err)
	}

	if err := idx.indexRepoCommits(alias, path); err != nil {
		return fmt.Errorf("failed to index commits: %w", err)
	}

	return nil
}

// GetStorage returns the underlying storage instance for direct access
func (idx *Indexer) GetStorage() *Storage {
	return idx.storage
}

// NeedsIndexing checks if a repository needs to be indexed or updated
func (idx *Indexer) NeedsIndexing(repos map[string]string) (bool, error) {
	stats, err := idx.storage.GetStats()
	if err != nil {
		return true, err // If we can't get stats, assume we need indexing
	}

	// If we have no files indexed, we need to build the index
	fileCount, ok := stats["file_count"].(int)
	if !ok || fileCount == 0 {
		return true, nil
	}

	// Check if all repositories are represented in the index
	repoCount, ok := stats["repository_count"].(int)
	if !ok || repoCount != len(repos) {
		return true, nil
	}

	return false, nil
}
