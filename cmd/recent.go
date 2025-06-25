package cmd

import (
	"fmt"
	"strings"
	"time"

	"gman/internal/di"
	"gman/pkg/types"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// recentCmd represents the recent command
var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show recently accessed repositories",
	Long: `Display a list of recently accessed repositories in order of last access.
This helps you quickly switch to repositories you've been working with recently.

Examples:
  gman recent           # Show recent repositories
  gman recent --limit 5 # Show only the 5 most recent`,
	Aliases: []string{},
	RunE:    runRecent,
}

var recentLimit int

func init() {
	// Command is now available via: gman repo recent
	// Removed direct rootCmd registration to avoid duplication
	recentCmd.Flags().IntVar(&recentLimit, "limit", 10, "Maximum number of recent repositories to show")
}

func runRecent(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := di.ConfigManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := configMgr.GetConfig()
	recentEntries := configMgr.GetRecentUsage()

	if len(recentEntries) == 0 {
		fmt.Println("No recent repository usage found.")
		fmt.Println("Use 'gman switch <alias>' to start tracking repository usage.")
		return nil
	}

	// Apply limit
	if recentLimit > 0 && len(recentEntries) > recentLimit {
		recentEntries = recentEntries[:recentLimit]
	}

	// Display recent repositories
	displayRecentRepositories(recentEntries, cfg.Repositories)
	return nil
}

func displayRecentRepositories(entries []types.RecentEntry, repos map[string]string) {
	fmt.Printf("\n%s\n", color.CyanString("Recently accessed repositories:"))
	fmt.Println(strings.Repeat("─", 60))

	maxAliasLen := 10
	for _, entry := range entries {
		if len(entry.Alias) > maxAliasLen {
			maxAliasLen = len(entry.Alias)
		}
	}

	for i, entry := range entries {
		path, exists := repos[entry.Alias]
		if !exists {
			// Repository no longer exists in config, skip it
			continue
		}

		// Format relative time
		timeAgo := formatTimeAgo(entry.AccessTime)

		// Truncate path if too long
		displayPath := path
		maxPathLen := 40
		if len(displayPath) > maxPathLen {
			displayPath = "..." + displayPath[len(displayPath)-(maxPathLen-3):]
		}

		fmt.Printf("%s %-*s %s %s\n",
			color.YellowString("[%d]", i+1),
			maxAliasLen, color.GreenString(entry.Alias),
			color.WhiteString("→ %-*s", maxPathLen, displayPath),
			color.BlueString("(%s)", timeAgo))
	}

	fmt.Printf("\n%s\n",
		color.MagentaString("Tip: Use 'gman switch <number>' or 'gman switch <alias>' to switch"))
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2")
	}
}
