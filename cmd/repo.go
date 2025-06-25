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

// repoAddCmd adds a repository (alias for existing add command)
var repoAddCmd = &cobra.Command{
	Use:   "add <alias> <path>",
	Short: "Add a repository to gman configuration",
	Long: `Add a Git repository to gman configuration with a friendly alias.

The alias should be short and memorable for easy switching.
The path must point to a valid Git repository.

Examples:
  gman repo add frontend /Users/john/projects/my-frontend
  gman repo add api ~/work/backend-api
  gman repo add . myproject  # Add current directory as 'myproject'`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd, // Reuse existing add command logic
	Aliases: []string{"a"},
}

// repoRemoveCmd removes a repository (alias for existing remove command)
var repoRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a repository from gman configuration",
	Long: `Remove a repository from gman configuration.

This only removes the repository from gman's tracking.
The actual repository files are not affected.

Examples:
  gman repo remove frontend
  gman repo remove api`,
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove, // Reuse existing remove command logic
	Aliases: []string{"rm", "delete", "del"},
}

// repoListCmd lists repositories (alias for existing list command)
var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured repositories",
	Long: `List all repositories currently managed by gman.

Shows repository aliases, paths, and optionally additional information
like last commit details and current status.

Examples:
  gman repo list                    # Basic list
  gman repo list --extended         # Show detailed information`,
	RunE:    runList, // Reuse existing list command logic
	Aliases: []string{"ls"},
}

// repoRecentCmd shows recent repositories (alias for existing recent command)
var repoRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show recently accessed repositories",
	Long: `Show recently accessed repositories ordered by last access time.

This helps you quickly return to repositories you've been working on.

Examples:
  gman repo recent              # Show recent repositories
  gman repo recent --limit 5    # Show only 5 most recent`,
	RunE:    runRecent, // Reuse existing recent command logic
	Aliases: []string{"r"},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	
	// Add subcommands to repo group
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoRecentCmd)
	
	// Add the group command as a subcommand (it's already defined)
	repoCmd.AddCommand(groupCmd)
	
	// Copy flags from original commands
	copyCommandFlags(repoListCmd, listCmd)
	copyCommandFlags(repoRecentCmd, recentCmd)
}

// copyCommandFlags copies flags from source command to destination
func copyCommandFlags(dst, src *cobra.Command) {
	src.Flags().VisitAll(func(flag *cobra.Flag) {
		dst.Flags().AddFlag(flag)
	})
}