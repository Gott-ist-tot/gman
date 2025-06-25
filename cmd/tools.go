package cmd

import (
	"github.com/spf13/cobra"
)

// toolsCmd represents the tools and utilities command group
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Tools and utilities",
	Long: `Advanced tools and utilities for power users.

This command group includes:
- search and indexing capabilities
- worktree management  
- interactive dashboard
- setup and configuration tools
- shell integration utilities

Examples:
  gman tools find "config.yaml"       # Search for files across repositories
  gman tools index rebuild            # Rebuild search index
  gman tools dashboard                 # Launch interactive TUI
  gman tools worktree add backend feature-auth  # Create worktree`,
	Aliases: []string{"t"},
}

// toolsFindCmd searches across repositories (alias for existing find command)
var toolsFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Search across repositories using fzf",
	Long: `Search for files and commits across all repositories using fuzzy finding.

Provides interactive search with preview capabilities for:
- File names and content
- Commit messages and changes
- Branch names

Examples:
  gman tools find                      # Interactive file search
  gman tools find --commits            # Search commit messages
  gman tools find --content "TODO"     # Search file contents`,
	RunE:    findCmd.RunE, // Use existing find command logic
	Aliases: []string{"search", "f"},
}

// toolsIndexCmd manages search index (alias for existing index command)
var toolsIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage search index for fast file and commit searching",
	Long: `Manage the search index used for fast file and commit searching.

The index enables quick searching across large numbers of repositories
and files without scanning the filesystem each time.

Examples:
  gman tools index build              # Build search index
  gman tools index rebuild            # Rebuild search index
  gman tools index status             # Show index status`,
	RunE:    indexCmd.RunE, // Use existing index command logic
	Aliases: []string{"idx"},
}

// toolsDashboardCmd launches TUI dashboard (alias for existing dashboard command)
var toolsDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive TUI dashboard",
	Long: `Launch the interactive Terminal User Interface (TUI) dashboard.

The dashboard provides a visual, interactive way to:
- Browse repositories and their status
- Search files and commits
- Manage repositories and groups
- Perform Git operations

Examples:
  gman tools dashboard                 # Launch dashboard
  gman tools dashboard --force         # Force launch even with compatibility issues`,
	RunE:    runDashboard, // Reuse existing dashboard command logic
	Aliases: []string{"dash", "ui", "tui"},
}

// toolsWorktreeCmd manages worktrees (alias for existing worktree command)
var toolsWorktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage Git worktrees for repositories",
	Long: `Manage Git worktrees to work on multiple branches simultaneously.

Worktrees allow you to have multiple working directories for the same
repository, each checked out to different branches.

Examples:
  gman tools worktree add backend feature-auth    # Create worktree for feature branch
  gman tools worktree list backend                # List worktrees for repository
  gman tools worktree remove backend-feature      # Remove worktree`,
	RunE:    worktreeCmd.RunE, // Use existing worktree command logic
	Aliases: []string{"wt"},
}

// toolsSetupCmd provides setup utilities (alias for existing setup command)
var toolsSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup and configuration utilities",
	Long: `Setup and configuration utilities for gman.

Includes the interactive setup wizard and repository discovery tools
to help new users get started quickly.

Examples:
  gman tools setup                     # Run interactive setup wizard
  gman tools setup discover ~/Projects # Discover repositories in directory`,
	RunE:    runSetup, // Reuse existing setup command logic
	Aliases: []string{"config"},
}

// toolsCompletionCmd generates completion scripts (alias for existing completion command)
var toolsCompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate completion script",
	Long: `Generate shell completion scripts for gman.

Supports bash, zsh, fish, and PowerShell completion.

Examples:
  gman tools completion bash           # Generate bash completion
  gman tools completion zsh            # Generate zsh completion
  gman tools completion fish           # Generate fish completion`,
	RunE:    completionCmd.RunE, // Reuse existing completion command logic
	Aliases: []string{"comp"},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
	
	// Add subcommands to tools group
	toolsCmd.AddCommand(toolsFindCmd)
	toolsCmd.AddCommand(toolsIndexCmd)
	toolsCmd.AddCommand(toolsDashboardCmd)
	toolsCmd.AddCommand(toolsWorktreeCmd)
	toolsCmd.AddCommand(toolsSetupCmd)
	toolsCmd.AddCommand(toolsCompletionCmd)
	
	// Add onboarding as subcommand (for advanced users who want direct access)
	toolsCmd.AddCommand(onboardingCmd)
	
	// Copy flags from original commands
	copyCommandFlags(toolsFindCmd, findCmd)
	copyCommandFlags(toolsIndexCmd, indexCmd)
	copyCommandFlags(toolsDashboardCmd, dashboardCmd)
	copyCommandFlags(toolsWorktreeCmd, worktreeCmd)
	copyCommandFlags(toolsSetupCmd, setupCmd)
	copyCommandFlags(toolsCompletionCmd, completionCmd)
}