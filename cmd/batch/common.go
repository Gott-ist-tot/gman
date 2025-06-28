package batch

import (
	"fmt"
	"sort"
	"sync"

	"gman/internal/config"
	"gman/internal/di"
	"gman/internal/git"
	"gman/internal/progress"
)

// Shared variables for batch operations
var (
	BatchGroupName string
	BatchDryRun    bool
	BatchProgress  bool
)

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Alias string
	Path  string
	Error error
}

// BatchOperation defines the interface for batch operations
type BatchOperation interface {
	ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error)
	Execute(gitMgr *git.Manager, alias, path string) error
	GetOperationName() string
}

// RunBatchOperation executes a batch operation across multiple repositories
func RunBatchOperation(operation BatchOperation, configMgr *config.Manager) error {
	// Get repositories to operate on
	reposToProcess, err := GetBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	gitMgr := di.GitManager()

	// Filter repositories that should be included
	filteredRepos := make(map[string]string)
	for alias, path := range reposToProcess {
		shouldInclude, err := operation.ShouldInclude(gitMgr, alias, path)
		if err != nil {
			fmt.Printf("Warning: Failed to check %s: %v\n", alias, err)
			continue
		}
		if shouldInclude {
			filteredRepos[alias] = path
		}
	}

	if len(filteredRepos) == 0 {
		fmt.Printf("No repositories qualify for %s operation.\n", operation.GetOperationName())
		return nil
	}

	// Handle dry-run mode
	if BatchDryRun {
		groupInfo := ""
		if BatchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", BatchGroupName)
		}
		fmt.Printf("DRY RUN: Would %s %d repositories%s:\n", operation.GetOperationName(), len(filteredRepos), groupInfo)
		for alias, path := range filteredRepos {
			fmt.Printf("  %s → %s\n", alias, path)
		}
		return nil
	}

	// Setup progress tracking
	var progressBar *progress.MultiBar
	if BatchProgress {
		progressBar = progress.NewMultiBar()
		for alias := range filteredRepos {
			progressBar.AddOperation(alias)
		}
	} else {
		groupInfo := ""
		if BatchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", BatchGroupName)
		}
		fmt.Printf("Running %s on %d repositories%s...\n\n", operation.GetOperationName(), len(filteredRepos), groupInfo)
	}

	// Execute operation concurrently
	resultChan := make(chan BatchResult, len(filteredRepos))
	var wg sync.WaitGroup

	cfg := configMgr.GetConfig()
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

			err := operation.Execute(gitMgr, alias, path)

			if progressBar != nil {
				progressBar.CompleteOperation(alias, err)
			}

			resultChan <- BatchResult{
				Alias: alias,
				Path:  path,
				Error: err,
			}
		}(alias, path)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var successCount, errorCount int
	var results []BatchResult
	for result := range resultChan {
		results = append(results, result)
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// Finish progress display
	if progressBar != nil {
		progressBar.Finish()
	}

	// Sort results by alias for consistent output order
	sort.Slice(results, func(i, j int) bool {
		return results[i].Alias < results[j].Alias
	})

	// Display detailed results if not using progress mode
	if !BatchProgress {
		for _, result := range results {
			if result.Error != nil {
				fmt.Printf("❌ %s: %v\n", result.Alias, result.Error)
			} else {
				fmt.Printf("✅ %s: %s completed successfully\n", result.Alias, operation.GetOperationName())
			}
		}
	}

	// Summary
	fmt.Printf("\n%s completed: %d successful, %d failed\n", operation.GetOperationName(), successCount, errorCount)

	// Show failed repositories if any
	if errorCount > 0 && BatchProgress {
		fmt.Println("\nFailed repositories:")
		for _, result := range results {
			if result.Error != nil {
				fmt.Printf("  ❌ %s: %v\n", result.Alias, result.Error)
			}
		}
		return fmt.Errorf("%s failed for %d repositories", operation.GetOperationName(), errorCount)
	} else if errorCount > 0 {
		return fmt.Errorf("%s failed for %d repositories", operation.GetOperationName(), errorCount)
	}

	return nil
}

// GetBatchRepositoriesToProcess returns the repositories to operate on based on group filter
func GetBatchRepositoriesToProcess(configMgr *config.Manager) (map[string]string, error) {
	cfg := configMgr.GetConfig()

	if BatchGroupName != "" {
		return configMgr.GetGroupRepositories(BatchGroupName)
	}

	return cfg.Repositories, nil
}

// AddCommonFlags adds common flags to a command
func AddCommonFlags(cmd interface{}) {
	// This will be implemented to add common flags
	// to avoid duplication across command files
}
