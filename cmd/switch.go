package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/interactive"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [alias]",
	Short: "Switch to a repository directory",
	Long: `Switch to the directory of the specified repository.
If no alias is provided, an interactive menu will be displayed.
This command outputs a special format that the shell wrapper function
can use to change the current working directory.

Examples:
  gman switch my-repo       # Switch to 'my-repo'
  gman switch proj          # Fuzzy match repositories containing 'proj'
  gman switch               # Interactive selection menu`,
	Args: cobra.RangeArgs(0, 1),
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
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	var alias string
	var err error

	if len(args) == 0 {
		// Interactive mode
		selector := interactive.NewRepositorySelector(cfg.Repositories)
		alias, err = selector.SelectRepository()
		if err != nil {
			return err
		}
	} else {
		// Direct alias or fuzzy match
		inputAlias := args[0]
		
		// Check for exact match first
		if _, exists := cfg.Repositories[inputAlias]; exists {
			alias = inputAlias
		} else {
			// Try fuzzy matching
			alias, err = fuzzyMatchRepository(inputAlias, cfg.Repositories)
			if err != nil {
				return err
			}
		}
	}

	targetPath := cfg.Repositories[alias]
	
	// Track recent usage
	if err := configMgr.TrackRecentUsage(alias); err != nil {
		// Don't fail the switch if tracking fails, just log it silently
		// Could add debug logging here in the future
	}

	// Output special format for shell wrapper to handle
	fmt.Printf("GMAN_CD:%s", targetPath)
	return nil
}

// fuzzyMatchRepository performs fuzzy matching on repository aliases
func fuzzyMatchRepository(input string, repos map[string]string) (string, error) {
	input = strings.ToLower(input)
	var matches []string

	for alias := range repos {
		if strings.Contains(strings.ToLower(alias), input) {
			matches = append(matches, alias)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no repositories found matching '%s'", input)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	return "", fmt.Errorf("multiple repositories match '%s': %s. Please be more specific", 
		input, strings.Join(matches, ", "))
}
