package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gman/internal/di"
	"gman/internal/interactive"
	"gman/pkg/types"

	"github.com/spf13/cobra"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [alias]",
	Short: "Switch to a repository directory",
	Long: `Switch to the directory of the specified repository.
If no alias is provided, an interactive menu will be displayed.
This command outputs a special format that the shell wrapper function
can use to change the current working directory.

Examples:
  gman switch my-repo       # Switch to 'my-repo'
  gman switch proj          # Fuzzy match repositories containing 'proj'
  gman switch               # Interactive selection menu`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runSwitch,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Load config and return repository aliases for completion
		configMgr := di.ConfigManager()
		if err := configMgr.Load(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cfg := configMgr.GetConfig()
		var aliases []string
		for alias := range cfg.Repositories {
			aliases = append(aliases, alias)
		}
		return aliases, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Check if shell integration is properly configured
	if !isShellIntegrationActive() {
		fmt.Println("⚠️  Shell integration not detected!")
		fmt.Println("")
		fmt.Println("The 'gman switch' command requires shell integration to change directories.")
		fmt.Println("To fix this, add the following to your ~/.zshrc or ~/.bashrc:")
		fmt.Println("")
		fmt.Println("  # gman shell integration")
		fmt.Println("  gman() {")
		fmt.Println("    local output")
		fmt.Println("    output=$(command gman \"$@\" 2>&1)")
		fmt.Println("    if [[ \"$output\" == GMAN_CD:* ]]; then")
		fmt.Println("      local target_dir=\"${output#GMAN_CD:}\"")
		fmt.Println("      if [ -d \"$target_dir\" ]; then")
		fmt.Println("        cd \"$target_dir\"")
		fmt.Println("        echo \"Switched to: $target_dir\"")
		fmt.Println("      else")
		fmt.Println("        echo \"Error: Directory not found: $target_dir\" >&2")
		fmt.Println("        return 1")
		fmt.Println("      fi")
		fmt.Println("    else")
		fmt.Println("      echo \"$output\"")
		fmt.Println("    fi")
		fmt.Println("  }")
		fmt.Println("")
		fmt.Println("Then reload your shell: source ~/.zshrc")
		fmt.Println("")
		fmt.Println("For more help, see: gman tools setup")
		return fmt.Errorf("shell integration required")
	}
	
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured. Use 'gman add' to add repositories")
	}

	// Collect all available switch targets (repositories + worktrees)
	targets, err := collectSwitchTargets(cfg.Repositories)
	if err != nil {
		return fmt.Errorf("failed to collect switch targets: %w", err)
	}

	if len(targets) == 0 {
		return fmt.Errorf("no repositories or worktrees available")
	}

	var selectedTarget *types.SwitchTarget

	if len(args) == 0 {
		// Interactive mode
		selector := interactive.NewSwitchTargetSelector(targets)
		selectedTarget, err = selector.SelectTarget()
		if err != nil {
			return err
		}
	} else {
		// Direct alias or fuzzy match
		inputAlias := args[0]
		selectedTarget, err = findSwitchTarget(inputAlias, targets)
		if err != nil {
			return err
		}
	}

	// Track recent usage (use repository alias for main repos, or parent repo for worktrees)
	trackingAlias := selectedTarget.Alias
	if selectedTarget.Type == "worktree" && selectedTarget.RepoAlias != "" {
		trackingAlias = selectedTarget.RepoAlias
	}

	if err := configMgr.TrackRecentUsage(trackingAlias); err != nil {
		// Don't fail the switch if tracking fails, just log it silently
		// Could add debug logging here in the future
	}

	// Output special format for shell wrapper to handle
	fmt.Printf("GMAN_CD:%s", selectedTarget.Path)
	return nil
}

// fuzzyMatchRepository performs fuzzy matching on repository aliases
func fuzzyMatchRepository(input string, repos map[string]string) (string, error) {
	input = strings.ToLower(input)
	var matches []string

	for alias := range repos {
		if strings.Contains(strings.ToLower(alias), input) {
			matches = append(matches, alias)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no repositories found matching '%s'", input)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	return "", fmt.Errorf("multiple repositories match '%s': %s. Please be more specific",
		input, strings.Join(matches, ", "))
}

// collectSwitchTargets gathers all available switch targets (repositories + worktrees)
func collectSwitchTargets(repositories map[string]string) ([]types.SwitchTarget, error) {
	var targets []types.SwitchTarget
	gitMgr := di.GitManager()

	// Add all repositories as targets
	for alias, path := range repositories {
		targets = append(targets, types.SwitchTarget{
			Alias:     alias,
			Path:      path,
			Type:      "repository",
			RepoAlias: alias,
		})

		// Add worktrees for this repository
		worktrees, err := gitMgr.ListWorktrees(path)
		if err != nil {
			// Don't fail if we can't list worktrees for a repo, just skip them
			continue
		}

		for _, wt := range worktrees {
			// Skip the main worktree (it's already included as the repository)
			if wt.Path == path {
				continue
			}

			// Create a unique alias for the worktree
			wtAlias := filepath.Base(wt.Path)
			if wtAlias == "." || wtAlias == "" {
				wtAlias = fmt.Sprintf("%s-worktree", alias)
			}

			// Ensure worktree alias is unique
			originalAlias := wtAlias
			counter := 1
			for isAliasUsed(wtAlias, targets) {
				wtAlias = fmt.Sprintf("%s-%d", originalAlias, counter)
				counter++
			}

			targets = append(targets, types.SwitchTarget{
				Alias:       wtAlias,
				Path:        wt.Path,
				Type:        "worktree",
				RepoAlias:   alias,
				Branch:      wt.Branch,
				Description: fmt.Sprintf("Worktree of %s", alias),
			})
		}
	}

	return targets, nil
}

// findSwitchTarget finds a switch target by alias or fuzzy matching
func findSwitchTarget(input string, targets []types.SwitchTarget) (*types.SwitchTarget, error) {
	input = strings.ToLower(input)

	// Try exact match first
	for _, target := range targets {
		if strings.ToLower(target.Alias) == input {
			return &target, nil
		}
	}

	// Try fuzzy matching
	var matches []types.SwitchTarget
	for _, target := range targets {
		if strings.Contains(strings.ToLower(target.Alias), input) {
			matches = append(matches, target)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no targets found matching '%s'", input)
	}

	if len(matches) == 1 {
		return &matches[0], nil
	}

	// Multiple matches
	var aliases []string
	for _, match := range matches {
		aliases = append(aliases, match.Alias)
	}
	return nil, fmt.Errorf("multiple targets match '%s': %s. Please be more specific",
		input, strings.Join(aliases, ", "))
}

// isAliasUsed checks if an alias is already used in the targets list
func isAliasUsed(alias string, targets []types.SwitchTarget) bool {
	for _, target := range targets {
		if target.Alias == alias {
			return true
		}
	}
	return false
}

// isShellIntegrationActive detects if gman is running within the shell wrapper function
func isShellIntegrationActive() bool {
	// Method 1: Check for GMAN_SHELL_INTEGRATION environment variable
	// The shell wrapper should set this to indicate it's active
	if os.Getenv("GMAN_SHELL_INTEGRATION") == "1" {
		return true
	}
	
	// Method 2: Check if we're being called by the gman shell function
	// Look for indicators that suggest we're in a shell wrapper
	// Note: This is a heuristic and not foolproof - we primarily rely on method 1
	if parent := os.Getenv("_"); parent != "" && strings.Contains(parent, "gman") {
		// Only trust this if we also have other shell integration indicators
		// This prevents false positives when gman is called directly
		if os.Getenv("GMAN_WRAPPER_ACTIVE") == "1" {
			return true
		}
	}
	
	// Method 3: Check the process hierarchy (simplified check)
	// If SHLVL exists and is reasonable, we might be in a shell environment
	if shlvl := os.Getenv("SHLVL"); shlvl != "" {
		// Shell level exists, check if we're in an interactive environment
		if os.Getenv("PS1") != "" || os.Getenv("ZSH_NAME") != "" || os.Getenv("BASH") != "" {
			// We're in an interactive shell - the user should set up the wrapper
			// For now, assume it's NOT set up since this is the main issue
			return false
		}
	}
	
	// Method 4: Check for explicit bypass flag for advanced users
	if os.Getenv("GMAN_SKIP_SHELL_CHECK") == "1" {
		return true
	}
	
	// Default: assume shell integration is not active
	// This will prompt users to set up the wrapper function
	return false
}
