package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gman/internal/di"
	"gman/internal/interactive"
	"gman/pkg/types"

	"github.com/spf13/cobra"
)

var (
	showRecentOnly bool
	recentLimit    int
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [alias]",
	Short: "Switch to a repository directory with recent history",
	Long: `Switch to the directory of the specified repository.
If no alias is provided, an interactive menu will be displayed showing
all repositories with recently accessed ones highlighted at the top.

The interactive menu prioritizes recently accessed repositories to 
improve navigation efficiency in your daily workflow.

Examples:
  gman switch my-repo       # Switch to 'my-repo'
  gman switch proj          # Fuzzy match repositories containing 'proj'
  gman switch               # Interactive selection menu with recent repos first
  gman switch --recent      # Show only recently accessed repositories
  gman switch --recent --limit 5    # Show only last 5 accessed repositories`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runSwitch,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Configuration is already loaded by root command's PersistentPreRunE
		configMgr := di.ConfigManager()

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
	
	switchCmd.Flags().BoolVar(&showRecentOnly, "recent", false, "Show only recently accessed repositories")
	switchCmd.Flags().IntVar(&recentLimit, "limit", 10, "Limit number of recent repositories shown (used with --recent)")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Check if shell integration is properly configured
	if !isShellIntegrationActive() {
		fmt.Println("⚠️  Shell integration not detected!")
		fmt.Println("")
		
		// Show diagnostic information
		fmt.Println("🔍 Diagnostic Information:")
		fmt.Println(getShellIntegrationDiagnostics())
		fmt.Println("")
		
		fmt.Println("📋 Quick Fix:")
		fmt.Println("The 'gman switch' command requires shell integration to change directories.")
		fmt.Println("Add the following to your shell configuration file:")
		fmt.Println("")
		fmt.Println("  # gman shell integration")
		fmt.Println("  gman() {")
		fmt.Println("    local output")
		fmt.Println("    export GMAN_SHELL_INTEGRATION=1")
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
		fmt.Println("🔄 Then reload your shell or run: source ~/.zshrc")
		fmt.Println("")
		fmt.Println("🚀 Alternative: Use 'gman tools init shell' for automated setup!")
		fmt.Println("📚 For detailed help: https://docs.anthropic.com/en/docs/claude-code/troubleshooting")
		return fmt.Errorf("shell integration required")
	}
	
	// Configuration is already loaded by root command's PersistentPreRunE
	configMgr := di.ConfigManager()
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

	// Filter targets if --recent flag is used
	if showRecentOnly {
		targets = filterRecentTargets(targets, cfg.RecentUsage, recentLimit)
		if len(targets) == 0 {
			return fmt.Errorf("no recently accessed repositories found")
		}
	} else {
		// Sort targets with recent repositories first
		targets = sortTargetsByRecency(targets, cfg.RecentUsage)
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

			// Create a descriptive alias for the worktree using repo prefix
			wtBaseName := filepath.Base(wt.Path)
			if wtBaseName == "." || wtBaseName == "" {
				wtBaseName = "worktree"
			}
			
			// Always prefix with repository alias to provide clear context
			wtAlias := fmt.Sprintf("%s/%s", alias, wtBaseName)

			// Ensure worktree alias is unique (fallback safety)
			originalAlias := wtAlias
			counter := 1
			for isAliasUsed(wtAlias, targets) {
				wtAlias = fmt.Sprintf("%s-%d", originalAlias, counter)
				counter++
			}

			// Enhanced description with branch, path, and repository context
			description := fmt.Sprintf("Worktree: %s", wtBaseName)
			if wt.Branch != "" {
				description = fmt.Sprintf("Worktree: %s (branch: %s) → %s", wtBaseName, wt.Branch, wt.Path)
			} else {
				description = fmt.Sprintf("Worktree: %s → %s", wtBaseName, wt.Path)
			}

			targets = append(targets, types.SwitchTarget{
				Alias:       wtAlias,
				Path:        wt.Path,
				Type:        "worktree",
				RepoAlias:   alias,
				Branch:      wt.Branch,
				Description: description,
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
	// Primary method: Check for GMAN_SHELL_INTEGRATION environment variable
	// The shell wrapper should set this to indicate it's active
	if os.Getenv("GMAN_SHELL_INTEGRATION") == "1" {
		return true
	}
	
	// Advanced users can bypass shell check entirely
	if os.Getenv("GMAN_SKIP_SHELL_CHECK") == "1" {
		return true
	}
	
	// If neither primary indicator is set, shell integration is not active
	return false
}

// getShellIntegrationDiagnostics provides detailed diagnostic information
func getShellIntegrationDiagnostics() string {
	var diagnostics []string
	
	// Check environment variables
	if os.Getenv("GMAN_SHELL_INTEGRATION") != "1" {
		diagnostics = append(diagnostics, "❌ GMAN_SHELL_INTEGRATION not set to '1'")
	} else {
		diagnostics = append(diagnostics, "✅ GMAN_SHELL_INTEGRATION properly set")
	}
	
	// Check shell type
	if shellName := os.Getenv("SHELL"); shellName != "" {
		diagnostics = append(diagnostics, fmt.Sprintf("📝 Shell: %s", shellName))
		
		// Suggest appropriate config file
		configFile := "~/.bashrc"
		if strings.Contains(shellName, "zsh") {
			configFile = "~/.zshrc"
		} else if strings.Contains(shellName, "fish") {
			configFile = "~/.config/fish/config.fish"
		}
		diagnostics = append(diagnostics, fmt.Sprintf("💡 Expected config file: %s", configFile))
	}
	
	// Check if we're in an interactive shell
	isInteractive := os.Getenv("PS1") != "" || os.Getenv("ZSH_NAME") != "" || os.Getenv("BASH") != ""
	if isInteractive {
		diagnostics = append(diagnostics, "✅ Interactive shell detected")
	} else {
		diagnostics = append(diagnostics, "❓ Non-interactive shell or environment")
	}
	
	// Check for bypass flag
	if os.Getenv("GMAN_SKIP_SHELL_CHECK") == "1" {
		diagnostics = append(diagnostics, "⚠️  Shell check bypass is active")
	}
	
	return strings.Join(diagnostics, "\n")
}

// filterRecentTargets returns only recently accessed repositories up to the specified limit
func filterRecentTargets(targets []types.SwitchTarget, recentRepos []types.RecentEntry, limit int) []types.SwitchTarget {
	if len(recentRepos) == 0 {
		return []types.SwitchTarget{}
	}

	var recentTargets []types.SwitchTarget
	recentMap := make(map[string]types.RecentEntry)
	
	// Create a map for quick lookup
	for _, recent := range recentRepos {
		recentMap[recent.Alias] = recent
	}

	// Find targets that match recent repositories
	for _, target := range targets {
		if recent, exists := recentMap[target.Alias]; exists {
			// Add recent information to the target
			targetCopy := target
			targetCopy.LastAccessed = recent.AccessTime
			recentTargets = append(recentTargets, targetCopy)
		}
	}

	// Sort by recency (most recent first)
	sort.Slice(recentTargets, func(i, j int) bool {
		return recentTargets[i].LastAccessed.After(recentTargets[j].LastAccessed)
	})

	// Apply limit
	if limit > 0 && len(recentTargets) > limit {
		recentTargets = recentTargets[:limit]
	}

	return recentTargets
}

// sortTargetsByRecency sorts targets with recent repositories first, then alphabetically
func sortTargetsByRecency(targets []types.SwitchTarget, recentRepos []types.RecentEntry) []types.SwitchTarget {
	if len(recentRepos) == 0 {
		// No recent repositories, just sort alphabetically
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Alias < targets[j].Alias
		})
		return targets
	}

	// Create a map for quick lookup of recent access times
	recentMap := make(map[string]time.Time)
	for _, recent := range recentRepos {
		recentMap[recent.Alias] = recent.AccessTime
	}

	// Add recent access times to targets
	for i := range targets {
		if lastAccessed, exists := recentMap[targets[i].Alias]; exists {
			targets[i].LastAccessed = lastAccessed
		}
	}

	// Sort: recent repositories first (by recency), then non-recent alphabetically
	sort.Slice(targets, func(i, j int) bool {
		iRecent := !targets[i].LastAccessed.IsZero()
		jRecent := !targets[j].LastAccessed.IsZero()

		if iRecent && jRecent {
			// Both are recent, sort by most recent first
			return targets[i].LastAccessed.After(targets[j].LastAccessed)
		} else if iRecent && !jRecent {
			// i is recent, j is not - i comes first
			return true
		} else if !iRecent && jRecent {
			// j is recent, i is not - j comes first
			return false
		} else {
			// Neither is recent, sort alphabetically
			return targets[i].Alias < targets[j].Alias
		}
	})

	return targets
}
