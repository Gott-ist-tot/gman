package cmd

import (
	"github.com/spf13/cobra"
)

// quickCmd represents quick access to most common commands
var quickCmd = &cobra.Command{
	Use:   "quick",
	Short: "Quick access to common operations",
	Long: `Quick access to the most commonly used gman operations.

This provides shortcuts to frequently used commands without
having to remember the full command structure.

Examples:
  gman quick status                    # Quick status check
  gman quick sync                      # Quick sync
  gman quick switch                    # Quick repository switching
  gman quick add myrepo /path          # Quick repository addition`,
	Aliases: []string{"q"},
}

// Create quick access commands that are aliases to the original commands
var quickStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Quick status check",
	RunE:    runStatus,
	Aliases: []string{"st"},
}

var quickSyncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Quick sync all repositories",
	RunE:    runSync,
	Aliases: []string{"s"},
}

var quickSwitchCmd = &cobra.Command{
	Use:     "switch [repository]",
	Short:   "Quick repository switching",
	RunE:    runSwitch,
	Aliases: []string{"sw", "cd"},
}

var quickListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Quick repository list",
	RunE:    runList,
	Aliases: []string{"ls"},
}

var quickAddCmd = &cobra.Command{
	Use:     "add <alias> <path>",
	Short:   "Quick add repository",
	Args:    cobra.ExactArgs(2),
	RunE:    runAdd,
	Aliases: []string{"a"},
}

func init() {
	rootCmd.AddCommand(quickCmd)

	// Add quick commands
	quickCmd.AddCommand(quickStatusCmd)
	quickCmd.AddCommand(quickSyncCmd)
	quickCmd.AddCommand(quickSwitchCmd)
	quickCmd.AddCommand(quickListCmd)
	quickCmd.AddCommand(quickAddCmd)

	// Copy flags from original commands
	copyCommandFlags(quickStatusCmd, statusCmd)
	copyCommandFlags(quickSyncCmd, syncCmd)
	copyCommandFlags(quickSwitchCmd, switchCmd)
	copyCommandFlags(quickListCmd, listCmd)
}
