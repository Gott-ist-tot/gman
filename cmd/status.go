package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/display"
	"gman/internal/git"
)

var extendedStatus bool

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all repositories",
	Long: `Display the current status of all configured repositories including:
- Current branch
- Workspace status (clean/dirty/stashed)
- Sync status with remote (ahead/behind/up-to-date)
- Last commit information

Use --extended to see additional information like file change counts and commit times.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&extendedStatus, "extended", "e", false, "Show extended information (file changes, commit times)")
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
	var displayer *display.StatusDisplayer
	if extendedStatus {
		displayer = display.NewExtendedStatusDisplayer(cfg.Settings.ShowLastCommit)
	} else {
		displayer = display.NewStatusDisplayer(cfg.Settings.ShowLastCommit)
	}
	displayer.Display(statuses)
	return nil
}
