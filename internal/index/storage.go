package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Storage handles SQLite operations for the index
type Storage struct {
	db   *sql.DB
	path string
}

// FileEntry represents a file in the index
type FileEntry struct {
	RepoAlias    string    `json:"repo_alias"`
	RelativePath string    `json:"relative_path"`
	AbsolutePath string    `json:"absolute_path"`
	ModTime      time.Time `json:"mod_time"`
	FileSize     int64     `json:"file_size"`
}

// CommitEntry represents a commit in the index
type CommitEntry struct {
	RepoAlias    string    `json:"repo_alias"`
	Hash         string    `json:"hash"`
	Author       string    `json:"author"`
	Subject      string    `json:"subject"`
	Date         time.Time `json:"date"`
	FilesChanged int       `json:"files_changed"`
}

// NewStorage creates a new storage instance
func NewStorage(configDir string) (*Storage, error) {
	dbPath := filepath.Join(configDir, "gman.index.db")
	
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{
		db:   db,
		path: dbPath,
	}

	if err := storage.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return storage, nil
}

// initialize creates the necessary tables and indexes
func (s *Storage) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS file_index (
		repo_alias TEXT NOT NULL,
		relative_path TEXT NOT NULL,
		absolute_path TEXT NOT NULL,
		mod_time INTEGER NOT NULL,
		file_size INTEGER NOT NULL,
		PRIMARY KEY (repo_alias, relative_path)
	);

	CREATE TABLE IF NOT EXISTS commit_index (
		repo_alias TEXT NOT NULL,
		hash TEXT NOT NULL,
		author TEXT NOT NULL,
		subject TEXT NOT NULL,
		date INTEGER NOT NULL,
		files_changed INTEGER NOT NULL,
		PRIMARY KEY (repo_alias, hash)
	);

	-- Create indexes for faster searching
	CREATE INDEX IF NOT EXISTS idx_file_path ON file_index(relative_path);
	CREATE INDEX IF NOT EXISTS idx_commit_subject ON commit_index(subject);
	CREATE INDEX IF NOT EXISTS idx_commit_author ON commit_index(author);
	CREATE INDEX IF NOT EXISTS idx_commit_date ON commit_index(date);

	-- Create virtual tables for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS file_search USING fts5(
		repo_alias,
		relative_path,
		content='file_index',
		content_rowid='rowid'
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS commit_search USING fts5(
		repo_alias,
		hash,
		author,
		subject,
		content='commit_index',
		content_rowid='rowid'
	);

	-- Triggers to keep FTS tables in sync
	CREATE TRIGGER IF NOT EXISTS file_ai AFTER INSERT ON file_index BEGIN
		INSERT INTO file_search(rowid, repo_alias, relative_path) 
		VALUES (new.rowid, new.repo_alias, new.relative_path);
	END;

	CREATE TRIGGER IF NOT EXISTS file_ad AFTER DELETE ON file_index BEGIN
		INSERT INTO file_search(file_search, rowid, repo_alias, relative_path) 
		VALUES('delete', old.rowid, old.repo_alias, old.relative_path);
	END;

	CREATE TRIGGER IF NOT EXISTS commit_ai AFTER INSERT ON commit_index BEGIN
		INSERT INTO commit_search(rowid, repo_alias, hash, author, subject) 
		VALUES (new.rowid, new.repo_alias, new.hash, new.author, new.subject);
	END;

	CREATE TRIGGER IF NOT EXISTS commit_ad AFTER DELETE ON commit_index BEGIN
		INSERT INTO commit_search(commit_search, rowid, repo_alias, hash, author, subject) 
		VALUES('delete', old.rowid, old.repo_alias, old.hash, old.author, old.subject);
	END;
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// InsertFiles inserts or updates file entries
func (s *Storage) InsertFiles(files []FileEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO file_index 
		(repo_alias, relative_path, absolute_path, mod_time, file_size) 
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, file := range files {
		_, err = stmt.Exec(
			file.RepoAlias,
			file.RelativePath,
			file.AbsolutePath,
			file.ModTime.Unix(),
			file.FileSize,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// InsertCommits inserts or updates commit entries
func (s *Storage) InsertCommits(commits []CommitEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO commit_index 
		(repo_alias, hash, author, subject, date, files_changed) 
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, commit := range commits {
		_, err = stmt.Exec(
			commit.RepoAlias,
			commit.Hash,
			commit.Author,
			commit.Subject,
			commit.Date.Unix(),
			commit.FilesChanged,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SearchFiles performs full-text search on files
func (s *Storage) SearchFiles(query string, repoFilter []string) ([]FileEntry, error) {
	var rows *sql.Rows
	var err error

	if len(repoFilter) > 0 {
		// Build placeholders for IN clause
		placeholders := ""
		args := []interface{}{query}
		for i, repo := range repoFilter {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, repo)
		}

		querySQL := fmt.Sprintf(`
			SELECT f.repo_alias, f.relative_path, f.absolute_path, f.mod_time, f.file_size
			FROM file_search fs
			JOIN file_index f ON f.rowid = fs.rowid
			WHERE file_search MATCH ? AND f.repo_alias IN (%s)
			ORDER BY rank
		`, placeholders)

		rows, err = s.db.Query(querySQL, args...)
	} else {
		rows, err = s.db.Query(`
			SELECT f.repo_alias, f.relative_path, f.absolute_path, f.mod_time, f.file_size
			FROM file_search fs
			JOIN file_index f ON f.rowid = fs.rowid
			WHERE file_search MATCH ?
			ORDER BY rank
		`, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileEntry
	for rows.Next() {
		var file FileEntry
		var modTime int64
		err = rows.Scan(
			&file.RepoAlias,
			&file.RelativePath,
			&file.AbsolutePath,
			&modTime,
			&file.FileSize,
		)
		if err != nil {
			return nil, err
		}
		file.ModTime = time.Unix(modTime, 0)
		files = append(files, file)
	}

	return files, rows.Err()
}

// SearchCommits performs full-text search on commits
func (s *Storage) SearchCommits(query string, repoFilter []string) ([]CommitEntry, error) {
	var rows *sql.Rows
	var err error

	if len(repoFilter) > 0 {
		// Build placeholders for IN clause
		placeholders := ""
		args := []interface{}{query}
		for i, repo := range repoFilter {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, repo)
		}

		querySQL := fmt.Sprintf(`
			SELECT c.repo_alias, c.hash, c.author, c.subject, c.date, c.files_changed
			FROM commit_search cs
			JOIN commit_index c ON c.rowid = cs.rowid
			WHERE commit_search MATCH ? AND c.repo_alias IN (%s)
			ORDER BY c.date DESC
		`, placeholders)

		rows, err = s.db.Query(querySQL, args...)
	} else {
		rows, err = s.db.Query(`
			SELECT c.repo_alias, c.hash, c.author, c.subject, c.date, c.files_changed
			FROM commit_search cs
			JOIN commit_index c ON c.rowid = cs.rowid
			WHERE commit_search MATCH ?
			ORDER BY c.date DESC
		`, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []CommitEntry
	for rows.Next() {
		var commit CommitEntry
		var date int64
		err = rows.Scan(
			&commit.RepoAlias,
			&commit.Hash,
			&commit.Author,
			&commit.Subject,
			&date,
			&commit.FilesChanged,
		)
		if err != nil {
			return nil, err
		}
		commit.Date = time.Unix(date, 0)
		commits = append(commits, commit)
	}

	return commits, rows.Err()
}

// GetAllFiles returns all files, optionally filtered by repository
func (s *Storage) GetAllFiles(repoFilter []string) ([]FileEntry, error) {
	var rows *sql.Rows
	var err error

	if len(repoFilter) > 0 {
		// Build placeholders for IN clause
		placeholders := ""
		args := []interface{}{}
		for i, repo := range repoFilter {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, repo)
		}

		querySQL := fmt.Sprintf(`
			SELECT repo_alias, relative_path, absolute_path, mod_time, file_size
			FROM file_index
			WHERE repo_alias IN (%s)
			ORDER BY repo_alias, relative_path
		`, placeholders)

		rows, err = s.db.Query(querySQL, args...)
	} else {
		rows, err = s.db.Query(`
			SELECT repo_alias, relative_path, absolute_path, mod_time, file_size
			FROM file_index
			ORDER BY repo_alias, relative_path
		`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileEntry
	for rows.Next() {
		var file FileEntry
		var modTime int64
		err = rows.Scan(
			&file.RepoAlias,
			&file.RelativePath,
			&file.AbsolutePath,
			&modTime,
			&file.FileSize,
		)
		if err != nil {
			return nil, err
		}
		file.ModTime = time.Unix(modTime, 0)
		files = append(files, file)
	}

	return files, rows.Err()
}

// GetAllCommits returns all commits, optionally filtered by repository
func (s *Storage) GetAllCommits(repoFilter []string) ([]CommitEntry, error) {
	var rows *sql.Rows
	var err error

	if len(repoFilter) > 0 {
		// Build placeholders for IN clause
		placeholders := ""
		args := []interface{}{}
		for i, repo := range repoFilter {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, repo)
		}

		querySQL := fmt.Sprintf(`
			SELECT repo_alias, hash, author, subject, date, files_changed
			FROM commit_index
			WHERE repo_alias IN (%s)
			ORDER BY date DESC
		`, placeholders)

		rows, err = s.db.Query(querySQL, args...)
	} else {
		rows, err = s.db.Query(`
			SELECT repo_alias, hash, author, subject, date, files_changed
			FROM commit_index
			ORDER BY date DESC
		`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []CommitEntry
	for rows.Next() {
		var commit CommitEntry
		var date int64
		err = rows.Scan(
			&commit.RepoAlias,
			&commit.Hash,
			&commit.Author,
			&commit.Subject,
			&date,
			&commit.FilesChanged,
		)
		if err != nil {
			return nil, err
		}
		commit.Date = time.Unix(date, 0)
		commits = append(commits, commit)
	}

	return commits, rows.Err()
}

// ClearRepository removes all entries for a specific repository
func (s *Storage) ClearRepository(repoAlias string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM file_index WHERE repo_alias = ?", repoAlias)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM commit_index WHERE repo_alias = ?", repoAlias)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetStats returns statistics about the index
func (s *Storage) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get file count
	var fileCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM file_index").Scan(&fileCount)
	if err != nil {
		return nil, err
	}
	stats["file_count"] = fileCount

	// Get commit count
	var commitCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM commit_index").Scan(&commitCount)
	if err != nil {
		return nil, err
	}
	stats["commit_count"] = commitCount

	// Get repository count
	var repoCount int
	err = s.db.QueryRow("SELECT COUNT(DISTINCT repo_alias) FROM file_index").Scan(&repoCount)
	if err != nil {
		return nil, err
	}
	stats["repository_count"] = repoCount

	// Get database file size
	if info, err := os.Stat(s.path); err == nil {
		stats["db_size"] = info.Size()
	}

	return stats, nil
}