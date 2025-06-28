package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gman/internal/di"
)

// toolsCmd represents the tools and utilities command group
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tools and utilities",
	Long: `Advanced tools and utilities for power users.

This command group includes:
- real-time search capabilities
- task-oriented file management
- setup and configuration tools
- shell integration utilities
- system health diagnostics

Examples:
  gman tools find file config.yaml          # Search for files across repositories
  gman tools find commit "bug fix"          # Search commits across repositories
  gman tools find content "TODO"            # Search file content across repositories
  gman tools task create feature-auth       # Create task collection
  gman tools task list-files auth | xargs aider  # External tool integration
  gman tools health                          # System diagnostics and health check`,
	Aliases: []string{"t"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Call parent's PersistentPreRunE first to ensure config is loaded
		if cmd.Parent() != nil && cmd.Parent().PersistentPreRunE != nil {
			if err := cmd.Parent().PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}

		// Skip repository check during testing or if explicitly disabled
		if os.Getenv("GMAN_SKIP_REPO_CHECK") == "true" {
			return nil
		}

		// Only check repositories for commands that need them
		// Commands like 'setup', 'health', 'init' don't require existing repositories
		commandsNeedingRepos := map[string]bool{
			"find": true,
			"task": true,
		}

		if commandsNeedingRepos[cmd.Name()] {
			configMgr := di.ConfigManager()
			cfg := configMgr.GetConfig()
			if len(cfg.Repositories) == 0 {
				return fmt.Errorf("no repositories configured. Use 'gman repo add <alias> <path>' to add repositories")
			}
		}
		return nil
	},
}


func init() {
	rootCmd.AddCommand(toolsCmd)

	// Add original commands directly to tools group to preserve subcommands
	toolsCmd.AddCommand(findCmd)
	toolsCmd.AddCommand(setupCmd)
	toolsCmd.AddCommand(completionCmd)
	toolsCmd.AddCommand(initCmd)
	toolsCmd.AddCommand(healthCmd)
	toolsCmd.AddCommand(taskCmd)

	// Add onboarding as subcommand (for advanced users who want direct access)
	toolsCmd.AddCommand(onboardingCmd)
}
