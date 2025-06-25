package cmd

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"gman/internal/config"
	"gman/internal/di"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// branchCmd represents the branch command
var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Cross-repository branch management",
	Long: `Manage branches across multiple repositories.
Provides unified branch operations for all configured repositories.

Examples:
  gman branch list                    # Show all branches across repositories
  gman branch create feature/new-ui  # Create branch in all repositories
  gman branch switch main            # Switch to main branch in all repositories
  gman branch clean                  # Clean merged branches in all repositories`,
}

// branchListCmd lists branches across repositories
var branchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List branches across all repositories",
	Long: `Display branch information for all configured repositories.
Shows current branch, available branches, and branch status.`,
	Aliases: []string{"ls"},
	RunE:    runBranchList,
}

// branchCreateCmd creates a branch in multiple repositories
var branchCreateCmd = &cobra.Command{
	Use:   "create <branch-name>",
	Short: "Create a branch in multiple repositories",
	Long: `Create a new branch with the same name across selected repositories.
The branch will be created from the current HEAD of each repository.`,
	Args: cobra.ExactArgs(1),
	RunE: runBranchCreate,
}

// branchSwitchCmd switches branches across repositories
var branchSwitchCmd = &cobra.Command{
	Use:   "switch <branch-name>",
	Short: "Switch to a branch across multiple repositories",
	Long: `Switch to the specified branch in all repositories where it exists.
Repositories without the branch will be skipped with a warning.`,
	Aliases: []string{"checkout", "co"},
	Args:    cobra.ExactArgs(1),
	RunE:    runBranchSwitch,
}

// branchCleanCmd cleans merged branches
var branchCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean merged branches across repositories",
	Long: `Remove local branches that have been merged into the main branch.
This helps keep the branch list clean by removing outdated feature branches.`,
	RunE: runBranchClean,
}

var (
	branchGroupName  string
	branchDryRun     bool
	branchShowRemote bool
	branchVerbose    bool
	branchMainBranch string
)

func init() {
	// Command is now available via: gman work branch
	// Removed direct rootCmd registration to avoid duplication

	// Add subcommands
	branchCmd.AddCommand(branchListCmd)
	branchCmd.AddCommand(branchCreateCmd)
	branchCmd.AddCommand(branchSwitchCmd)
	branchCmd.AddCommand(branchCleanCmd)

	// Add flags for branch operations
	branchCmd.PersistentFlags().StringVar(&branchGroupName, "group", "", "Operate only on repositories in the specified group")
	branchCmd.PersistentFlags().BoolVar(&branchDryRun, "dry-run", false, "Show what would be done without executing")

	// Flags for branch list
	branchListCmd.Flags().BoolVar(&branchShowRemote, "remote", false, "Show remote branches")
	branchListCmd.Flags().BoolVarP(&branchVerbose, "verbose", "v", false, "Show detailed branch information")

	// Flags for branch clean
	branchCleanCmd.Flags().StringVar(&branchMainBranch, "main", "", "Specify main branch (default: auto-detect)")
}

func runBranchList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	gitMgr := di.GitManager()

	// Collect branch information
	type branchInfo struct {
		alias         string
		currentBranch string
		branches      []string
		error         error
	}

	resultChan := make(chan branchInfo, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()

			info := branchInfo{alias: alias}

			// Get current branch
			if current, err := gitMgr.GetCurrentBranch(path); err != nil {
				info.error = err
			} else {
				info.currentBranch = current
			}

			// Get all branches
			if branches, err := gitMgr.GetBranches(path, branchShowRemote); err != nil {
				if info.error == nil {
					info.error = err
				}
			} else {
				info.branches = branches
			}

			resultChan <- info
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	fmt.Printf("\n%s:\n\n", color.CyanString("Branch Information"))

	var results []branchInfo
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

		if result.currentBranch != "" {
			fmt.Printf("   Current: %s\n",
				color.GreenString("* "+result.currentBranch))
		}

		if branchVerbose && len(result.branches) > 0 {
			fmt.Printf("   Branches: ")
			var branchStrs []string
			for _, branch := range result.branches {
				if branch == result.currentBranch {
					branchStrs = append(branchStrs, color.GreenString("*"+branch))
				} else {
					branchStrs = append(branchStrs, branch)
				}
			}
			fmt.Printf("%s\n", strings.Join(branchStrs, ", "))
		}

		fmt.Println()
	}

	return nil
}

func runBranchCreate(cmd *cobra.Command, args []string) error {
	branchName := args[0]

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if branchDryRun {
		groupInfo := ""
		if branchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", branchGroupName)
		}
		fmt.Printf("DRY RUN: Would create branch '%s' in %d repositories%s:\n\n",
			branchName, len(reposToProcess), groupInfo)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := di.GitManager()

	// Create branches
	fmt.Printf("Creating branch '%s' in %d repositories...\n\n", branchName, len(reposToProcess))

	type createResult struct {
		alias string
		error error
	}

	resultChan := make(chan createResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			err := gitMgr.CreateBranch(path, branchName)
			resultChan <- createResult{alias: alias, error: err}
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
			fmt.Printf("✅ %s: branch created\n", result.alias)
			successCount++
		}
	}

	fmt.Printf("\nBranch creation completed: %d successful, %d failed\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("branch creation failed for %d repositories", errorCount)
	}

	return nil
}

func runBranchSwitch(cmd *cobra.Command, args []string) error {
	branchName := args[0]

	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if branchDryRun {
		groupInfo := ""
		if branchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", branchGroupName)
		}
		fmt.Printf("DRY RUN: Would switch to branch '%s' in %d repositories%s:\n\n",
			branchName, len(reposToProcess), groupInfo)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := di.GitManager()

	// Switch branches
	fmt.Printf("Switching to branch '%s' in %d repositories...\n\n", branchName, len(reposToProcess))

	type switchResult struct {
		alias string
		error error
	}

	resultChan := make(chan switchResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			err := gitMgr.SwitchBranch(path, branchName)
			resultChan <- switchResult{alias: alias, error: err}
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
			fmt.Printf("✅ %s: switched to %s\n", result.alias, branchName)
			successCount++
		}
	}

	fmt.Printf("\nBranch switch completed: %d successful, %d failed\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("branch switch failed for %d repositories", errorCount)
	}

	return nil
}

func runBranchClean(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get repositories to operate on
	reposToProcess, err := getRepositoriesToProcess(configMgr)
	if err != nil {
		return err
	}

	if len(reposToProcess) == 0 {
		fmt.Println("No repositories to process.")
		return nil
	}

	if branchDryRun {
		groupInfo := ""
		if branchGroupName != "" {
			groupInfo = fmt.Sprintf(" from group '%s'", branchGroupName)
		}
		fmt.Printf("DRY RUN: Would clean merged branches in %d repositories%s:\n\n",
			len(reposToProcess), groupInfo)
		for alias := range reposToProcess {
			fmt.Printf("  %s\n", alias)
		}
		return nil
	}

	gitMgr := di.GitManager()

	// Clean merged branches
	fmt.Printf("Cleaning merged branches in %d repositories...\n\n", len(reposToProcess))

	type cleanResult struct {
		alias   string
		cleaned []string
		error   error
	}

	resultChan := make(chan cleanResult, len(reposToProcess))
	var wg sync.WaitGroup

	for alias, path := range reposToProcess {
		wg.Add(1)
		go func(alias, path string) {
			defer wg.Done()
			cleaned, err := gitMgr.CleanMergedBranches(path, branchMainBranch)
			resultChan <- cleanResult{alias: alias, cleaned: cleaned, error: err}
		}(alias, path)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Display results
	successCount := 0
	errorCount := 0
	totalCleaned := 0

	for result := range resultChan {
		if result.error != nil {
			fmt.Printf("❌ %s: %v\n", result.alias, result.error)
			errorCount++
		} else {
			if len(result.cleaned) > 0 {
				fmt.Printf("✅ %s: cleaned %d branches (%s)\n",
					result.alias, len(result.cleaned), strings.Join(result.cleaned, ", "))
				totalCleaned += len(result.cleaned)
			} else {
				fmt.Printf("✅ %s: no branches to clean\n", result.alias)
			}
			successCount++
		}
	}

	fmt.Printf("\nBranch cleanup completed: %d successful, %d failed, %d branches cleaned\n",
		successCount, errorCount, totalCleaned)

	if errorCount > 0 {
		return fmt.Errorf("branch cleanup failed for %d repositories", errorCount)
	}

	return nil
}

// getRepositoriesToProcess returns the repositories to operate on based on group filter
func getRepositoriesToProcess(configMgr *config.Manager) (map[string]string, error) {
	cfg := configMgr.GetConfig()

	if branchGroupName != "" {
		return configMgr.GetGroupRepositories(branchGroupName)
	}

	return cfg.Repositories, nil
}
