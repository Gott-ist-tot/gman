package display

import (
	"fmt"
	"strings"

	"gman/pkg/types"
	"github.com/fatih/color"
)

// StatusDisplayer handles displaying repository status in a nice format
type StatusDisplayer struct {
	showLastCommit bool
}

// NewStatusDisplayer creates a new status displayer
func NewStatusDisplayer(showLastCommit bool) *StatusDisplayer {
	return &StatusDisplayer{
		showLastCommit: showLastCommit,
	}
}

// Display shows the repository status in a formatted table
func (d *StatusDisplayer) Display(statuses []types.RepoStatus) {
	if len(statuses) == 0 {
		fmt.Println("No repositories to display.")
		return
	}

	// Calculate column widths
	maxAlias := len("Alias")
	maxBranch := len("Branch")
	maxWorkspace := len("Workspace")
	maxSync := len("Sync Status")
	maxCommit := len("Last Commit")

	for _, status := range statuses {
		if len(status.Alias) > maxAlias {
			maxAlias = len(status.Alias)
		}
		if len(status.Branch) > maxBranch {
			maxBranch = len(status.Branch)
		}
		workspaceStr := stripAnsiCodes(status.Workspace.String())
		if len(workspaceStr) > maxWorkspace {
			maxWorkspace = len(workspaceStr)
		}
		syncStr := stripAnsiCodes(status.SyncStatus.String())
		if len(syncStr) > maxSync {
			maxSync = len(syncStr)
		}
		if d.showLastCommit && len(status.LastCommit) > maxCommit {
			maxCommit = len(status.LastCommit)
		}
	}

	// Add padding
	maxAlias += 3
	maxBranch += 2
	maxWorkspace += 2
	maxSync += 2
	if d.showLastCommit {
		maxCommit += 2
	}

	// Print header with colors
	fmt.Printf("%-*s %-*s %-*s %-*s", 
		maxAlias, color.CyanString("Alias"), 
		maxBranch, color.CyanString("Branch"), 
		maxWorkspace, color.CyanString("Workspace"), 
		maxSync, color.CyanString("Sync Status"))
	if d.showLastCommit {
		fmt.Printf(" %-*s", maxCommit, color.CyanString("Last Commit"))
	}
	fmt.Println()

	// Print separator
	fmt.Printf("%s %s %s %s", 
		strings.Repeat("─", maxAlias), 
		strings.Repeat("─", maxBranch), 
		strings.Repeat("─", maxWorkspace), 
		strings.Repeat("─", maxSync))
	if d.showLastCommit {
		fmt.Printf(" %s", strings.Repeat("─", maxCommit))
	}
	fmt.Println()

	// Print repository status
	for _, status := range statuses {
		if status.Error != nil {
			fmt.Printf("%-*s %-*s %-*s %-*s", 
				maxAlias, d.formatAlias(status.Alias, status.IsCurrent),
				maxBranch, color.RedString("ERROR"),
				maxWorkspace, color.RedString(truncateString(status.Error.Error(), maxWorkspace-2)),
				maxSync, "")
			if d.showLastCommit {
				fmt.Printf(" %-*s", maxCommit, "")
			}
			fmt.Println()
			continue
		}

		fmt.Printf("%-*s %-*s %-*s %-*s", 
			maxAlias, d.formatAlias(status.Alias, status.IsCurrent),
			maxBranch, d.formatBranch(status.Branch),
			maxWorkspace, status.Workspace.String(),
			maxSync, status.SyncStatus.String())
		
		if d.showLastCommit {
			fmt.Printf(" %-*s", maxCommit, d.formatCommit(status.LastCommit))
		}
		fmt.Println()
	}

	fmt.Println() // Add empty line at the end
}

// formatAlias formats the alias with current indicator
func (d *StatusDisplayer) formatAlias(alias string, isCurrent bool) string {
	if isCurrent {
		return color.YellowString("* %s", alias)
	}
	return fmt.Sprintf("  %s", alias)
}

// formatBranch formats the branch name
func (d *StatusDisplayer) formatBranch(branch string) string {
	return color.CyanString(branch)
}

// formatCommit formats the commit message
func (d *StatusDisplayer) formatCommit(commit string) string {
	if len(commit) > 50 {
		return commit[:47] + "..."
	}
	return commit
}

// PrintRepositoryList displays the repository list in a formatted way
func PrintRepositoryList(repositories map[string]string) {
	if len(repositories) == 0 {
		fmt.Println("No repositories configured. Use 'gman add' to add repositories.")
		return
	}

	fmt.Printf("\n%s\n\n", color.GreenString("Configured repositories (%d):", len(repositories)))

	// Calculate max alias width for alignment
	maxAliasLen := 0
	for alias := range repositories {
		if len(alias) > maxAliasLen {
			maxAliasLen = len(alias)
		}
	}

	// Print header
	fmt.Printf("%-*s   %s\n", maxAliasLen, color.CyanString("Alias"), color.CyanString("Path"))
	fmt.Printf("%s   %s\n", strings.Repeat("─", maxAliasLen), strings.Repeat("─", 40))

	// Print repositories
	for alias, path := range repositories {
		fmt.Printf("%-*s → %s\n", maxAliasLen, color.YellowString(alias), path)
	}
	fmt.Println()
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", color.GreenString("✅"), message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("%s %s\n", color.RedString("❌"), message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", color.YellowString("⚠️"), message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", color.BlueString("ℹ️"), message)
}

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// stripAnsiCodes removes ANSI color codes for length calculation
func stripAnsiCodes(s string) string {
	// Simple approach: remove common ANSI sequences
	// This is a basic implementation, might need improvement for complex cases
	result := s
	for i := 0; i < len(result); i++ {
		if result[i] == '\033' || result[i] == '\x1b' {
			// Find the end of the ANSI sequence
			j := i + 1
			for j < len(result) && (result[j] < 'A' || result[j] > 'Z') && (result[j] < 'a' || result[j] > 'z') {
				j++
			}
			if j < len(result) {
				j++ // Include the final character
			}
			// Remove the ANSI sequence
			result = result[:i] + result[j:]
			i-- // Recheck this position
		}
	}
	return result
}