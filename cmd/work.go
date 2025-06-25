package cmd

import (
	"github.com/spf13/cobra"
)

// workCmd represents the Git workflow command group
var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Git workflow commands",
	Long: `Git workflow commands for day-to-day development operations.

This command group includes:
- status checking across repositories
- synchronization with remotes  
- branch management
- commit and push operations
- stash management

Examples:
  gman work status                     # Check status of all repositories
  gman work sync                       # Sync all repositories
  gman work branch list               # List branches across repositories
  gman work commit -m "Fix bug"       # Commit across repositories`,
	Aliases: []string{"w"},
}

// workStatusCmd shows repository status (alias for existing status command)
var workStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all repositories",
	Long: `Show the current status of all configured repositories.

This includes:
- Working directory status (clean/dirty/stashed)
- Sync status with remote (ahead/behind/up-to-date)
- Current branch information
- Recent commit details

Examples:
  gman work status                 # Basic status
  gman work status --extended      # Detailed status with file counts
  gman work status --group webdev  # Status for specific group only`,
	RunE:    runStatus, // Reuse existing status command logic
	Aliases: []string{"st"},
}

// workSyncCmd synchronizes repositories (alias for existing sync command)
var workSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize all repositories with their remotes",
	Long: `Synchronize all repositories with their remote counterparts.

This performs git pull operations across all configured repositories,
respecting the configured sync mode (ff-only, rebase, autostash).

Examples:
  gman work sync                       # Sync all repositories
  gman work sync --only-behind         # Only sync repositories behind remote
  gman work sync --dry-run             # Preview what would be synced
  gman work sync --group webdev        # Sync specific group only`,
	RunE:    runSync, // Reuse existing sync command logic
	Aliases: []string{"s"},
}

// workCommitCmd commits changes (alias for existing commit command)
var workCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes across multiple repositories",
	Long: `Commit changes across multiple repositories with the same message.

This is useful for making consistent commits across related projects
or when applying the same fix to multiple repositories.

Examples:
  gman work commit -m "Update dependencies"     # Commit with message
  gman work commit -m "Fix bug" --add           # Add and commit
  gman work commit -m "Release v1.0" --group prod  # Commit specific group`,
	RunE:    runBatchCommit, // Use existing batch commit logic
	Aliases: []string{"c"},
}

// workPushCmd pushes changes (alias for existing push command)
var workPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes across multiple repositories",
	Long: `Push local commits to remote repositories.

Pushes committed changes from all repositories that have unpushed commits.
Useful for batch pushing after making commits across multiple projects.

Examples:
  gman work push                        # Push all repositories
  gman work push --force                # Force push (use carefully)
  gman work push --set-upstream         # Set upstream for new branches
  gman work push --group webdev         # Push specific group only`,
	RunE:    runBatchPush, // Use existing batch push logic
	Aliases: []string{"p"},
}

// workPullCmd pulls changes (alias for existing pull command)
var workPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes across multiple repositories", 
	Long: `Pull changes from remote repositories.

This is similar to sync but specifically performs git pull operations.
Useful when you want to explicitly pull without other sync behaviors.

Examples:
  gman work pull                        # Pull all repositories
  gman work pull --group webdev         # Pull specific group only`,
	RunE:    runBatchPull, // Use existing batch pull logic
	Aliases: []string{"pl"},
}

// workStashCmd manages stashes (alias for existing stash command)
var workStashCmd = &cobra.Command{
	Use:   "stash",
	Short: "Manage stashes across multiple repositories",
	Long: `Manage Git stashes across multiple repositories.

Allows you to save, apply, list, and clear stashes across all
your configured repositories simultaneously.

Examples:
  gman work stash save "WIP feature"    # Stash changes with message
  gman work stash pop                   # Apply and remove latest stash
  gman work stash list                  # List all stashes
  gman work stash clear                 # Clear all stashes`,
	RunE:    batchStashCmd.RunE, // Use existing batch stash logic
	Aliases: []string{"st"},
}

func init() {
	rootCmd.AddCommand(workCmd)
	
	// Add subcommands to work group
	workCmd.AddCommand(workStatusCmd)
	workCmd.AddCommand(workSyncCmd)
	workCmd.AddCommand(workCommitCmd)
	workCmd.AddCommand(workPushCmd)
	workCmd.AddCommand(workPullCmd)
	workCmd.AddCommand(workStashCmd)
	
	// Add branch and diff as subcommands (they're already defined)
	workCmd.AddCommand(branchCmd)
	workCmd.AddCommand(diffCmd)
	
	// Copy flags from original commands
	copyCommandFlags(workStatusCmd, statusCmd)
	copyCommandFlags(workSyncCmd, syncCmd)
	copyCommandFlags(workCommitCmd, batchCommitCmd)
	copyCommandFlags(workPushCmd, batchPushCmd)
	copyCommandFlags(workPullCmd, batchPullCmd)
	copyCommandFlags(workStashCmd, batchStashCmd)
}