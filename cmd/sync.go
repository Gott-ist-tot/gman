package cmd

import (
	"fmt"
	"sort"
	"sync"

	"gman/internal/config"
	"gman/internal/di"
	"gman/internal/progress"
	"gman/pkg/types"

	"github.com/spf13/cobra"
)

var (
	dryRun       bool
	showProgress bool
	groupName    string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Safely synchronize all repositories with their remotes",
	Long: `Safely synchronize all configured repositories with their remote origins.
This command performs git pull --ff-only on all repositories concurrently for maximum safety.

The sync operation will only succeed if the merge can be fast-forwarded,
preventing accidental merge commits and preserving repository history.

Options:
  --dry-run      : Show what would be synced without executing
  --progress     : Show detailed progress during sync operations
  --group        : Sync only repositories in the specified group

For more complex merge strategies, use native git commands in individual repositories.`,
	RunE: runSync,
}

func init() {
	// Command is now available via: gman work sync
	// Removed direct rootCmd registration to avoid duplication

	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without executing")
	syncCmd.Flags().BoolVar(&showProgress, "progress", false, "Show detailed progress during sync operations")
	syncCmd.Flags().StringVar(&groupName, "group", "", "Sync only repositories in the specified group")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load and validate configuration
	configMgr, cfg, err := validateAndLoadConfig()
	if err != nil {
		return err
	}

	// Determine which repositories to sync
	reposToSync, err := determineRepositoriesToSync(configMgr, cfg)
	if err != nil {
		return err
	}

	if len(reposToSync) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}

	// Handle dry-run mode
	if dryRun {
		return displayDryRunPreview(reposToSync)
	}

	// Execute the actual sync operations (always ff-only mode)
	results, err := executeSyncOperations(reposToSync, cfg)
	if err != nil {
		return err
	}

	// Display results and summary
	return displaySyncResults(results)
}

// Always use ff-only mode for safety
func getSyncMode() string {
	return "ff-only"
}

type syncResult struct {
	alias string
	path  string
	error error
}

// No filtering - sync all repositories for simplicity and consistency

// validateAndLoadConfig loads and validates the configuration
func validateAndLoadConfig() (*config.Manager, *types.Config, error) {
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		fmt.Println("No repositories configured. Use 'gman add' to add repositories.")
		return configMgr, cfg, nil
	}

	return configMgr, cfg, nil
}

// determineRepositoriesToSync determines which repositories to sync
func determineRepositoriesToSync(configMgr *config.Manager, cfg *types.Config) (map[string]string, error) {
	// Get repositories to sync (either all or group-specific)
	var reposToSync map[string]string
	var err error

	if groupName != "" {
		reposToSync, err = configMgr.GetGroupRepositories(groupName)
		if err != nil {
			return nil, fmt.Errorf("failed to get group repositories: %w", err)
		}
		if len(reposToSync) == 0 {
			fmt.Printf("Group '%s' has no repositories.\n", groupName)
			return nil, nil
		}
	} else {
		reposToSync = cfg.Repositories
	}

	return reposToSync, nil
}

// displayDryRunPreview shows what would be synced in dry-run mode
func displayDryRunPreview(reposToSync map[string]string) error {
	groupInfo := ""
	if groupName != "" {
		groupInfo = fmt.Sprintf(" from group '%s'", groupName)
	}
	fmt.Printf("DRY RUN: Would synchronize %d repositories%s (mode: ff-only):\n\n", len(reposToSync), groupInfo)
	for alias, path := range reposToSync {
		fmt.Printf("  %s → %s\n", alias, path)
	}
	return nil
}

// executeSyncOperations performs the actual sync operations across repositories
func executeSyncOperations(reposToSync map[string]string, cfg *types.Config) ([]syncResult, error) {
	// Setup progress tracking
	var progressBar *progress.MultiBar
	if showProgress {
		progressBar = progress.NewMultiBar()
		for alias := range reposToSync {
			progressBar.AddOperation(alias)
		}
	} else {
		groupInfo := ""
		if groupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", groupName)
		}
		fmt.Printf("Synchronizing %d repositories%s (mode: ff-only)...\n\n", len(reposToSync), groupInfo)
	}

	// Use a channel to collect results
	resultChan := make(chan syncResult, len(reposToSync))
	var wg sync.WaitGroup

	maxConcurrency := cfg.Settings.ParallelJobs
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	semaphore := make(chan struct{}, maxConcurrency)
	syncMgr := di.SyncManager()

	for alias, path := range reposToSync {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if progressBar != nil {
				progressBar.StartOperation(alias)
			}

			// Always use ff-only mode for safety
			err := syncMgr.SyncRepository(path, "ff-only")

			if progressBar != nil {
				progressBar.CompleteOperation(alias, err)
			}

			resultChan <- syncResult{
				alias: alias,
				path:  path,
				error: err,
			}
		}(alias, path)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []syncResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Finish progress display
	if progressBar != nil {
		progressBar.Finish()
	}

	// Sort results by alias for consistent output order
	sort.Slice(results, func(i, j int) bool {
		return results[i].alias < results[j].alias
	})

	return results, nil
}

// displaySyncResults displays the results and returns appropriate error if needed
func displaySyncResults(results []syncResult) error {
	var successCount, errorCount int
	
	// Count results
	for _, result := range results {
		if result.error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Display detailed results if not using progress mode
	if !showProgress {
		for _, result := range results {
			if result.error != nil {
				fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			} else {
				fmt.Printf("✅ %s: synced successfully\n", result.alias)
			}
		}
	}

	// Summary
	fmt.Printf("\nSync completed: %d successful, %d failed\n", successCount, errorCount)

	// Show failed repositories if any
	if errorCount > 0 && showProgress {
		fmt.Println("\nFailed repositories:")
		for _, result := range results {
			if result.error != nil {
				fmt.Printf("  ❌ %s: %v\n", result.alias, result.error)
			}
		}
		return fmt.Errorf("sync failed for %d repositories", errorCount)
	} else if errorCount > 0 {
		return fmt.Errorf("sync failed for %d repositories", errorCount)
	}

	return nil
}
