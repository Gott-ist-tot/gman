package batch

import (
	"fmt"

	"gman/internal/di"
	"gman/internal/git"

	"github.com/spf13/cobra"
)

var (
	commitMessage string
	commitAddAll  bool
)

// CommitOperation implements BatchOperation for commit operations
type CommitOperation struct{}

func (c *CommitOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	hasChanges, err := gitMgr.HasUncommittedChanges(path)
	if err != nil {
		return false, err
	}
	return hasChanges, nil
}

func (c *CommitOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	return gitMgr.CommitChanges(path, commitMessage, commitAddAll)
}

func (c *CommitOperation) GetOperationName() string {
	return "commit"
}

// NewCommitCmd creates the batch commit command
func NewCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	// Add common flags
	cmd.PersistentFlags().StringVar(&BatchGroupName, "group", "", "Operate only on repositories in the specified group")
	cmd.PersistentFlags().BoolVar(&BatchDryRun, "dry-run", false, "Show what would be done without executing")
	cmd.PersistentFlags().BoolVar(&BatchProgress, "progress", false, "Show detailed progress during operations")

	// Commit-specific flags
	cmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (required)")
	cmd.Flags().BoolVarP(&commitAddAll, "add", "a", false, "Add all changes before committing")
	cmd.MarkFlagRequired("message")

	return cmd
}

func runBatchCommit(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute commit operation
	operation := &CommitOperation{}
	return RunBatchOperation(operation, configMgr)
}
