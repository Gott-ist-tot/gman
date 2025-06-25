package batch

import (
	"fmt"
	"sort"
	"strings"

	"gman/internal/di"
	"gman/internal/git"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// StashSaveOperation implements BatchOperation for stash save
type StashSaveOperation struct {
	Message string
}

func (s *StashSaveOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	hasChanges, err := gitMgr.HasUncommittedChanges(path)
	if err != nil {
		return false, err
	}
	return hasChanges, nil
}

func (s *StashSaveOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	return gitMgr.StashSave(path, s.Message)
}

func (s *StashSaveOperation) GetOperationName() string {
	return "stash save"
}

// StashPopOperation implements BatchOperation for stash pop
type StashPopOperation struct{}

func (s *StashPopOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	// Check if repository has any stashes
	stashes, err := gitMgr.StashList(path)
	if err != nil {
		return false, err
	}
	return len(stashes) > 0, nil
}

func (s *StashPopOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	return gitMgr.StashPop(path)
}

func (s *StashPopOperation) GetOperationName() string {
	return "stash pop"
}

// StashClearOperation implements BatchOperation for stash clear
type StashClearOperation struct{}

func (s *StashClearOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	// Check if repository has any stashes
	stashes, err := gitMgr.StashList(path)
	if err != nil {
		return false, err
	}
	return len(stashes) > 0, nil
}

func (s *StashClearOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	return gitMgr.StashClear(path)
}

func (s *StashClearOperation) GetOperationName() string {
	return "stash clear"
}

// NewStashCmd creates the batch stash command with subcommands
func NewStashCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	// Add subcommands
	cmd.AddCommand(NewStashSaveCmd())
	cmd.AddCommand(NewStashPopCmd())
	cmd.AddCommand(NewStashListCmd())
	cmd.AddCommand(NewStashClearCmd())

	// Add common flags to parent command
	cmd.PersistentFlags().StringVar(&BatchGroupName, "group", "", "Operate only on repositories in the specified group")
	cmd.PersistentFlags().BoolVar(&BatchDryRun, "dry-run", false, "Show what would be done without executing")

	return cmd
}

// NewStashSaveCmd creates the stash save subcommand
func NewStashSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save [message]",
		Short: "Save current changes to stash across repositories",
		Long:  "Save uncommitted changes to stash in all repositories with dirty workspace.",
		RunE:  runStashSave,
	}

	return cmd
}

// NewStashPopCmd creates the stash pop subcommand
func NewStashPopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pop",
		Short: "Apply and remove the latest stash across repositories",
		Long:  "Apply the most recent stash and remove it from the stash list.",
		RunE:  runStashPop,
	}

	return cmd
}

// NewStashListCmd creates the stash list subcommand
func NewStashListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List stashes across all repositories",
		Long:    "Display all stashes for each repository.",
		Aliases: []string{"ls"},
		RunE:    runStashList,
	}

	return cmd
}

// NewStashClearCmd creates the stash clear subcommand
func NewStashClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all stashes across repositories",
		Long:  "Remove all stashes from all repositories.",
		RunE:  runStashClear,
	}

	return cmd
}

func runStashSave(cmd *cobra.Command, args []string) error {
	message := "gman batch stash"
	if len(args) > 0 {
		message = strings.Join(args, " ")
	}

	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute stash save operation
	operation := &StashSaveOperation{Message: message}
	return RunBatchOperation(operation, configMgr)
}

func runStashPop(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute stash pop operation
	operation := &StashPopOperation{}
	return RunBatchOperation(operation, configMgr)
}

func runStashList(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

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

	// Collect stash information
	type repoStashes struct {
		alias   string
		path    string
		stashes []string
		error   error
	}

	var results []repoStashes
	for alias, path := range reposToProcess {
		stashes, err := gitMgr.StashList(path)
		results = append(results, repoStashes{
			alias:   alias,
			path:    path,
			stashes: stashes,
			error:   err,
		})
	}

	// Sort results by alias
	sort.Slice(results, func(i, j int) bool {
		return results[i].alias < results[j].alias
	})

	// Display results
	hasStashes := false
	for _, result := range results {
		if result.error != nil {
			color.Red("❌ %s: Error - %v", result.alias, result.error)
			continue
		}

		if len(result.stashes) > 0 {
			hasStashes = true
			color.Green("📦 %s (%d stashes):", result.alias, len(result.stashes))
			for i, stash := range result.stashes {
				fmt.Printf("  %d: %s\n", i, stash)
			}
			fmt.Println()
		}
	}

	if !hasStashes {
		fmt.Println("No stashes found in any repository.")
	}

	return nil
}

func runStashClear(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute stash clear operation
	operation := &StashClearOperation{}
	return RunBatchOperation(operation, configMgr)
}
