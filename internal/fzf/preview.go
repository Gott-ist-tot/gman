package fzf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PreviewGenerator generates preview commands for fzf
type PreviewGenerator struct {
	batAvailable bool
}

// NewPreviewGenerator creates a new preview generator
func NewPreviewGenerator() *PreviewGenerator {
	_, err := exec.LookPath("bat")
	return &PreviewGenerator{
		batAvailable: err == nil,
	}
}

// FilePreviewCommand generates a preview command for files
func (pg *PreviewGenerator) FilePreviewCommand() string {
	if pg.batAvailable {
		// Use bat with syntax highlighting
		return "bat --style=numbers --color=always --line-range :100 {2}"
	}

	// Fallback to cat with head to limit lines
	return "head -100 {2} 2>/dev/null || echo 'Cannot preview file'"
}

// CommitPreviewCommand generates a preview command for commits
func (pg *PreviewGenerator) CommitPreviewCommand(repoPath string) string {
	// Extract commit hash and show commit details
	return fmt.Sprintf("echo {} | cut -d' ' -f2 | xargs -I{{}} git -C %s show --color=always --stat {{}}", repoPath)
}

// CustomFilePreview generates a custom preview for a specific file
func (pg *PreviewGenerator) CustomFilePreview(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "File not found", nil
	}

	// Check file size (avoid previewing very large files)
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Sprintf("Error accessing file: %v", err), nil
	}

	if info.Size() > 1024*1024 { // 1MB limit
		return fmt.Sprintf("File too large to preview (%d bytes)", info.Size()), nil
	}

	// Generate appropriate preview based on file type
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return pg.imagePreview(filePath), nil
	case ".pdf":
		return pg.pdfPreview(filePath), nil
	case ".zip", ".tar", ".gz", ".bz2":
		return pg.archivePreview(filePath), nil
	default:
		return pg.textPreview(filePath), nil
	}
}

// textPreview generates preview for text files
func (pg *PreviewGenerator) textPreview(filePath string) string {
	if pg.batAvailable {
		return fmt.Sprintf("bat --style=numbers --color=always --line-range :100 %s", shellEscape(filePath))
	}

	return fmt.Sprintf("head -100 %s 2>/dev/null || echo 'Cannot preview file'", shellEscape(filePath))
}

// imagePreview generates preview for image files
func (pg *PreviewGenerator) imagePreview(filePath string) string {
	// Check if we have image preview tools available
	if _, err := exec.LookPath("chafa"); err == nil {
		return fmt.Sprintf("chafa --size=60x40 %s 2>/dev/null || echo 'Image: %s'", shellEscape(filePath), filepath.Base(filePath))
	}

	if _, err := exec.LookPath("catimg"); err == nil {
		return fmt.Sprintf("catimg -w 60 %s 2>/dev/null || echo 'Image: %s'", shellEscape(filePath), filepath.Base(filePath))
	}

	// Fallback to file info
	return fmt.Sprintf("file %s 2>/dev/null || echo 'Image: %s'", shellEscape(filePath), filepath.Base(filePath))
}

// pdfPreview generates preview for PDF files
func (pg *PreviewGenerator) pdfPreview(filePath string) string {
	// Check if pdftotext is available
	if _, err := exec.LookPath("pdftotext"); err == nil {
		return fmt.Sprintf("pdftotext -l 5 -nopgbrk -q %s - 2>/dev/null || echo 'PDF: %s'", shellEscape(filePath), filepath.Base(filePath))
	}

	// Fallback to file info
	return fmt.Sprintf("file %s 2>/dev/null || echo 'PDF: %s'", shellEscape(filePath), filepath.Base(filePath))
}

// archivePreview generates preview for archive files
func (pg *PreviewGenerator) archivePreview(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".zip":
		return fmt.Sprintf("unzip -l %s 2>/dev/null | head -20 || echo 'Archive: %s'", shellEscape(filePath), filepath.Base(filePath))
	case ".tar":
		return fmt.Sprintf("tar -tf %s 2>/dev/null | head -20 || echo 'Archive: %s'", shellEscape(filePath), filepath.Base(filePath))
	case ".gz":
		if strings.HasSuffix(filePath, ".tar.gz") {
			return fmt.Sprintf("tar -tzf %s 2>/dev/null | head -20 || echo 'Archive: %s'", shellEscape(filePath), filepath.Base(filePath))
		}
		return fmt.Sprintf("gzip -l %s 2>/dev/null || echo 'Compressed: %s'", shellEscape(filePath), filepath.Base(filePath))
	default:
		return fmt.Sprintf("file %s 2>/dev/null || echo 'Archive: %s'", shellEscape(filePath), filepath.Base(filePath))
	}
}

// GenerateFileSearchPreview generates a preview command for file search results
func (pg *PreviewGenerator) GenerateFileSearchPreview() string {
	// The input format is "repo-alias: relative/path"
	// We need to extract the path part and preview it
	script := `
#!/bin/bash
line="$1"
if [[ "$line" =~ ^([^:]+):[[:space:]]*(.+)$ ]]; then
    repo_alias="${BASH_REMATCH[1]}"
    file_path="${BASH_REMATCH[2]}"
    
    # Get the absolute path by looking up the repo
    # This is a simplified version - in practice, you'd need to resolve the repo path
    echo "Repository: $repo_alias"
    echo "File: $file_path"
    echo "----------------------------------------"
    
    # Try to preview the file content
    if [[ -f "$file_path" ]]; then
        if command -v bat >/dev/null 2>&1; then
            bat --style=numbers --color=always --line-range :50 "$file_path" 2>/dev/null
        else
            head -50 "$file_path" 2>/dev/null
        fi
    else
        echo "File not found or not accessible"
    fi
else
    echo "Invalid format: $line"
fi
`

	// Create a temporary script file
	scriptFile, cleanup, err := CreateTempPreviewScript(script)
	if err != nil {
		// Fallback to simple preview
		return "echo {}"
	}

	// Note: In a real implementation, you'd need to manage the cleanup
	// For now, we'll just return the script path
	_ = cleanup // Avoid unused variable warning

	return scriptFile + " {}"
}

// GenerateCommitSearchPreview generates a preview command for commit search results
func (pg *PreviewGenerator) GenerateCommitSearchPreview() string {
	// The input format is "repo-alias: hash author subject"
	script := `
#!/bin/bash
line="$1"
if [[ "$line" =~ ^([^:]+):[[:space:]]*([a-f0-9]+)[[:space:]]+(.+)$ ]]; then
    repo_alias="${BASH_REMATCH[1]}"
    commit_hash="${BASH_REMATCH[2]}"
    
    echo "Repository: $repo_alias"
    echo "Commit: $commit_hash"
    echo "----------------------------------------"
    
    # Try to show the commit (this would need actual repo path resolution)
    # For now, just show the parsed information
    echo "Hash: $commit_hash"
    echo "Details: ${BASH_REMATCH[3]}"
else
    echo "Invalid format: $line"
fi
`

	scriptFile, cleanup, err := CreateTempPreviewScript(script)
	if err != nil {
		return "echo {}"
	}

	_ = cleanup
	return scriptFile + " {}"
}

// shellEscape escapes a string for safe use in shell commands
func shellEscape(s string) string {
	// Simple shell escaping - wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// BuildFilePreviewCommand builds a complete preview command for file searching using secure internal command
func (pg *PreviewGenerator) BuildFilePreviewCommand(repoMap map[string]string) string {
	// Use secure internal preview command instead of shell scripts
	// The input format should be: "repo-alias:relative/path" or "absolute/path"
	script := fmt.Sprintf(`
#!/bin/bash
line="$1"

# Try to parse as "repo-alias:relative/path" format
if [[ "$line" =~ ^([^:]+):[[:space:]]*(.+)$ ]]; then
    repo_alias="${BASH_REMATCH[1]}"
    rel_path="${BASH_REMATCH[2]}"
    
    # Repository path mapping
    case "$repo_alias" in
%s
        *)
            echo "Unknown repository: $repo_alias"
            exit 1
            ;;
    esac
    
    abs_path="$repo_path/$rel_path"
else
    # Treat as absolute path
    abs_path="$line"
fi

# Create secure preview request
request=$(echo '{"type":"file","path":"'"$abs_path"'","repo_path":"","line_number":0,"commit_hash":""}' | base64 -w 0)

# Call secure internal preview command
gman internal-preview "$request"
`, pg.generateRepoMapping(repoMap))

	scriptFile, _, err := CreateTempPreviewScript(script)
	if err != nil {
		return "echo 'Preview error'"
	}

	return scriptFile + " {}"
}

// generateRepoMapping generates shell case statements for repository path mapping
func (pg *PreviewGenerator) generateRepoMapping(repoMap map[string]string) string {
	var cases []string
	for alias, path := range repoMap {
		cases = append(cases, fmt.Sprintf("        %s)\n            repo_path=%s\n            ;;",
			shellEscape(alias), shellEscape(path)))
	}
	return strings.Join(cases, "\n")
}

// BuildCommitPreviewCommand builds a complete preview command for commit searching using secure internal command
func (pg *PreviewGenerator) BuildCommitPreviewCommand(repoMap map[string]string) string {
	script := fmt.Sprintf(`
#!/bin/bash
line="$1"
if [[ "$line" =~ ^([^:]+):[[:space:]]*([a-f0-9]+)[[:space:]]+(.+)$ ]]; then
    repo_alias="${BASH_REMATCH[1]}"
    commit_hash="${BASH_REMATCH[2]}"
    
    # Repository path mapping
    case "$repo_alias" in
%s
        *)
            echo "Unknown repository: $repo_alias"
            exit 1
            ;;
    esac
    
    # Create secure preview request for commit
    request=$(echo '{"type":"commit","path":"","repo_path":"'"$repo_path"'","line_number":0,"commit_hash":"'"$commit_hash"'"}' | base64 -w 0)
    
    # Call secure internal preview command
    gman internal-preview "$request"
else
    echo "Invalid format: $line"
fi
`, pg.generateRepoMapping(repoMap))

	scriptFile, _, err := CreateTempPreviewScript(script)
	if err != nil {
		return "echo 'Preview error'"
	}

	return scriptFile + " {}"
}

// CheckPreviewDependencies checks what preview tools are available
func (pg *PreviewGenerator) CheckPreviewDependencies() map[string]bool {
	tools := map[string]bool{
		"bat":       false,
		"chafa":     false,
		"catimg":    false,
		"pdftotext": false,
	}

	for tool := range tools {
		_, err := exec.LookPath(tool)
		tools[tool] = err == nil
	}

	return tools
}

// GetPreviewInstructions returns instructions for installing preview tools
func GetPreviewInstructions() string {
	return `For enhanced preview experience, consider installing these tools:

bat - Syntax highlighting for text files:
  macOS: brew install bat
  Ubuntu: apt install bat
  
chafa - Terminal image viewer:
  macOS: brew install chafa
  Ubuntu: apt install chafa
  
pdftotext - PDF text extraction:
  macOS: brew install poppler
  Ubuntu: apt install poppler-utils

catimg - Alternative image viewer:
  npm install -g catimg`
}
