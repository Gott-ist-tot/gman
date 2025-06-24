package cmd

import (
	"fmt"
	"sort"

	"gman/internal/config"
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
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		fmt.Println("No repositories configured. Use 'gman add' to add repositories.")
		return nil
	}

	// Sort repositories by alias
	type repoItem struct {
		alias string
		path  string
	}

	var repos []repoItem
	for alias, path := range cfg.Repositories {
		repos = append(repos, repoItem{alias: alias, path: path})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].alias < repos[j].alias
	})

	// Calculate maximum alias length for alignment
	maxAliasLen := 0
	for _, repo := range repos {
		if len(repo.alias) > maxAliasLen {
			maxAliasLen = len(repo.alias)
		}
	}

	// Display repositories
	fmt.Printf("Configured repositories (%d):\n\n", len(repos))
	for _, repo := range repos {
		fmt.Printf("  %-*s -> %s\n", maxAliasLen, repo.alias, repo.path)
	}

	return nil
}