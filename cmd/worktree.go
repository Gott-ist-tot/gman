package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"gman/internal/git"
)

var (
	worktreeBranch string
	worktreeForce  bool
)

// worktreeCmd represents the worktree command
var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage Git worktrees for repositories",
	Long: `Manage Git worktrees to enable parallel development on different branches
within the same repository. This is the recommended solution for scenarios
where you need multiple working directories for the same repository.

Git worktrees share the same .git directory but have separate working
directories, eliminating the need for multiple repository clones.

Examples:
  # Create a worktree for feature development
  gman worktree add myrepo ../myrepo-feature --branch feature-xyz
  
  # List all worktrees for a repository
  gman worktree list myrepo
  
  # Remove a worktree when development is complete
  gman worktree remove myrepo ../myrepo-feature`,
}

// worktreeAddCmd represents the worktree add command
var worktreeAddCmd = &cobra.Command{
	Use:   "add <repo> <path> --branch <branch>",
	Short: "Create a new worktree for a repository",
	Long: `Create a new Git worktree at the specified path for the given branch.
This allows you to work on multiple branches simultaneously without the
overhead of multiple repository clones.

The path can be relative or absolute. If the branch doesn't exist,
it will be created automatically.

Examples:
  gman worktree add myrepo ../myrepo-feature --branch feature-xyz
  gman worktree add backend /tmp/backend-hotfix --branch hotfix/urgent-fix`,
	Args: cobra.ExactArgs(2),
	RunE: runWorktreeAdd,
}

// worktreeListCmd represents the worktree list command
var worktreeListCmd = &cobra.Command{
	Use:   "list <repo>",
	Short: "List all worktrees for a repository",
	Long: `Display all Git worktrees associated with the specified repository.
Shows the worktree path, checked out branch, and HEAD commit.

This helps you understand all active working directories for a repository
and their current state.

Examples:
  gman worktree list myrepo
  gman worktree list backend`,
	Args: cobra.ExactArgs(1),
	RunE: runWorktreeList,
}

// worktreeRemoveCmd represents the worktree remove command
var worktreeRemoveCmd = &cobra.Command{
	Use:   "remove <repo> <worktree_path>",
	Short: "Remove a worktree from a repository",
	Long: `Remove a Git worktree at the specified path. This will clean up
the worktree directory and remove it from Git's worktree list.

Use --force to remove a worktree even if it has uncommitted changes.

Examples:
  gman worktree remove myrepo ../myrepo-feature
  gman worktree remove backend /tmp/backend-hotfix --force`,
	Args: cobra.ExactArgs(2),
	RunE: runWorktreeRemove,
}

func init() {
	rootCmd.AddCommand(worktreeCmd)
	worktreeCmd.AddCommand(worktreeAddCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)

	// Add flags
	worktreeAddCmd.Flags().StringVarP(&worktreeBranch, "branch", "b", "", "Branch to checkout in the new worktree (required)")
	worktreeAddCmd.MarkFlagRequired("branch")
	
	worktreeRemoveCmd.Flags().BoolVarP(&worktreeForce, "force", "f", false, "Force removal even with uncommitted changes")
}

func runWorktreeAdd(cmd *cobra.Command, args []string) error {
	repoAlias := args[0]
	worktreePath := args[1]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	repoPath, exists := cfg.Repositories[repoAlias]
	if !exists {
		return fmt.Errorf("repository '%s' not found. Use 'gman list' to see available repositories", repoAlias)
	}

	// Convert to absolute path
	absWorktreePath, err := filepath.Abs(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create git manager and add worktree
	gitMgr := git.NewManager()
	if err := gitMgr.AddWorktree(repoPath, absWorktreePath, worktreeBranch); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	fmt.Printf("✅ Created worktree for branch '%s' at: %s\n", worktreeBranch, absWorktreePath)
	fmt.Printf("💡 You can now switch to it using: gman switch %s\n", filepath.Base(absWorktreePath))
	
	return nil
}

func runWorktreeList(cmd *cobra.Command, args []string) error {
	repoAlias := args[0]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	repoPath, exists := cfg.Repositories[repoAlias]
	if !exists {
		return fmt.Errorf("repository '%s' not found. Use 'gman list' to see available repositories", repoAlias)
	}

	// Create git manager and list worktrees
	gitMgr := git.NewManager()
	worktrees, err := gitMgr.ListWorktrees(repoPath)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Printf("No worktrees found for repository '%s'\n", repoAlias)
		fmt.Printf("Create one with: gman worktree add %s <path> --branch <branch>\n", repoAlias)
		return nil
	}

	fmt.Printf("Worktrees for repository '%s':\n\n", repoAlias)
	for i, wt := range worktrees {
		status := "✅"
		if wt.IsBare {
			status = "📦"
		} else if wt.IsDetached {
			status = "🔄"
		}
		
		fmt.Printf("%d. %s %s\n", i+1, status, wt.Path)
		fmt.Printf("   Branch: %s\n", wt.Branch)
		if wt.Commit != "" {
			fmt.Printf("   Commit: %s\n", wt.Commit[:8])
		}
		if i < len(worktrees)-1 {
			fmt.Println()
		}
	}

	return nil
}

func runWorktreeRemove(cmd *cobra.Command, args []string) error {
	repoAlias := args[0]
	worktreePath := args[1]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	repoPath, exists := cfg.Repositories[repoAlias]
	if !exists {
		return fmt.Errorf("repository '%s' not found. Use 'gman list' to see available repositories", repoAlias)
	}

	// Convert to absolute path for comparison
	absWorktreePath, err := filepath.Abs(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create git manager and remove worktree
	gitMgr := git.NewManager()
	if err := gitMgr.RemoveWorktree(repoPath, absWorktreePath, worktreeForce); err != nil {
		if strings.Contains(err.Error(), "uncommitted changes") && !worktreeForce {
			return fmt.Errorf("worktree has uncommitted changes. Use --force to remove anyway: %w", err)
		}
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	fmt.Printf("🗑️  Removed worktree: %s\n", absWorktreePath)
	return nil
}