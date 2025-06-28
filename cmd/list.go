package cmd

import (
	"gman/internal/di"
	"gman/internal/display"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured repositories",
	Long: `List all repositories configured in gman with their aliases and paths.
This shows the mapping between repository aliases and their local filesystem paths.`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	// Command is now available via: gman repo list
	// Removed direct rootCmd registration to avoid duplication
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	// Configuration is already loaded by root command's PersistentPreRunE
	configMgr := di.ConfigManager()

	cfg := configMgr.GetConfig()
	display.PrintRepositoryList(cfg.Repositories)
	return nil
}
