package cmd

import (
	"github.com/spf13/cobra"
)

// toolsCmd represents the tools and utilities command group
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tools and utilities",
	Long: `Advanced tools and utilities for power users.

This command group includes:
- search and indexing capabilities
- worktree management  
- interactive dashboard
- setup and configuration tools
- shell integration utilities

Examples:
  gman tools find file config.yaml          # Search for files across repositories
  gman tools find commit "bug fix"          # Search commits across repositories
  gman tools index rebuild                  # Rebuild search index
  gman tools dashboard                       # Launch interactive TUI
  gman tools worktree add backend feature-auth  # Create worktree`,
	Aliases: []string{"t"},
}


func init() {
	rootCmd.AddCommand(toolsCmd)

	// Add original commands directly to tools group to preserve subcommands
	toolsCmd.AddCommand(findCmd)
	toolsCmd.AddCommand(indexCmd)
	toolsCmd.AddCommand(dashboardCmd)
	toolsCmd.AddCommand(worktreeCmd)
	toolsCmd.AddCommand(setupCmd)
	toolsCmd.AddCommand(completionCmd)

	// Add onboarding as subcommand (for advanced users who want direct access)
	toolsCmd.AddCommand(onboardingCmd)
}
