package cmd

import (
	"fmt"

	"gman/internal/config"
	"github.com/spf13/cobra"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch <alias>",
	Short: "Switch to a repository directory",
	Long: `Switch to the directory of the specified repository.
This command outputs a special format that the shell wrapper function
can use to change the current working directory.`,
	Args: cobra.ExactArgs(1),
	RunE: runSwitch,
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
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	alias := args[0]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	targetPath, exists := cfg.Repositories[alias]
	if !exists {
		return fmt.Errorf("repository '%s' not found", alias)
	}

	// Output special format for shell wrapper to handle
	fmt.Printf("GMAN_CD:%s", targetPath)
	return nil
}