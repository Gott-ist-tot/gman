package cmd

import (
	"fmt"
	"sort"

	"gman/internal/di"
	"gman/internal/display"

	"github.com/spf13/cobra"
)

var verboseStatus bool

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all repositories",
	Long: `Display the current status of all configured repositories including:
- Current branch
- Workspace status (clean/dirty/stashed)
- Sync status with remote (ahead/behind/up-to-date)
- Last commit information

Use --verbose to see detailed information including file change counts, commit times,
remote URLs, stash counts, and branch statistics.`,
	RunE: runStatus,
}

func init() {
	// Command is now available via: gman work status
	// Removed direct rootCmd registration to avoid duplication
	statusCmd.Flags().BoolVarP(&verboseStatus, "verbose", "v", false, "Show detailed information (file changes, commit times, remote URLs, stash counts)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Configuration is already loaded by PersistentPreRunE
	// Repository check is already done by work group's PersistentPreRunE
	configMgr := di.ConfigManager()
	cfg := configMgr.GetConfig()

	// Get status for all repositories
	gitMgr := di.GitManager()
	statuses, err := gitMgr.GetAllRepoStatus(cfg.Repositories)
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	// Sort by alias for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Alias < statuses[j].Alias
	})

	// Display results
	var displayer *display.StatusDisplayer
	if verboseStatus {
		displayer = display.NewSuperExtendedStatusDisplayer(cfg.Settings.ShowLastCommit)
	} else {
		displayer = display.NewStatusDisplayer(cfg.Settings.ShowLastCommit)
	}
	displayer.Display(statuses)
	return nil
}
