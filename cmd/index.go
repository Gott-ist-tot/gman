package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/index"
	"github.com/fatih/color"
)

var (
	indexForce bool
	indexStats bool
	indexQuiet bool
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage search index for fast file and commit searching",
	Long: `Manage the search index used by the 'gman find' commands.
The index stores file listings and commit history from all configured
repositories to enable fast searching with fzf.

Examples:
  gman index --rebuild     # Force rebuild the entire index
  gman index --stats       # Show index statistics
  gman index update        # Update index for all repositories`,
}

// indexRebuildCmd rebuilds the search index
var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the search index from scratch",
	Long: `Completely rebuild the search index for all configured repositories.
This will remove all existing index data and scan all repositories
to build a fresh index. This may take some time for large repositories.

Use this command if:
- The index seems corrupted or incomplete
- You want to ensure all recent changes are indexed
- You've made significant changes to repository structure`,
	RunE: runIndexRebuild,
}

// indexUpdateCmd updates the search index
var indexUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the search index incrementally",
	Long: `Update the search index incrementally for all configured repositories.
This is faster than a full rebuild and is automatically performed
when you run other gman commands.

This command is useful if you want to manually update the index
after making changes to your repositories.`,
	RunE: runIndexUpdate,
}

// indexStatsCmd shows index statistics
var indexStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show search index statistics",
	Long: `Display statistics about the current search index, including:
- Number of indexed files and commits
- Number of repositories indexed
- Index database size
- Last update time`,
	RunE: runIndexStats,
}

// indexClearCmd clears the search index
var indexClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the search index",
	Long: `Remove all data from the search index. This will require
rebuilding the index before 'gman find' commands will work.

Use this command if you want to start fresh or if the index
is taking up too much disk space.`,
	RunE: runIndexClear,
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.AddCommand(indexRebuildCmd)
	indexCmd.AddCommand(indexUpdateCmd)
	indexCmd.AddCommand(indexStatsCmd)
	indexCmd.AddCommand(indexClearCmd)

	// Add flags
	indexCmd.PersistentFlags().BoolVar(&indexQuiet, "quiet", false, "Suppress progress output")
	indexRebuildCmd.Flags().BoolVar(&indexForce, "force", false, "Force rebuild without confirmation")
	indexClearCmd.Flags().BoolVar(&indexForce, "force", false, "Force clear without confirmation")
}

func runIndexRebuild(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Confirm destructive operation unless forced
	if !indexForce {
		fmt.Printf("%s This will completely rebuild the search index.\n", color.YellowString("⚠️"))
		fmt.Printf("All existing index data will be removed. Continue? [y/N]: ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation canceled.")
			return nil
		}
	}

	// Initialize indexer
	indexer, err := index.NewIndexer(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize indexer: %w", err)
	}
	defer indexer.Close()

	// Progress tracking
	var progressFunc func(string, int, int)
	if !indexQuiet {
		progressFunc = func(message string, current, total int) {
			if total > 0 {
				percent := float64(current) / float64(total) * 100
				fmt.Fprintf(os.Stderr, "\r%s [%3.0f%%] %s", 
					color.BlueString("🔄"), percent, message)
			} else {
				fmt.Fprintf(os.Stderr, "\r%s %s", color.BlueString("🔄"), message)
			}
		}
	}

	start := time.Now()
	
	if !indexQuiet {
		fmt.Printf("%s Starting index rebuild for %d repositories...\n", 
			color.GreenString("🔧"), len(cfg.Repositories))
	}

	// Rebuild the index
	err = indexer.RebuildIndex(cfg.Repositories, progressFunc)
	if err != nil {
		if !indexQuiet {
			fmt.Fprintf(os.Stderr, "\n")
		}
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	if !indexQuiet {
		fmt.Fprintf(os.Stderr, "\n")
		duration := time.Since(start)
		fmt.Printf("%s Index rebuild completed in %v\n", 
			color.GreenString("✅"), duration.Round(time.Second))
		
		// Show stats
		return showIndexStats(indexer)
	}

	return nil
}

func runIndexUpdate(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Initialize indexer
	indexer, err := index.NewIndexer(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize indexer: %w", err)
	}
	defer indexer.Close()

	start := time.Now()
	
	if !indexQuiet {
		fmt.Printf("%s Updating index for %d repositories...\n", 
			color.BlueString("🔄"), len(cfg.Repositories))
	}

	// Update the index
	err = indexer.UpdateIndex(cfg.Repositories)
	if err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	if !indexQuiet {
		duration := time.Since(start)
		fmt.Printf("%s Index update completed in %v\n", 
			color.GreenString("✅"), duration.Round(time.Second))
		
		// Show stats
		return showIndexStats(indexer)
	}

	return nil
}

func runIndexStats(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize indexer
	indexer, err := index.NewIndexer(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize indexer: %w", err)
	}
	defer indexer.Close()

	return showIndexStats(indexer)
}

func runIndexClear(cmd *cobra.Command, args []string) error {
	// Confirm destructive operation unless forced
	if !indexForce {
		fmt.Printf("%s This will remove all search index data.\n", color.YellowString("⚠️"))
		fmt.Printf("You will need to rebuild the index before using 'gman find'. Continue? [y/N]: ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation canceled.")
			return nil
		}
	}

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()

	// Initialize indexer
	indexer, err := index.NewIndexer(configMgr)
	if err != nil {
		return fmt.Errorf("failed to initialize indexer: %w", err)
	}
	defer indexer.Close()

	// Clear index for all repositories
	storage := indexer.GetStorage()
	for alias := range cfg.Repositories {
		if err := storage.ClearRepository(alias); err != nil {
			return fmt.Errorf("failed to clear index for %s: %w", alias, err)
		}
	}

	fmt.Printf("%s Search index cleared successfully\n", color.GreenString("✅"))
	fmt.Printf("%s Run 'gman index rebuild' to rebuild the index\n", color.BlueString("💡"))

	return nil
}

func showIndexStats(indexer *index.Indexer) error {
	stats, err := indexer.GetStorage().GetStats()
	if err != nil {
		return fmt.Errorf("failed to get index statistics: %w", err)
	}

	fmt.Printf("\n%s:\n", color.CyanString("Search Index Statistics"))
	fmt.Println(strings.Repeat("─", 30))

	// File count
	if fileCount, ok := stats["file_count"].(int); ok {
		fmt.Printf("Files indexed:      %s\n", color.GreenString("%d", fileCount))
	}

	// Commit count
	if commitCount, ok := stats["commit_count"].(int); ok {
		fmt.Printf("Commits indexed:    %s\n", color.GreenString("%d", commitCount))
	}

	// Repository count
	if repoCount, ok := stats["repository_count"].(int); ok {
		fmt.Printf("Repositories:       %s\n", color.GreenString("%d", repoCount))
	}

	// Database size
	if dbSize, ok := stats["db_size"].(int64); ok {
		fmt.Printf("Database size:      %s\n", color.BlueString(formatBytes(dbSize)))
	}

	fmt.Println()
	return nil
}

// formatBytes formats a byte count in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

