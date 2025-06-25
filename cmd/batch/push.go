package batch

import (
	"fmt"

	"gman/internal/di"
	"gman/internal/git"

	"github.com/spf13/cobra"
)

var (
	pushForce       bool
	pushSetUpstream bool
)

// PushOperation implements BatchOperation for push operations
type PushOperation struct{}

func (p *PushOperation) ShouldInclude(gitMgr *git.Manager, alias, path string) (bool, error) {
	hasUnpushed, err := gitMgr.HasUnpushedCommits(path)
	if err != nil {
		return false, err
	}
	return hasUnpushed, nil
}

func (p *PushOperation) Execute(gitMgr *git.Manager, alias, path string) error {
	return gitMgr.PushChanges(path, pushForce, pushSetUpstream)
}

func (p *PushOperation) GetOperationName() string {
	return "push"
}

// NewPushCmd creates the batch push command
func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	// Add common flags
	cmd.PersistentFlags().StringVar(&BatchGroupName, "group", "", "Operate only on repositories in the specified group")
	cmd.PersistentFlags().BoolVar(&BatchDryRun, "dry-run", false, "Show what would be done without executing")
	cmd.PersistentFlags().BoolVar(&BatchProgress, "progress", false, "Show detailed progress during operations")

	// Push-specific flags
	cmd.Flags().BoolVarP(&pushForce, "force", "f", false, "Force push (use with caution)")
	cmd.Flags().BoolVarP(&pushSetUpstream, "set-upstream", "u", false, "Set upstream for new branches")

	return cmd
}

func runBatchPush(cmd *cobra.Command, args []string) error {
	// Get configuration from DI container
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and execute push operation
	operation := &PushOperation{}
	return RunBatchOperation(operation, configMgr)
}
