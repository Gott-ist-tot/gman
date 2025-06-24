package cmd

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/git"
	"gman/internal/progress"
	"github.com/fatih/color"
)

// batchCommitCmd commits changes across multiple repositories
var batchCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes across multiple repositories",
	Long: `Commit changes with the same message across multiple repositories.
Only commits repositories that have staged or unstaged changes.

Examples:
  gman commit -m "Fix critical bug"
  gman commit -m "Update dependencies" --group backend
  gman commit -m "Feature: new dashboard" --add`,
	RunE: runBatchCommit,
}

// batchPushCmd pushes changes across multiple repositories
var batchPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes across multiple repositories",
	Long: `Push local commits to remote repositories.
Only pushes repositories that have unpushed commits.

Examples:
  gman push
  gman push --group frontend
  gman push --dry-run`,
	RunE: runBatchPush,
}

// batchPullCmd pulls changes across multiple repositories
var batchPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes across multiple repositories",
	Long: `Pull changes from remote repositories.
Uses the same sync logic as the sync command but with pull-specific options.

Examples:
  gman pull
  gman pull --group backend
  gman pull --rebase`,
	RunE: runBatchPull,
}

// batchStashCmd manages stashes across multiple repositories
var batchStashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Manage stashes across multiple repositories",
	Long: `Manage stashes across all configured repositories.
Supports save, pop, list, and clear operations.

Examples:
  gman stash save "Work in progress"
  gman stash pop
  gman stash list
  gman stash clear`,
}

// Stash subcommands
var stashSaveCmd = &cobra.Command{
	Use:   "save [message]",
	Short: "Save current changes to stash across repositories",
	Long:  "Save uncommitted changes to stash in all repositories with dirty workspace.",
	RunE:  runStashSave,
}

var stashPopCmd = &cobra.Command{
	Use:   "pop",
	Short: "Apply and remove the latest stash across repositories",
	Long:  "Apply the most recent stash and remove it from the stash list.",
	RunE:  runStashPop,
}

var stashListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stashes across all repositories",
	Long:  "Display all stashes for each repository.",
	Aliases: []string{"ls"},
	RunE:  runStashList,
}

var stashClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all stashes across repositories",
	Long:  "Remove all stashes from all repositories.",
	RunE:  runStashClear,
}

var (
	commitMessage    string
	commitAddAll     bool
	pushForce        bool
	pushSetUpstream  bool
	pullRebase       bool
	pullStrategy     string
	batchGroupName   string
	batchDryRun      bool
	batchProgress    bool
)

func init() {
	rootCmd.AddCommand(batchCommitCmd)
	rootCmd.AddCommand(batchPushCmd)
	rootCmd.AddCommand(batchPullCmd)
	rootCmd.AddCommand(batchStashCmd)
	
	// Add stash subcommands
	batchStashCmd.AddCommand(stashSaveCmd)
	batchStashCmd.AddCommand(stashPopCmd)
	batchStashCmd.AddCommand(stashListCmd)
	batchStashCmd.AddCommand(stashClearCmd)
	
	// Common flags
	batchCommitCmd.PersistentFlags().StringVar(&batchGroupName, "group", "", "Operate only on repositories in the specified group")
	batchPushCmd.PersistentFlags().StringVar(&batchGroupName, "group", "", "Operate only on repositories in the specified group")
	batchPullCmd.PersistentFlags().StringVar(&batchGroupName, "group", "", "Operate only on repositories in the specified group")
	batchStashCmd.PersistentFlags().StringVar(&batchGroupName, "group", "", "Operate only on repositories in the specified group")
	
	batchCommitCmd.PersistentFlags().BoolVar(&batchDryRun, "dry-run", false, "Show what would be done without executing")
	batchPushCmd.PersistentFlags().BoolVar(&batchDryRun, "dry-run", false, "Show what would be done without executing")
	batchPullCmd.PersistentFlags().BoolVar(&batchDryRun, "dry-run", false, "Show what would be done without executing")
	batchStashCmd.PersistentFlags().BoolVar(&batchDryRun, "dry-run", false, "Show what would be done without executing")
	
	batchCommitCmd.PersistentFlags().BoolVar(&batchProgress, "progress", false, "Show detailed progress during operations")
	batchPushCmd.PersistentFlags().BoolVar(&batchProgress, "progress", false, "Show detailed progress during operations")
	batchPullCmd.PersistentFlags().BoolVar(&batchProgress, "progress", false, "Show detailed progress during operations")
	
	// Commit-specific flags
	batchCommitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (required)")
	batchCommitCmd.Flags().BoolVarP(&commitAddAll, "add", "a", false, "Add all changes before committing")
	batchCommitCmd.MarkFlagRequired("message")
	
	// Push-specific flags
	batchPushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Force push (use with caution)")
	batchPushCmd.Flags().BoolVarP(&pushSetUpstream, "set-upstream", "u", false, "Set upstream for new branches")
	
	// Pull-specific flags
	batchPullCmd.Flags().BoolVar(&pullRebase, "rebase", false, "Use rebase instead of merge")
	batchPullCmd.Flags().StringVar(&pullStrategy, "strategy", "", "Merge strategy (ours, theirs, etc.)")
}

func runBatchCommit(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	gitMgr := git.NewManager()
	
	// Filter repositories with changes
	reposWithChanges := make(map[string]string)
	for alias, path := range reposToProcess {
		hasChanges, err := gitMgr.HasUncommittedChanges(path)
		if err != nil {
			fmt.Printf("Warning: Failed to check changes in %s: %v\n", alias, err)
			continue
		}
		if hasChanges {
			reposWithChanges[alias] = path
		}
	}

	if len(reposWithChanges) == 0 {
		fmt.Println("No repositories have uncommitted changes.")
		return nil
	}

	if batchDryRun {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("DRY RUN: Would commit changes in %d repositories%s:\n", len(reposWithChanges), groupInfo)
		fmt.Printf("Message: %s\n\n", commitMessage)
		for alias := range reposWithChanges {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	// Setup progress tracking
	var progressBar *progress.MultiBar
	if batchProgress {
		progressBar = progress.NewMultiBar()
		for alias := range reposWithChanges {
			progressBar.AddOperation(alias)
		}
	} else {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("Committing changes in %d repositories%s...\n\n", len(reposWithChanges), groupInfo)
	}

	// Commit changes
	type commitResult struct {
		alias string
		error error
	}

	resultChan := make(chan commitResult, len(reposWithChanges))
	var wg sync.WaitGroup

	for alias, path := range reposWithChanges {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			
			if progressBar != nil {
				progressBar.StartOperation(alias)
			}
			
			err := gitMgr.CommitChanges(path, commitMessage, commitAddAll)
			
			if progressBar != nil {
				progressBar.CompleteOperation(alias, err)
			}
			
			resultChan <- commitResult{alias: alias, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.error != nil {
			if !batchProgress {
				fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			}
			errorCount++
		} else {
			if !batchProgress {
				fmt.Printf("✅ %s: committed successfully\n", result.alias)
			}
			successCount++
		}
	}

	if progressBar != nil {
		progressBar.Finish()
	}

	fmt.Printf("\nCommit completed: %d successful, %d failed\n", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("commit failed for %d repositories", errorCount)
	}

	return nil
}

func runBatchPush(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	gitMgr := git.NewManager()
	
	// Filter repositories with unpushed commits
	reposToPush := make(map[string]string)
	for alias, path := range reposToProcess {
		hasUnpushed, err := gitMgr.HasUnpushedCommits(path)
		if err != nil {
			fmt.Printf("Warning: Failed to check unpushed commits in %s: %v\n", alias, err)
			continue
		}
		if hasUnpushed {
			reposToPush[alias] = path
		}
	}

	if len(reposToPush) == 0 {
		fmt.Println("No repositories have unpushed commits.")
		return nil
	}

	if batchDryRun {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("DRY RUN: Would push changes in %d repositories%s:\n\n", len(reposToPush), groupInfo)
		for alias := range reposToPush {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	// Setup progress tracking
	var progressBar *progress.MultiBar
	if batchProgress {
		progressBar = progress.NewMultiBar()
		for alias := range reposToPush {
			progressBar.AddOperation(alias)
		}
	} else {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("Pushing changes in %d repositories%s...\n\n", len(reposToPush), groupInfo)
	}

	// Push changes
	type pushResult struct {
		alias string
		error error
	}

	resultChan := make(chan pushResult, len(reposToPush))
	var wg sync.WaitGroup

	for alias, path := range reposToPush {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			
			if progressBar != nil {
				progressBar.StartOperation(alias)
			}
			
			err := gitMgr.PushChanges(path, pushForce, pushSetUpstream)
			
			if progressBar != nil {
				progressBar.CompleteOperation(alias, err)
			}
			
			resultChan <- pushResult{alias: alias, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.error != nil {
			if !batchProgress {
				fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			}
			errorCount++
		} else {
			if !batchProgress {
				fmt.Printf("✅ %s: pushed successfully\n", result.alias)
			}
			successCount++
		}
	}

	if progressBar != nil {
		progressBar.Finish()
	}

	fmt.Printf("\nPush completed: %d successful, %d failed\n", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("push failed for %d repositories", errorCount)
	}

	return nil
}

func runBatchPull(cmd *cobra.Command, args []string) error {
	// Use existing sync functionality with pull-specific options
	syncMode := "ff-only"
	if pullRebase {
		syncMode = "rebase"
	}
	
	// This reuses the sync logic from sync.go
	// We could refactor to share the implementation
	fmt.Printf("Pull operation using sync mode: %s\n", syncMode)
	
	// For now, delegate to sync command logic
	// In a full implementation, we'd extract the sync logic to a shared function
	return fmt.Errorf("pull command not fully implemented - use 'gman sync' instead")
}

func runStashSave(cmd *cobra.Command, args []string) error {
	message := "gman batch stash"
	if len(args) > 0 {
		message = strings.Join(args, " ")
	}
	
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if batchDryRun {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("DRY RUN: Would stash changes in %d repositories%s:\n", len(reposToProcess), groupInfo)
		fmt.Printf("Message: %s\n\n", message)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := git.NewManager()
	
	// Stash changes
	fmt.Printf("Stashing changes in %d repositories...\n\n", len(reposToProcess))
	
	type stashResult struct {
		alias string
		error error
	}

	resultChan := make(chan stashResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			err := gitMgr.StashSave(path, message)
			resultChan <- stashResult{alias: alias, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.error != nil {
			fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			errorCount++
		} else {
			fmt.Printf("✅ %s: stashed successfully\n", result.alias)
			successCount++
		}
	}

	fmt.Printf("\nStash save completed: %d successful, %d failed\n", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("stash save failed for %d repositories", errorCount)
	}

	return nil
}

func runStashPop(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if batchDryRun {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("DRY RUN: Would pop stash in %d repositories%s:\n\n", len(reposToProcess), groupInfo)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := git.NewManager()
	
	// Pop stashes
	fmt.Printf("Popping stash in %d repositories...\n\n", len(reposToProcess))
	
	type stashResult struct {
		alias string
		error error
	}

	resultChan := make(chan stashResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			err := gitMgr.StashPop(path)
			resultChan <- stashResult{alias: alias, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.error != nil {
			fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			errorCount++
		} else {
			fmt.Printf("✅ %s: stash popped successfully\n", result.alias)
			successCount++
		}
	}

	fmt.Printf("\nStash pop completed: %d successful, %d failed\n", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("stash pop failed for %d repositories", errorCount)
	}

	return nil
}

func runStashList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	gitMgr := git.NewManager()
	
	// Collect stash information
	type stashInfo struct {
		alias   string
		stashes []string
		error   error
	}

	resultChan := make(chan stashInfo, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			stashes, err := gitMgr.StashList(path)
			resultChan <- stashInfo{alias: alias, stashes: stashes, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	fmt.Printf("\n%s:\n\n", color.CyanString("Stash Information"))

	var results []stashInfo
	for result := range resultChan {
		results = append(results, result)
	}

	// Sort by alias
	sort.Slice(results, func(i, j int) bool {
		return results[i].alias < results[j].alias
	})

	for _, result := range results {
		if result.error != nil {
			fmt.Printf("%s %s: %v\n", 
				color.RedString("❌"), 
				color.YellowString(result.alias), 
				result.error)
			continue
		}

		fmt.Printf("%s %s\n", 
			color.GreenString("📁"), 
			color.YellowString(result.alias))
		
		if len(result.stashes) == 0 {
			fmt.Printf("   No stashes\n")
		} else {
			for i, stash := range result.stashes {
				fmt.Printf("   stash@{%d}: %s\n", i, stash)
			}
		}
		
		fmt.Println()
	}

	return nil
}

func runStashClear(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getBatchRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if batchDryRun {
		groupInfo := ""
		if batchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", batchGroupName)
		}
		fmt.Printf("DRY RUN: Would clear all stashes in %d repositories%s:\n\n", len(reposToProcess), groupInfo)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := git.NewManager()
	
	// Clear stashes
	fmt.Printf("Clearing all stashes in %d repositories...\n\n", len(reposToProcess))
	
	type stashResult struct {
		alias string
		error error
	}

	resultChan := make(chan stashResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			err := gitMgr.StashClear(path)
			resultChan <- stashResult{alias: alias, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	
	for result := range resultChan {
		if result.error != nil {
			fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			errorCount++
		} else {
			fmt.Printf("✅ %s: stashes cleared successfully\n", result.alias)
			successCount++
		}
	}

	fmt.Printf("\nStash clear completed: %d successful, %d failed\n", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("stash clear failed for %d repositories", errorCount)
	}

	return nil
}

// getBatchRepositoriesToProcess returns the repositories to operate on based on group filter
func getBatchRepositoriesToProcess(configMgr *config.Manager) (map[string]string, error) {
	cfg := configMgr.GetConfig()
	
	if batchGroupName != "" {
		return configMgr.GetGroupRepositories(batchGroupName)
	}
	
	return cfg.Repositories, nil
}