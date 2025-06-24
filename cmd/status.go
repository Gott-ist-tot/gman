package cmd

import (
	"fmt"
	"sort"

	"gman/internal/config"
	"gman/internal/git"
	"gman/pkg/types"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all repositories",
	Long: `Display the current status of all configured repositories including:
- Current branch
- Workspace status (clean/dirty/stashed)
- Sync status with remote (ahead/behind/up-to-date)
- Last commit information`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		fmt.Println("No repositories configured. Use 'gman add' to add repositories.")
		return nil
	}

	// Get status for all repositories
	gitMgr := git.NewManager()
	statuses, err := gitMgr.GetAllRepoStatus(cfg.Repositories)
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	// Sort by alias for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Alias < statuses[j].Alias
	})

	// Display results
	displayStatus(statuses, cfg.Settings.ShowLastCommit)
	return nil
}

func displayStatus(statuses []types.RepoStatus, showLastCommit bool) {
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
		if len(status.Workspace.String()) > maxWorkspace {
			maxWorkspace = len(status.Workspace.String())
		}
		if len(status.SyncStatus.String()) > maxSync {
			maxSync = len(status.SyncStatus.String())
		}
		if showLastCommit && len(status.LastCommit) > maxCommit {
			maxCommit = len(status.LastCommit)
		}
	}

	// Add some padding
	maxAlias += 2
	maxBranch += 2
	maxWorkspace += 2
	maxSync += 2
	if showLastCommit {
		maxCommit += 2
	}

	// Print header
	fmt.Printf("%-*s %-*s %-*s %-*s", maxAlias, "Alias", maxBranch, "Branch", maxWorkspace, "Workspace", maxSync, "Sync Status")
	if showLastCommit {
		fmt.Printf(" %-*s", maxCommit, "Last Commit")
	}
	fmt.Println()

	// Print separator
	fmt.Printf("%s %s %s %s", 
		repeatString("-", maxAlias), 
		repeatString("-", maxBranch), 
		repeatString("-", maxWorkspace), 
		repeatString("-", maxSync))
	if showLastCommit {
		fmt.Printf(" %s", repeatString("-", maxCommit))
	}
	fmt.Println()

	// Print repository status
	for _, status := range statuses {
		if status.Error != nil {
			fmt.Printf("%-*s %-*s %-*s %-*s", 
				maxAlias, formatAlias(status.Alias, status.IsCurrent),
				maxBranch, "ERROR",
				maxWorkspace, status.Error.Error(),
				maxSync, "")
			if showLastCommit {
				fmt.Printf(" %-*s", maxCommit, "")
			}
			fmt.Println()
			continue
		}

		fmt.Printf("%-*s %-*s %-*s %-*s", 
			maxAlias, formatAlias(status.Alias, status.IsCurrent),
			maxBranch, status.Branch,
			maxWorkspace, status.Workspace.String(),
			maxSync, status.SyncStatus.String())
		
		if showLastCommit {
			fmt.Printf(" %-*s", maxCommit, truncateString(status.LastCommit, maxCommit-2))
		}
		fmt.Println()
	}
}

func formatAlias(alias string, isCurrent bool) string {
	if isCurrent {
		return "* " + alias
	}
	return "  " + alias
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}