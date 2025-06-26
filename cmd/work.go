package cmd

import (
	"gman/cmd/batch"

	"github.com/spf13/cobra"
)

// workCmd represents the Git workflow command group
var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Git workflow commands",
	Long: `Git workflow commands for day-to-day development operations.

This command group includes:
- status checking across repositories
- synchronization with remotes  
- branch management
- commit and push operations
- stash management

Examples:
  gman work status                     # Check status of all repositories
  gman work sync                       # Sync all repositories
  gman work branch list               # List branches across repositories
  gman work commit -m "Fix bug"       # Commit across repositories`,
	Aliases: []string{"w"},
}


func init() {
	rootCmd.AddCommand(workCmd)

	// Add original commands directly to work group to preserve subcommands and functionality
	workCmd.AddCommand(statusCmd)
	workCmd.AddCommand(syncCmd)
	workCmd.AddCommand(batch.NewCommitCmd())
	workCmd.AddCommand(batch.NewPushCmd())
	workCmd.AddCommand(batch.NewPullCmd())
	workCmd.AddCommand(batch.NewStashCmd())

	// Add branch and diff as subcommands (they're already defined)
	workCmd.AddCommand(branchCmd)
	workCmd.AddCommand(diffCmd)
}
