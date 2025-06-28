package cmd

import (
	"github.com/spf13/cobra"
)

// repoCmd represents the repository management command group
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository management commands",
	Long: `Repository management commands for adding, removing, and organizing repositories.

This command group includes:
- add/remove repositories
- list and view repository information
- organize repositories into groups
- view recent activity

Examples:
  gman repo add myproject /path/to/project     # Add a repository
  gman repo list                               # List all repositories  
  gman repo remove myproject                   # Remove a repository
  gman repo group create webdev frontend backend  # Create repository group`,
	Aliases: []string{"r"},
}

func init() {
	rootCmd.AddCommand(repoCmd)

	// Add original commands directly to repo group to preserve all functionality
	// This follows the same pattern as work.go and ensures ValidArgsFunction and flags are preserved
	repoCmd.AddCommand(addCmd)    // from cmd/add.go
	repoCmd.AddCommand(removeCmd) // from cmd/remove.go (includes ValidArgsFunction for alias completion)
	repoCmd.AddCommand(listCmd)   // from cmd/list.go
	repoCmd.AddCommand(recentCmd) // from cmd/recent.go (includes --limit flag)
	repoCmd.AddCommand(groupCmd)  // from cmd/group.go

	// No need for copyCommandFlags as we're using original commands with their flags intact
}
