package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/display"
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
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	display.PrintRepositoryList(cfg.Repositories)
	return nil
}
