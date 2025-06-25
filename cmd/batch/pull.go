package batch

import (
	"fmt"

	"gman/internal/di"
	"gman/internal/git"

	"github.com/spf13/cobra"
)

var (
	pullRebase   bool
	pullStrategy string
)

// PullOperation implements BatchOperation for pull operations
type PullOperation struct{}

func (p *PullOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	// For pull operations, we typically want to include all repositories
	// unless they have specific conditions that would prevent pulling
	return true, nil
}

func (p *PullOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	// Determine pull mode based on flags
	mode := "ff-only" // default
	if pullRebase {
		mode = "rebase"
	}

	// Use the same sync logic as the sync command
	return gitMgr.SyncRepository(path, mode)
}

func (p *PullOperation) GetOperationName() string {
	return "pull"
}

// NewPullCmd creates the batch pull command
func NewPullCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	// Add common flags
	cmd.PersistentFlags().StringVar(&BatchGroupName, "group", "", "Operate only on repositories in the specified group")
	cmd.PersistentFlags().BoolVar(&BatchDryRun, "dry-run", false, "Show what would be done without executing")
	cmd.PersistentFlags().BoolVar(&BatchProgress, "progress", false, "Show detailed progress during operations")

	// Pull-specific flags
	cmd.Flags().BoolVar(&pullRebase, "rebase", false, "Use rebase instead of merge")
	cmd.Flags().StringVar(&pullStrategy, "strategy", "", "Merge strategy (ours, theirs, etc.)")

	return cmd
}

func runBatchPull(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute pull operation
	operation := &PullOperation{}
	return RunBatchOperation(operation, configMgr)
}
