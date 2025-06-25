package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gman/internal/config"
	"gman/internal/tui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive TUI dashboard",
	Long: `Launch the interactive Terminal User Interface (TUI) dashboard for gman.

The dashboard provides a unified interface for managing multiple Git repositories
with four main panels:

1. Repository Panel - Browse and select repositories
2. Status Panel - View detailed repository status
3. Search Panel - Search files and commits across repositories
4. Preview Panel - Preview files and commit content

Key Features:
• Real-time repository status monitoring
• Integrated search with fzf (Phase 5.1)
• Keyboard-driven navigation
• Repository grouping and filtering
• File and commit previews

Navigation:
• Tab/Shift+Tab: Navigate between panels
• 1-4: Jump directly to specific panels
• Arrow keys: Navigate within panels
• ?: Toggle help
• q: Quit

The dashboard integrates with all existing gman functionality while providing
a modern, intuitive interface for repository management.`,
	Aliases: []string{"dash", "tui", "ui"},
	RunE:    runDashboard,
}

var dashboardFlags struct {
	theme string
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	// Add flags
	dashboardCmd.Flags().StringVar(&dashboardFlags.theme, "theme", "dark", "Color theme (dark, light)")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Initialize configuration manager
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if repositories are configured
	repos := configMgr.GetConfig().Repositories
	if repos == nil {
		repos = make(map[string]string)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("Add repositories with: gman add <alias> <path>")
		fmt.Println("Or run: gman list to see available commands")
		return nil
	}

	// Validate theme
	if dashboardFlags.theme != "dark" && dashboardFlags.theme != "light" {
		return fmt.Errorf("invalid theme: %s (available: dark, light)", dashboardFlags.theme)
	}

	// Set up terminal for TUI
	if err := setupTerminal(); err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}
	defer restoreTerminal()

	// Run the TUI dashboard
	if err := tui.Run(configMgr); err != nil {
		return fmt.Errorf("dashboard error: %w", err)
	}

	return nil
}

// setupTerminal prepares the terminal for TUI mode
func setupTerminal() error {
	// Check if we're in a suitable terminal
	if !isTerminalSuitable() {
		return fmt.Errorf("terminal does not support TUI mode")
	}

	// Additional terminal setup could go here
	// For now, Bubble Tea handles most of the setup
	
	return nil
}

// restoreTerminal restores the terminal after TUI mode
func restoreTerminal() {
	// Bubble Tea handles most of the cleanup
	// Additional cleanup could go here if needed
}

// isTerminalSuitable checks if the terminal supports TUI mode
func isTerminalSuitable() bool {
	// Check if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Check terminal environment variables
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check for minimum terminal size
	// This could be enhanced with actual terminal size detection
	
	return true
}