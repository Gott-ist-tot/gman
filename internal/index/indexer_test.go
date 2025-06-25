package index

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gman/internal/config"
	"gman/internal/di"
	"gman/pkg/types"
)

func TestIndexer_shouldIgnoreFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		fileInfo os.FileInfo
		want     bool
	}{
		{
			name: "should ignore hidden files",
			path: "/repo/.hidden",
			fileInfo: &mockFileInfo{
				name:  ".hidden",
				isDir: false,
				size:  100,
			},
			want: true,
		},
		{
			name: "should ignore node_modules",
			path: "/repo/node_modules/package/file.js",
			fileInfo: &mockFileInfo{
				name:  "file.js",
				isDir: false,
				size:  100,
			},
			want: true,
		},
		{
			name: "should ignore binary files",
			path: "/repo/binary.exe",
			fileInfo: &mockFileInfo{
				name:  "binary.exe",
				isDir: false,
				size:  100,
			},
			want: true,
		},
		{
			name: "should ignore large files",
			path: "/repo/large.txt",
			fileInfo: &mockFileInfo{
				name:  "large.txt",
				isDir: false,
				size:  20 * 1024 * 1024, // 20MB
			},
			want: true,
		},
		{
			name: "should not ignore normal files",
			path: "/repo/main.go",
			fileInfo: &mockFileInfo{
				name:  "main.go",
				isDir: false,
				size:  1024,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnoreFile(tt.path, tt.fileInfo); got != tt.want {
				t.Errorf("shouldIgnoreFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_FileOperations(t *testing.T) {
	// Create temporary storage
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test file insertion
	files := []FileEntry{
		{
			RepoAlias:    "test-repo",
			RelativePath: "main.go",
			AbsolutePath: "/path/to/test-repo/main.go",
			ModTime:      time.Now(),
			FileSize:     1024,
		},
		{
			RepoAlias:    "test-repo",
			RelativePath: "config.yml",
			AbsolutePath: "/path/to/test-repo/config.yml",
			ModTime:      time.Now(),
			FileSize:     512,
		},
	}

	err = storage.InsertFiles(files)
	if err != nil {
		t.Fatalf("Failed to insert files: %v", err)
	}

	// Test file retrieval
	allFiles, err := storage.GetAllFiles([]string{"test-repo"})
	if err != nil {
		t.Fatalf("Failed to get files: %v", err)
	}

	if len(allFiles) != 2 {
		t.Errorf("Expected 2 files, got %d", len(allFiles))
	}

	// Test file search
	searchResults, err := storage.SearchFiles("main", []string{"test-repo"})
	if err != nil {
		t.Fatalf("Failed to search files: %v", err)
	}

	if len(searchResults) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searchResults))
	}

	if searchResults[0].RelativePath != "main.go" {
		t.Errorf("Expected main.go, got %s", searchResults[0].RelativePath)
	}
}

func TestStorage_CommitOperations(t *testing.T) {
	// Create temporary storage
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test commit insertion
	commits := []CommitEntry{
		{
			RepoAlias:    "test-repo",
			Hash:         "abc123",
			Author:       "John Doe",
			Subject:      "Fix critical bug",
			Date:         time.Now(),
			FilesChanged: 3,
		},
		{
			RepoAlias:    "test-repo",
			Hash:         "def456",
			Author:       "Jane Smith",
			Subject:      "Add new feature",
			Date:         time.Now().Add(-time.Hour),
			FilesChanged: 5,
		},
	}

	err = storage.InsertCommits(commits)
	if err != nil {
		t.Fatalf("Failed to insert commits: %v", err)
	}

	// Test commit retrieval
	allCommits, err := storage.GetAllCommits([]string{"test-repo"})
	if err != nil {
		t.Fatalf("Failed to get commits: %v", err)
	}

	if len(allCommits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(allCommits))
	}

	// Test commit search
	searchResults, err := storage.SearchCommits("bug", []string{"test-repo"})
	if err != nil {
		t.Fatalf("Failed to search commits: %v", err)
	}

	if len(searchResults) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searchResults))
	}

	if searchResults[0].Subject != "Fix critical bug" {
		t.Errorf("Expected 'Fix critical bug', got %s", searchResults[0].Subject)
	}
}

func TestStorage_Stats(t *testing.T) {
	// Create temporary storage
	tempDir := t.TempDir()
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Insert test data
	files := []FileEntry{
		{
			RepoAlias:    "repo1",
			RelativePath: "file1.go",
			AbsolutePath: "/path/to/repo1/file1.go",
			ModTime:      time.Now(),
			FileSize:     1024,
		},
		{
			RepoAlias:    "repo2",
			RelativePath: "file2.go",
			AbsolutePath: "/path/to/repo2/file2.go",
			ModTime:      time.Now(),
			FileSize:     2048,
		},
	}

	commits := []CommitEntry{
		{
			RepoAlias:    "repo1",
			Hash:         "abc123",
			Author:       "John Doe",
			Subject:      "Initial commit",
			Date:         time.Now(),
			FilesChanged: 1,
		},
	}

	err = storage.InsertFiles(files)
	if err != nil {
		t.Fatalf("Failed to insert files: %v", err)
	}

	err = storage.InsertCommits(commits)
	if err != nil {
		t.Fatalf("Failed to insert commits: %v", err)
	}

	// Test stats
	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if fileCount, ok := stats["file_count"].(int); !ok || fileCount != 2 {
		t.Errorf("Expected file_count 2, got %v", stats["file_count"])
	}

	if commitCount, ok := stats["commit_count"].(int); !ok || commitCount != 1 {
		t.Errorf("Expected commit_count 1, got %v", stats["commit_count"])
	}

	if repoCount, ok := stats["repository_count"].(int); !ok || repoCount != 2 {
		t.Errorf("Expected repository_count 2, got %v", stats["repository_count"])
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// Helper function to create a test config manager
func createTestConfigManager(tempDir string) *config.Manager {
	configMgr := di.ConfigManager()

	// Create a mock config
	cfg := &types.Config{
		Repositories: map[string]string{
			"test-repo": filepath.Join(tempDir, "test-repo"),
		},
		Settings: types.Settings{
			ParallelJobs:    5,
			ShowLastCommit:  true,
			DefaultSyncMode: "ff-only",
		},
	}

	// Note: In a real test, you'd want to set this config properly
	// This is a simplified version for demonstration
	_ = cfg

	return configMgr
}

func BenchmarkFileIndexing(b *testing.B) {
	tempDir := b.TempDir()
	storage, err := NewStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create test files
	files := make([]FileEntry, 1000)
	for i := 0; i < 1000; i++ {
		files[i] = FileEntry{
			RepoAlias:    "test-repo",
			RelativePath: fmt.Sprintf("file%d.go", i),
			AbsolutePath: fmt.Sprintf("/path/to/test-repo/file%d.go", i),
			ModTime:      time.Now(),
			FileSize:     1024,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := storage.InsertFiles(files)
		if err != nil {
			b.Fatalf("Failed to insert files: %v", err)
		}
	}
}

func BenchmarkFileSearch(b *testing.B) {
	tempDir := b.TempDir()
	storage, err := NewStorage(tempDir)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Insert test data
	files := make([]FileEntry, 10000)
	for i := 0; i < 10000; i++ {
		files[i] = FileEntry{
			RepoAlias:    "test-repo",
			RelativePath: fmt.Sprintf("file%d.go", i),
			AbsolutePath: fmt.Sprintf("/path/to/test-repo/file%d.go", i),
			ModTime:      time.Now(),
			FileSize:     1024,
		}
	}

	err = storage.InsertFiles(files)
	if err != nil {
		b.Fatalf("Failed to insert files: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := storage.SearchFiles("file", []string{"test-repo"})
		if err != nil {
			b.Fatalf("Failed to search files: %v", err)
		}
	}
}
