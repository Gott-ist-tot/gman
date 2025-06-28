package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gman/internal/di"
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

		// Check if repositories are configured for work commands
		configMgr := di.ConfigManager()
		cfg := configMgr.GetConfig()
		if len(cfg.Repositories) == 0 {
			return fmt.Errorf("no repositories configured. Use 'gman repo add <alias> <path>' to add repositories")
		}
		return nil
	},
}


func init() {
	rootCmd.AddCommand(workCmd)

	// Add core safe workflow commands
	workCmd.AddCommand(statusCmd)
	workCmd.AddCommand(syncCmd)
}
