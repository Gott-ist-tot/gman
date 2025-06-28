package cmd

import (
	"github.com/spf13/cobra"
)

// workCmd represents the Git workflow command group
var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Git workflow commands",
	Long: `Git workflow commands for safe, read-focused development operations.

This command group includes:
- status checking across repositories
- synchronization with remotes (safe pulls only)

Examples:
  gman work status                     # Check status of all repositories
  gman work sync                       # Safe sync all repositories with ff-only`,
	Aliases: []string{"w"},
}


func init() {
	rootCmd.AddCommand(workCmd)

	// Add core safe workflow commands
	workCmd.AddCommand(statusCmd)
	workCmd.AddCommand(syncCmd)
}
