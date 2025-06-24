package cmd

import (
	"fmt"
	"sync"

	"gman/internal/config"
	"gman/internal/git"
	"gman/pkg/types"
	"github.com/spf13/cobra"
)

var (
	syncMode     string
	syncRebase   bool
	syncAutostash bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize all repositories with their remotes",
	Long: `Synchronize all configured repositories with their remote origins.
This command runs git pull on all repositories concurrently.

Sync modes:
  ff-only    : Only fast-forward merges (default, safest)
  rebase     : Use git pull --rebase
  autostash  : Use git pull --autostash
`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVar(&syncMode, "mode", "", "Sync mode: ff-only, rebase, autostash")
	syncCmd.Flags().BoolVar(&syncRebase, "rebase", false, "Use git pull --rebase")
	syncCmd.Flags().BoolVar(&syncAutostash, "autostash", false, "Use git pull --autostash")
}

func runSync(cmd *cobra.Command, args []string) error {
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

	// Determine sync mode
	mode := determineSyncMode(cfg)

	fmt.Printf("Synchronizing %d repositories (mode: %s)...\n\n", len(cfg.Repositories), mode)

	// Sync all repositories
	gitMgr := git.NewManager()
	
	// Use a channel to collect results
	resultChan := make(chan syncResult, len(cfg.Repositories))
	var wg sync.WaitGroup

	maxConcurrency := cfg.Settings.ParallelJobs
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	semaphore := make(chan struct{}, maxConcurrency)

	for alias, path := range cfg.Repositories {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			err := gitMgr.SyncRepository(path, mode)
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

	// Display results as they come in
	var successCount, errorCount int
	for result := range resultChan {
		if result.error != nil {
			fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			errorCount++
		} else {
			fmt.Printf("✅ %s: synced successfully\n", result.alias)
			successCount++
		}
	}

	// Summary
	fmt.Printf("\nSync completed: %d successful, %d failed\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("sync failed for %d repositories", errorCount)
	}

	return nil
}

func determineSyncMode(cfg *types.Config) string {
	// Command line flags take precedence
	if syncRebase {
		return "rebase"
	}
	if syncAutostash {
		return "autostash"
	}
	if syncMode != "" {
		return syncMode
	}

	// Use config default
	if cfg.Settings.DefaultSyncMode != "" {
		return cfg.Settings.DefaultSyncMode
	}

	// Default to ff-only (safest)
	return "ff-only"
}

type syncResult struct {
	alias string
	path  string
	error error
}