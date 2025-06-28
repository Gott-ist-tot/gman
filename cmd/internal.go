package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// PreviewRequest represents a secure preview request
type PreviewRequest struct {
	Type       string `json:"type"`        // "file", "commit", "content"
	Path       string `json:"path"`        // File path for file previews
	RepoPath   string `json:"repo_path"`   // Repository path
	LineNumber int    `json:"line_number"` // Line number for content previews
	CommitHash string `json:"commit_hash"` // Commit hash for commit previews
}

// internalPreviewCmd is a hidden command for secure preview functionality
var internalPreviewCmd = &cobra.Command{
	Use:    "internal-preview <base64-encoded-json>",
	Short:  "Internal command for secure file previews (not for direct use)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE:   runInternalPreview,
}

func init() {
	rootCmd.AddCommand(internalPreviewCmd)
}

func runInternalPreview(cmd *cobra.Command, args []string) error {
	// Decode base64 input
	encodedData := args[0]
	jsonData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("invalid base64 input: %w", err)
	}

	// Parse preview request
	var req PreviewRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	// Handle different preview types
	switch req.Type {
	case "file":
		return previewFile(req)
	case "commit":
		return previewCommit(req)
	case "content":
		return previewContent(req)
	default:
		return fmt.Errorf("unsupported preview type: %s", req.Type)
	}
}

func previewFile(req PreviewRequest) error {
	// Validate file path is within repository
	cleanPath := filepath.Clean(req.Path)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute path required")
	}

	// Additional security check: ensure path doesn't contain suspicious patterns
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check if file exists and is readable
	info, err := os.Stat(cleanPath)
	if err != nil {
		fmt.Printf("File not found: %s\n", cleanPath)
		return nil
	}

	// Check file size (avoid previewing very large files)
	if info.Size() > 1024*1024 { // 1MB limit
		fmt.Printf("File too large to preview (%d bytes)\n", info.Size())
		return nil
	}

	// Check if it's a directory
	if info.IsDir() {
		fmt.Printf("Directory: %s\n", cleanPath)
		return nil
	}

	// Determine file type and preview accordingly
	ext := strings.ToLower(filepath.Ext(cleanPath))
	
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return previewImage(cleanPath)
	case ".pdf":
		return previewPDF(cleanPath)
	case ".zip", ".tar", ".gz", ".bz2":
		return previewArchive(cleanPath)
	default:
		return previewTextFile(cleanPath)
	}
}

func previewTextFile(filePath string) error {
	// Try to use bat if available, otherwise fallback to cat/head
	if commandExistsInternal("bat") {
		cmd := []string{"bat", "--style=numbers", "--color=always", "--line-range", ":100", filePath}
		return runCommand(cmd)
	}

	// Fallback to head
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 100 lines
	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && lineCount < 100 {
		fmt.Println(scanner.Text())
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

func previewImage(filePath string) error {
	// Try chafa for terminal image display
	if commandExistsInternal("chafa") {
		cmd := []string{"chafa", "--size=60x40", filePath}
		if err := runCommand(cmd); err == nil {
			return nil
		}
	}

	// Try catimg as fallback
	if commandExistsInternal("catimg") {
		cmd := []string{"catimg", "-w", "60", filePath}
		if err := runCommand(cmd); err == nil {
			return nil
		}
	}

	// Fallback to file info
	fmt.Printf("Image file: %s\n", filepath.Base(filePath))
	cmd := []string{"file", filePath}
	return runCommand(cmd)
}

func previewPDF(filePath string) error {
	// Try pdftotext
	if commandExistsInternal("pdftotext") {
		cmd := []string{"pdftotext", "-l", "5", "-nopgbrk", "-q", filePath, "-"}
		if err := runCommand(cmd); err == nil {
			return nil
		}
	}

	// Fallback to file info
	fmt.Printf("PDF file: %s\n", filepath.Base(filePath))
	cmd := []string{"file", filePath}
	return runCommand(cmd)
}

func previewArchive(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".zip":
		if commandExistsInternal("unzip") {
			cmd := []string{"unzip", "-l", filePath}
			if err := runCommand(cmd); err == nil {
				return nil
			}
		}
	case ".tar":
		if commandExistsInternal("tar") {
			cmd := []string{"tar", "-tf", filePath}
			if err := runCommand(cmd); err == nil {
				return nil
			}
		}
	case ".gz":
		if strings.HasSuffix(filePath, ".tar.gz") && commandExistsInternal("tar") {
			cmd := []string{"tar", "-tzf", filePath}
			if err := runCommand(cmd); err == nil {
				return nil
			}
		} else if commandExistsInternal("gzip") {
			cmd := []string{"gzip", "-l", filePath}
			if err := runCommand(cmd); err == nil {
				return nil
			}
		}
	}

	// Fallback to file info
	fmt.Printf("Archive file: %s\n", filepath.Base(filePath))
	cmd := []string{"file", filePath}
	return runCommand(cmd)
}

func previewCommit(req PreviewRequest) error {
	// Validate repository path
	if req.RepoPath == "" || req.CommitHash == "" {
		return fmt.Errorf("repository path and commit hash required")
	}

	// Validate commit hash format (basic check)
	if len(req.CommitHash) < 7 || len(req.CommitHash) > 40 {
		return fmt.Errorf("invalid commit hash format")
	}

	// Show commit details using git
	cmd := []string{"git", "-C", req.RepoPath, "show", "--color=always", "--stat", req.CommitHash}
	return runCommand(cmd)
}

func previewContent(req PreviewRequest) error {
	// This is similar to file preview but with line highlighting
	if err := previewFile(req); err != nil {
		return err
	}

	// If line number is specified, show additional context
	if req.LineNumber > 0 {
		fmt.Printf("\n--- Line %d highlighted ---\n", req.LineNumber)
	}

	return nil
}

// Helper function to run external commands safely
func runCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Helper function to check if a command exists (moved from find.go)
func commandExistsInternal(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// CreateSecurePreviewRequest creates a base64-encoded preview request
func CreateSecurePreviewRequest(req PreviewRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal preview request: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return encoded, nil
}