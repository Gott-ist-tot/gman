package cmd

import (
	"fmt"

	"gman/internal/config"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a repository from gman configuration",
	Long: `Remove a repository from gman configuration.
This only removes the repository from gman's tracking list.
The actual repository files on disk are not affected.`,
	Aliases: []string{"rm", "del", "delete"},
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Load config and return repository aliases for completion
		configMgr := config.NewManager()
		if err := configMgr.Load(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cfg := configMgr.GetConfig()
		var aliases []string
		for alias := range cfg.Repositories {
			aliases = append(aliases, alias)
		}
		return aliases, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	alias := args[0]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get the path before removing (for display)
	cfg := configMgr.GetConfig()
	path, exists := cfg.Repositories[alias]
	if !exists {
		return fmt.Errorf("repository '%s' not found", alias)
	}

	// Remove repository
	if err := configMgr.RemoveRepository(alias); err != nil {
		return err
	}

	fmt.Printf("Removed repository: %s (%s)\n", alias, path)
	return nil
}