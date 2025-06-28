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
- real-time search capabilities
- interactive dashboard
- setup and configuration tools
- shell integration utilities
- system health diagnostics

Examples:
  gman tools find file config.yaml          # Search for files across repositories
  gman tools find commit "bug fix"          # Search commits across repositories
  gman tools find content "TODO"            # Search file content across repositories
  gman tools dashboard                       # Launch interactive TUI
  gman tools health                          # System diagnostics and health check`,
	Aliases: []string{"t"},
}


func init() {
	rootCmd.AddCommand(toolsCmd)

	// Add original commands directly to tools group to preserve subcommands
	toolsCmd.AddCommand(findCmd)
	toolsCmd.AddCommand(dashboardCmd)
	toolsCmd.AddCommand(setupCmd)
	toolsCmd.AddCommand(completionCmd)
	toolsCmd.AddCommand(initCmd)
	toolsCmd.AddCommand(healthCmd)

	// Add onboarding as subcommand (for advanced users who want direct access)
	toolsCmd.AddCommand(onboardingCmd)
}
