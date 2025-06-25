package cmd

import (
	"fmt"
	"sync"

	"gman/internal/di"
	"gman/internal/git"
	"gman/internal/progress"
	"gman/pkg/types"

	"github.com/spf13/cobra"
)

var (
	syncMode      string
	syncRebase    bool
	syncAutostash bool
	onlyDirty     bool
	onlyBehind    bool
	onlyAhead     bool
	dryRun        bool
	showProgress  bool
	groupName     string
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

Conditional sync options:
  --only-dirty   : Sync only repositories with uncommitted changes
  --only-behind  : Sync only repositories that are behind remote
  --only-ahead   : Sync only repositories with unpushed commits
  --dry-run      : Show what would be synced without executing
  --progress     : Show detailed progress during sync operations
  --group        : Sync only repositories in the specified group
`,
	RunE: runSync,
}

func init() {
	// Command is now available via: gman work sync
	// Removed direct rootCmd registration to avoid duplication

	syncCmd.Flags().StringVar(&syncMode, "mode", "", "Sync mode: ff-only, rebase, autostash")
	syncCmd.Flags().BoolVar(&syncRebase, "rebase", false, "Use git pull --rebase")
	syncCmd.Flags().BoolVar(&syncAutostash, "autostash", false, "Use git pull --autostash")

	// Conditional sync options
	syncCmd.Flags().BoolVar(&onlyDirty, "only-dirty", false, "Sync only repositories with uncommitted changes")
	syncCmd.Flags().BoolVar(&onlyBehind, "only-behind", false, "Sync only repositories that are behind remote")
	syncCmd.Flags().BoolVar(&onlyAhead, "only-ahead", false, "Sync only repositories with unpushed commits")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without executing")
	syncCmd.Flags().BoolVar(&showProgress, "progress", false, "Show detailed progress during sync operations")
	syncCmd.Flags().StringVar(&groupName, "group", "", "Sync only repositories in the specified group")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		fmt.Println("No repositories configured. Use 'gman add' to add repositories.")
		return nil
	}

	// Get repositories to sync (either all or group-specific)
	var reposToSync map[string]string
	var err error

	if groupName != "" {
		reposToSync, err = configMgr.GetGroupRepositories(groupName)
		if err != nil {
			return fmt.Errorf("failed to get group repositories: %w", err)
		}
		if len(reposToSync) == 0 {
			fmt.Printf("Group '%s' has no repositories.\n", groupName)
			return nil
		}
	} else {
		reposToSync = cfg.Repositories
	}

	// Determine sync mode
	mode := determineSyncMode(cfg)

	// Get Git manager for status checking
	gitMgr := di.GitManager()

	// Filter repositories based on conditions
	filteredRepos, err := filterRepositories(reposToSync, gitMgr)
	if err != nil {
		return fmt.Errorf("failed to filter repositories: %w", err)
	}

	if len(filteredRepos) == 0 {
		fmt.Println("No repositories match the specified criteria.")
		return nil
	}

	// Display what will be synced
	if dryRun {
		groupInfo := ""
		if groupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", groupName)
		}
		fmt.Printf("DRY RUN: Would synchronize %d repositories%s (mode: %s):\n\n", len(filteredRepos), groupInfo, mode)
		for alias, path := range filteredRepos {
			fmt.Printf("  %s → %s\n", alias, path)
		}
		return nil
	}

	// Setup progress tracking
	var progressBar *progress.MultiBar
	if showProgress {
		progressBar = progress.NewMultiBar()
		for alias := range filteredRepos {
			progressBar.AddOperation(alias)
		}
	} else {
		groupInfo := ""
		if groupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", groupName)
		}
		fmt.Printf("Synchronizing %d repositories%s (mode: %s)...\n\n", len(filteredRepos), groupInfo, mode)
	}

	// Use a channel to collect results
	resultChan := make(chan syncResult, len(filteredRepos))
	var wg sync.WaitGroup

	maxConcurrency := cfg.Settings.ParallelJobs
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	semaphore := make(chan struct{}, maxConcurrency)

	for alias, path := range filteredRepos {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if progressBar != nil {
				progressBar.StartOperation(alias)
			}

			err := gitMgr.SyncRepository(path, mode)

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
	var successCount, errorCount int
	var results []syncResult
	for result := range resultChan {
		results = append(results, result)
		if result.error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Finish progress display
	if progressBar != nil {
		progressBar.Finish()
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

// filterRepositories filters repositories based on conditional sync options
func filterRepositories(repos map[string]string, gitMgr *git.Manager) (map[string]string, error) {
	// If no filter options are specified, return all repositories
	if !onlyDirty && !onlyBehind && !onlyAhead {
		return repos, nil
	}

	filtered := make(map[string]string)

	for alias, path := range repos {
		// Get repository status
		status := gitMgr.GetRepoStatus(alias, path)
		if status.Error != nil {
			// Skip repositories with errors, but log them
			fmt.Printf("Warning: Skipping %s due to error: %v\n", alias, status.Error)
			continue
		}

		shouldInclude := false

		// Check conditions
		if onlyDirty && status.Workspace != types.Clean {
			shouldInclude = true
		}

		if onlyBehind && status.SyncStatus.Behind > 0 {
			shouldInclude = true
		}

		if onlyAhead && status.SyncStatus.Ahead > 0 {
			shouldInclude = true
		}

		if shouldInclude {
			filtered[alias] = path
		}
	}

	return filtered, nil
}
